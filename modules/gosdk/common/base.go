package common

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
)

const FIXED = 0
const RANDOM = 1
const PUSH_INTERVAL = 10 * time.Second // push metrics to graphite interval: seconds

type Interval struct {
	mark         float64
	intervalType int
}

func NewInterval(p string) *Interval {
	fixed := regexp.MustCompile(`[0-9]*\.?[0-9]*s`)
	random := regexp.MustCompile(`[0-9]*\.?[0-9]*r`)
	if fixed.MatchString(p) {
		v, err := strconv.ParseFloat(strings.Replace(p, "s", "", 1), 64)
		if err != nil {
			panic("Error parsing float " + p)
		}
		return &Interval{v, FIXED}
	} else if random.MatchString(p) {
		v, err := strconv.ParseFloat(strings.Replace(p, "r", "", 1), 64)
		if err != nil {
			panic("Error parsing float " + p)
		}
		return &Interval{v, RANDOM}
	} else {
		panic("Not supported intervals " + p)
	}
}

type Base struct {
	InstanceID        int
	IterationCount    string
	RetryCount        int
	CertRootDir       string
	ConnectionProfile string

	Hostname string

	count        int
	gauge        float64
	interval     *Interval
	currentIter  int
	currentRetry int

	metricsStopChan chan struct{}
}

func (p *Base) IncCount()                   { p.count++ }
func (p *Base) SetGauge(g float64)          { p.gauge = g }
func (p *Base) AddToGauge(g float64)        { p.gauge += g }
func (p *Base) DoMetrics(metricName string) {}
func (p *Base) SetIterationInterval(pattern string) {
	p.interval = NewInterval(pattern)
}
func (p *Base) Wait() {
	if p.interval.mark == 0 {
	} else if p.interval.intervalType == FIXED {
		time.Sleep(time.Duration(p.interval.mark*1e9) * time.Nanosecond)
	} else if p.interval.intervalType == RANDOM {
		time.Sleep(time.Duration(rand.Float64()*p.interval.mark*1e9) * time.Nanosecond)
	}
}
func (p *Base) CurrentIter() int               { return p.currentIter }
func (p *Base) SetCurrentIter(currentIter int) { p.currentIter = currentIter }
func (p *Base) ResetCurrentIter() int          { p.currentIter = 0; return p.currentIter }
func (p *Base) CurrentRetry() int              { return p.currentRetry }
func (p *Base) ResetCurrentRetry() int         { p.currentRetry = 0; return p.currentRetry }
func (p *Base) Next() int                      { p.currentIter++; return p.currentIter }
func (p *Base) NextRetry() int                 { p.currentRetry++; return p.currentRetry }
func TrackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	timer := metrics.GetOrRegisterTimer(name+".timer", metrics.DefaultRegistry)
	timer.Update(elapsed) // unit: ms
	Logger.Debug(fmt.Sprintf("Metric %s updated, elapsed time %s", name, elapsed))
}

func InitializeMetrics(name string) {
	metrics.GetOrRegisterCounter(name+".counter", metrics.DefaultRegistry)
	metrics.GetOrRegisterTimer(name+".timer", metrics.DefaultRegistry)
}

func TrackCount(name string, increment int64) {
	counter := metrics.GetOrRegisterCounter(name+".counter", metrics.DefaultRegistry)
	counter.Inc(increment)
}

func TrackGauge(name string, value int64) {
	gauge := metrics.GetOrRegisterGauge(name+".gauge", metrics.DefaultRegistry)
	gauge.Update(value)
}

func Delay(delayTime string) {
	newTime := strings.TrimSpace(delayTime)
	if len(newTime) > 0 {
		delay := NewInterval(newTime)
		time.Sleep(time.Duration(delay.mark*1e9) * time.Nanosecond)
	}
}

// Print metrics summary into console and do clean up
func (p *Base) PrintMetrics(name string) {
	timer := metrics.GetOrRegisterTimer(name+".timer", metrics.DefaultRegistry)
	snap := timer.Snapshot()
	metricsStr := "*******************************************************\n"
	metricsStr = metricsStr + fmt.Sprintf("Request succeeded count: %d\n", snap.Count())
	if strings.Contains(name, "invoke") {
		counter := metrics.GetOrRegisterCounter(name+".fail.counter", metrics.DefaultRegistry)
		snapCounter := counter.Snapshot()
		metricsStr = metricsStr + fmt.Sprintf("Request failed count: %d\n", snapCounter.Count())
	}
	metricsStr = metricsStr + fmt.Sprintf("Average rate: %f req/s\n", snap.RateMean())
	metricsStr = metricsStr + fmt.Sprintf("Standard deviation: %f\n", snap.StdDev())
	metricsStr = metricsStr + fmt.Sprintf("Max response time: %s\n", time.Duration(snap.Max()))
	metricsStr = metricsStr + fmt.Sprintf("Min response time: %s\n", time.Duration(snap.Min()))
	metricsStr = metricsStr + fmt.Sprintf("Average response time: %s\n", time.Duration(snap.Mean()))
	metricsStr = metricsStr + fmt.Sprintf("95 percents of requests completed in %s\n", time.Duration(snap.Percentile(0.95)))
	metricsStr = metricsStr + "*******************************************************\n"
	fmt.Print(metricsStr)
	if viper.GetString("METRICS_TARGET_URL") != "" || viper.GetString("PROMETHEUS_TARGET_URL") != "" {
		StopPush()
		select {
		case <-p.metricsStopChan:
			fmt.Println("Push metrics routine exited, exiting the main routine")
		case <-time.After(PUSH_INTERVAL + SCRAPE_INTERVAL + PUSH_INTERVAL):
			fmt.Println("Timedout to stop the metrics routine, exiting the main routine")
		}
	}
}

func NewBase() *Base {
	np := Base{}
	np.InstanceID = 0
	np.IterationCount = "1"
	np.RetryCount = 3
	np.CertRootDir = ""
	np.ConnectionProfile = ""
	np.Hostname, _ = os.Hostname()

	np.count = 0
	np.gauge = 0
	np.currentIter = 0
	np.currentRetry = 0
	np.metricsStopChan = make(chan struct{})
	np.SetIterationInterval("0s")
	viper.AutomaticEnv()
	metricsTargetUrl := viper.GetString("METRICS_TARGET_URL")
	prometheusTargetUrl := viper.GetString("PROMETHEUS_TARGET_URL")
	if prometheusTargetUrl != "" {
		addr1, err1 := net.ResolveTCPAddr("tcp", prometheusTargetUrl)
		if err1 == nil {
			fmt.Printf("metrics will be pushed to %s\n", prometheusTargetUrl)
			test_id := viper.GetString("TEST_ID")
			go PromPush(metrics.DefaultRegistry, PUSH_INTERVAL,
				np.Hostname, test_id, addr1, np.metricsStopChan)
		} else {
			viper.Set("PROMETHEUS_TARGET_URL", "")
			fmt.Printf("error resolving TCP Addr: %s\n", err1)
		}
	} else {
		fmt.Println("PROMETHEUS_TARGET_URL not defined, will not push metrics out", prometheusTargetUrl)
	}
	if metricsTargetUrl != "" {
		addr, err := net.ResolveTCPAddr("tcp", metricsTargetUrl)
		if err == nil {
			fmt.Printf("metrics will be pushed to %s\n", metricsTargetUrl)
			test_id := viper.GetString("TEST_ID")
			go PromPush(metrics.DefaultRegistry, PUSH_INTERVAL,
				np.Hostname, test_id, addr, np.metricsStopChan)
		} else {
			viper.Set("METRICS_TARGET_URL", "")
			fmt.Printf("error resolving TCP Addr: %s\n", err)
		}
	} else {
		fmt.Println("METRICS_TARGET_URL not defined, will not push metrics out")
	}
	return &np
}
