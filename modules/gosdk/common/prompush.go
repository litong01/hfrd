package common

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"sync/atomic"

	metrics "github.com/rcrowley/go-metrics"
)

// Config provides a container with configuration parameters for
// the Graphite exporter
type Config struct {
	Addr            *net.TCPAddr     // Network address to connect to
	Registry        metrics.Registry // Registry to be exported
	FlushInterval   time.Duration    // Flush interval
	DurationUnit    time.Duration    // Time conversion unit for durations
	InstanceId      string           // Instance identifier which normally maps to hostname
	TestId          string           // Test identifier
	Percentiles     []float64        // Percentiles to export from timers and histograms
	metricsStopChan chan struct{}    // the channel to notify main goroutine to exit
}

const (
	MaxIdleConnections int = 20
	RequestTimeout     int = 5
	SCRAPE_INTERVAL        = 10 * time.Second
)

var httpClient *http.Client
var metricTarget string
var stopFlag int32 = 0

func createHTTPClient(c Config) {
	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
		},
		Timeout: time.Duration(RequestTimeout) * time.Second,
	}
	metricTarget = strings.Join([]string{"http://", c.Addr.String(),
		"/metrics/job/hfrd/instance/", c.InstanceId, "/testid/", c.TestId}, "")
	fmt.Printf("The full metric push target url is %s\n", metricTarget)
}

// PromPush is a blocking exporter function which reports metrics in r
// to a prometheus push gateway server located at addr, flushing them
// every d duration.
func PromPush(r metrics.Registry, d time.Duration, instance_id string, test_id string, addr *net.TCPAddr, metricsStopChan chan (struct{})) {
	WithConfig(Config{
		Addr:            addr,
		Registry:        r,
		FlushInterval:   d,
		DurationUnit:    time.Millisecond,
		InstanceId:      instance_id,
		TestId:          test_id,
		Percentiles:     []float64{0.5, 0.75, 0.95, 0.99, 0.999},
		metricsStopChan: metricsStopChan,
	})
}

// WithConfig is a blocking exporter function just like Graphite,
// but it takes a GraphiteConfig instead.
func WithConfig(c Config) {
	createHTTPClient(c)
	for _ = range time.Tick(c.FlushInterval) {
		if err := prometheus(&c); nil != err {
			log.Println(err)
		}
		if atomic.LoadInt32(&stopFlag) == 1 {
			// We need to remove these metrics
			// so that the gateway won't kept these values around, this is to
			// avoid prometheus scraper keep pulling values
			var buffer bytes.Buffer
			buffer.WriteString("")
			// wait for prom to scrape metrics before deleting metrics in gateway
			time.Sleep(SCRAPE_INTERVAL + PUSH_INTERVAL)
			pushmetrics(&buffer, http.MethodDelete)
			close(c.metricsStopChan) // we can notify the main routine to exit after the last metrics push
			break
		}
	}
}

func StopPush() {
	atomic.StoreInt32(&stopFlag, 1)
	fmt.Println("\nStoping pushing metrics. Metrics will be all removed from gateway.")
}

func pushmetrics(content *bytes.Buffer, method string) {
	if metricTarget == "" {
		Logger.Info("Ignoring empty metric target...")
		return
	}
	req, err := http.NewRequest(method, metricTarget, content)
	if err != nil {
		Logger.Error(fmt.Sprintf("Error Occured. %+v", err))
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	response, err := httpClient.Do(req)
	if err != nil && response == nil {
		Logger.Error(fmt.Sprintf("Publish metrics has failed with error. %+v\n", err))
		return
	}
	// Close the connection to reuse it
	defer response.Body.Close()

	// Let's check if the work actually is done
	// We have seen inconsistencies even when we get 200 OK response
	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		Logger.Error(fmt.Sprintf("Couldn't parse response body from pushgateway. %+v", err))
	}
}

func prometheus(c *Config) error {
	flushSeconds := float64(c.FlushInterval) / float64(time.Second)
	var buffer bytes.Buffer
	c.Registry.Each(func(name string, i interface{}) {
		mname := strings.Replace(name, ".", "_", -1)
		switch metric := i.(type) {
		case metrics.Counter:
			count := metric.Snapshot().Count()
			metric.Clear() // clear the counter after each flushInterval so we can get the correct count/ps
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s counter\nhfrd_%s %d\n",
				mname, mname, count))
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s_ps gauge\nhfrd_%s_ps %0.2f\n",
				mname, mname, float64(count)/flushSeconds))
		case metrics.Gauge:
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s gauge\nhfrd_%s %d\n",
				mname, mname, metric.Value()))
		case metrics.GaugeFloat64:
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s gauge\nhfrd_%s %f\n",
				mname, mname, metric.Value()))
		case metrics.Meter:
			m := metric.Snapshot()
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s_tx_count counter\nhfrd_%s_count %d\n",
				mname, mname, m.Count()))
		case metrics.Timer:
			t := metric.Snapshot()
			count := t.Count()
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s_count counter\nhfrd_%s_count %d\n",
				mname, mname, count))
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s_min gauge\nhfrd_%s_min %d\n",
				mname, mname, time.Duration(t.Min())/1000000))
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s_max gauge\nhfrd_%s_max %d\n",
				mname, mname, time.Duration(t.Max())/1000000))
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s_mean gauge\nhfrd_%s_mean %d\n",
				mname, mname, time.Duration(t.Mean())/1000000))
			buffer.WriteString(fmt.Sprintf("# TYPE hfrd_%s_tx_rate1s gauge\nhfrd_%s_tx_rate1s %f\n",
				mname, mname, t.RateMean()))
		default:
			fmt.Printf("unable to record metric of type %T\n", i)
		}
	})
	if buffer.Len() > 0 {
		pushmetrics(&buffer, http.MethodPost)
	}
	return nil
}
