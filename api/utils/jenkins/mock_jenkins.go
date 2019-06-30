package jenkins

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
)

func NewMockJenkins() Jenkins {
	network := &job{name: NETWORK, queue: map[int]int{}, currentQueueId: 1, currentJobId: 1}
	networkCM := &job{name: NETWORK_CM, queue: map[int]int{}, currentQueueId: 1, currentJobId: 1}
	connection := &job{name: CONNECTION, queue: map[int]int{}, currentQueueId: 1, currentJobId: 1}
	connectionCM := &job{name: CONNECTION_CM, queue: map[int]int{}, currentQueueId: 1, currentJobId: 1}
	test := &job{name: TEST, queue: map[int]int{}, currentQueueId: 1, currentJobId: 1}
	return &mock{jobs: []*job{network, networkCM, connection, connectionCM, test}}
}

// Mock jenkins
type mock struct {
	jobs []*job
}

type job struct {
	name           string
	queue          map[int]int // queueid to jobid mapping
	currentQueueId int
	currentJobId   int
	sync.RWMutex
}

// Input jobname and jobid string
// Returns console text as []byte and error
func (mock *mock) GetConsoleText(jobname, jobid string) ([]byte, error) {
	if mock.exist(jobname, jobid) {
		return []byte("Mocked console text"), nil
	}
	return []byte{}, errors.New("W")
}

// Input jobname and jobid string
// Returns the content of workdir/results/service.json as []byte and error
func (mock *mock) GetServiceJson(jobname, jobid string) ([]byte, error) {
	if !mock.exist(jobname, jobid) {
		return []byte{}, errors.New("No service.json")
	}
	switch jobname {
	case NETWORK, NETWORK_CM, CONNECTION, CONNECTION_CM:
		service, _ := json.Marshal(map[string]string{"serviceid": "mockedid", "servicename": "mockedservicekey"})
		return service, nil
	default:
		return []byte{}, errors.New("No service.json")
	}
}

// Input jobname and jobid string
// Returns the content of workdir/results/package.tar as []byte and error
func (mock *mock) ServeTar(jobname string, jobid string) ([]byte, error) {
	if mock.exist(jobname, jobid) {
		return []byte("Mocked package tar"), nil
	}
	return []byte{}, errors.New("No package tar")
}

// Input jobname string and parameters map. The params map will be passed
// to the job--Build with Parameters
// Returns queue id as string and error
func (mock *mock) TriggerJob(jobname string, params map[string]string) (string, error) {
	for _, job := range mock.jobs {
		if jobname == job.name {
			job.Lock()
			job.queue[job.currentJobId] = job.currentJobId
			job.currentJobId += 1
			job.currentQueueId += 1
			job.Unlock()
			return strconv.Itoa(job.currentQueueId - 1), nil
		}
	}
	return "", fmt.Errorf("No job with name %s", jobname)
}

// Input queueid and jobname string
// Return jobid, jobstatus string and error only based on jenkins job status
func (mock *mock) GetJobIdAndStatus(queueid string, jobname string) (string, string, error) {
	for _, job := range mock.jobs {
		if job.name == jobname {
			qid, err := strconv.ParseInt(queueid, 10, 0)
			if err != nil {
				break
			}
			job.RLock()
			if jobid, exist := job.queue[int(qid)]; exist {
				job.RUnlock()
				// TODO: always return success here?
				return string(jobid), SUCCESS, nil
			}
			job.RUnlock()
		}
	}
	return "", "", errors.New("No jobid found")
}

// Input queueid and jobname string
// Return jobid, jobstatus string and error
func (mock *mock) GetJobIdByQueueId(queueid string, jobname string) (string, string, error) {
	for _, job := range mock.jobs {
		if job.name == jobname {
			qid, err := strconv.ParseInt(queueid, 10, 0)
			if err != nil {
				break
			}
			job.RLock()
			if jobid, exist := job.queue[int(qid)]; exist {
				job.RUnlock()
				// TODO: always return success here?
				return string(jobid), SUCCESS, nil
			}
			job.RUnlock()
		}
	}
	return "", "", errors.New("No jobid found")
}

func (mock *mock) exist(jobname, jobid string) bool {
	for _, job := range mock.jobs {
		if job.name == jobname {
			job.RLock()
			for _, id := range job.queue {
				if jobid == string(id) {
					job.RUnlock()
					return true
				}
			}
			job.RUnlock()
			return false
		}
	}
	return false
}
