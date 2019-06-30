package common

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
	"github.com/xixuejia/viper"
	"gopkg.in/yaml.v2"
	"strings"
)

const ROUTINE_UPPER_LIMIT = 2000
const DEFAULT_ROUTINE_LIMIT = 500

var IGNORED_ERRORS = []string{"MVCC_READ_CONFLICT"}

// Max number of active go routines
var routineLimit = DEFAULT_ROUTINE_LIMIT

func GetViperInstance(file string, configType string) (*viper.Viper, error) {
	viperInstance := viper.New()
	viperInstance.SetConfigType(configType)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := viperInstance.ReadConfig(bufio.NewReader(f)); err != nil {
		return nil, err
	}
	return viperInstance, nil
}

// GetConfigBackends gets the config backends from multiple configuration file
func GetConfigBackends(stringPath ...string) (core.ConfigProvider, error) {
	var configBackends []core.ConfigBackend
	for _, backendPath := range stringPath {
		configProvider := config.FromFile(pathvar.Subst(backendPath))
		newBackend, err := configProvider()
		if err != nil {
			return nil, err
		}
		configBackends = append(configBackends, newBackend...)
	}
	return func() ([]core.ConfigBackend, error) {
		return configBackends, nil
	}, nil
}

//	GetTempChannelConfigFile will generate the temporary channel yaml file and return the file path
func GetTempChannelConfigFile(channelID string, peers []string) (string, error) {
	peersChannel := make(map[string]PeerChannel)
	for _, peer := range peers {
		peerChannel := PeerChannel{true, true, true, true}
		peersChannel[peer] = peerChannel
	}
	channel := Channel{
		peersChannel,
	}
	channels := make(map[string]Channel)
	channels[channelID] = channel
	connectionProfile := Channels{
		Channels: channels,
	}
	data, err := yaml.Marshal(&connectionProfile)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile("config/"+channelID+".yaml", data, 0644)
	if err != nil {
		return "", err
	}
	return "config/" + channelID + ".yaml", nil
}

// return the number of successful execution and error(if any)
// sequential: false when iterFunc can be executed concurrently; true when iterFunc should be sequentially executed
// chaincode instantiate should be executed sequentially
func IterateFunc(base *Base, iterFunc func(iterationIndex int) error, sequential bool) (successCount uint64, failCount uint64, retErr error) {
	successCount = 0
	failCount = 0
	if sequential {
		routineLimit = 1
	}
	errChan := make(chan error, routineLimit)
	throttle := make(chan struct{}, routineLimit)
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		for len(errChan) > 0 {
			if err := <-errChan; err != nil {
				if !(strings.ToUpper(viper.GetString(IGNORE_ERRORS)) == "TRUE") {
					retErr = err
					break
				}
			}
		}
		close(throttle)
		close(errChan)
	}()
	goFunc := func(iterationIndex int) {
		defer func() {
			wg.Done()
			<-throttle
		}()
		err := iterFunc(iterationIndex)
		if err != nil {
			atomic.AddUint64(&failCount, 1)
			// check whether we should ignore the list
			for _, e := range IGNORED_ERRORS {
				if strings.Contains(err.Error(), e) {
					Logger.Warn(fmt.Sprintf("Ignoring error: %s", err))
					return
				}
			}
			Logger.Error(fmt.Sprintf("Error in iteration function %d: %s\n", iterationIndex, err))
			select {
			case errChan <- err:
			default:
				{
					// in this case errChan is already full
					return
				}
			}
		} else {
			atomic.AddUint64(&successCount, 1)
		}
	}
	// try to parse IterationCount as integer
	iterationCount, err := strconv.Atoi(base.IterationCount)
	if err == nil {
		Logger.Info(fmt.Sprintf("iterationCount: %d", iterationCount-base.currentIter))
	countLoop:
		for base.CurrentIter() < iterationCount {
			throttle <- struct{}{}
			wg.Add(1)
			go goFunc(base.CurrentIter())
			if base.currentIter != iterationCount-1 {
				base.Wait()
			}
			select {
			case err := <-errChan:
				if err != nil {
					if !(strings.ToUpper(viper.GetString(IGNORE_ERRORS)) == "TRUE") {
						retErr = err
						break countLoop
					}
					Logger.Error(fmt.Sprintf("Error detected: %s", err))
					base.Next()
				}
			default:
				base.Next()
			}

		}
	} else if duration, err := time.ParseDuration(base.IterationCount); err == nil {
		Logger.Info(fmt.Sprintf("iterationDuration is: %s\n", duration))
		timer := time.NewTimer(duration)
	durationLoop:
		for {
			select {
			case <-timer.C:
				break durationLoop
			default:
				throttle <- struct{}{}
				wg.Add(1)
				go goFunc(base.CurrentIter())
				base.Wait()
				select {
				case err := <-errChan:
					if err != nil {
						if !(strings.ToUpper(viper.GetString(IGNORE_ERRORS)) == "TRUE") {
							retErr = err
							break durationLoop
						}
						Logger.Error(fmt.Sprintf("Error detected: %s", err))
						base.Next()
					}
				default:
					base.Next()
				}
			}
		}
	} else {
		retErr = errors.New("Error parsing iterationCount: " + base.IterationCount)
	}
	return
}

func SetRoutineLimit(num int) {
	switch {
	case num < 0:
		Logger.Info(fmt.Sprintf("Ignoring goroutine number limit setting: %d", num))
	case num > ROUTINE_UPPER_LIMIT:
		Logger.Info(fmt.Sprintf("Ignoring goroutine number(%d) that exceeds uppper limit %d", num, ROUTINE_UPPER_LIMIT))
	default:
		routineLimit = num
		Logger.Info(fmt.Sprintf("Setting goroutine limit to %d", num))
	}
}
