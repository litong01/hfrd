package jenkins

import (
	"encoding/json"
	"errors"
	"fmt"
	. "hfrd/api/utils"
	"hfrd/api/utils/hfrdlogging"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

var (
	StatusForbiddenErr = errors.New("http status forbidden error")
	logger             = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_JKS)
)

type jenkins struct {
	crumbRequestField string
	crumb             string
	baseUrl           string
	crumbUrl          string
	jobUrl            string
	jobStatusUrl      string
	buildUrl          string
}

func NewJenkins() Jenkins {
	// TODO: is crumb required here given we have put username:passwd in baseUrl???
	var crumbRequestField, crumb string
	var baseUrl = GetValue("jenkins.baseUrl").(string)
	var crumbUrl = GetValue("jenkins.crumbUrl").(string)
	var jobUrl = GetValue("jenkins.jobUrl").(string)
	var jobStatusUrl = GetValue("jenkins.jobStatusUrl").(string)
	var buildUrl = GetValue("jenkins.buildUrl").(string)
	switch "" {
	case baseUrl, crumbUrl, jobUrl, jobStatusUrl, buildUrl:
		// Jenkins won't work without these parameters
		panic("Required parameter(s) not provided")
	default:
		logger.Debugf("Successfully loaded all parameters")
	}
	return &jenkins{crumbRequestField, crumb,
		baseUrl, crumbUrl, jobUrl, jobStatusUrl, buildUrl}
}

// Currently, the jobname should be network[-cm], connection[-cm] or test
func (j *jenkins) GetConsoleText(jobname, jobid string) ([]byte, error) {
	path := "/job/" + jobname + "/" + jobid + "/consoleText"
	return j.doGetWithRetry(path)
}

// Currently, the jobname should be network[-cm], connection[-cm] or test
func (j *jenkins) GetServiceJson(jobname, jobid string) ([]byte, error) {
	path := "/job/" + jobname + "/" + jobid + "/artifact/workdir/results/service.json"
	return j.doGetWithRetry(path)
}

// Currently, the jobname should be network[-cm], connection[-cm] or test
func (j *jenkins) ServeTar(jobname string, jobid string) ([]byte, error) {
	path := "/job/" + jobname + "/" + jobid + "/artifact/workdir/results/package.tar"
	return j.doGetWithRetry(path)
}

// Currently, the jobname should be network[-cm], connection[-cm] or test
// Return queueid, error
func (j *jenkins) TriggerJob(jobname string, params map[string]string) (string, error) {
	path := strings.Replace(j.buildUrl, "{JobName}", jobname, 1)
	header, _, err := j.doPostWithRetry(path, params)
	if err != nil {
		return "", err
	}
	var parts = strings.Split(strings.Trim(header.Get("Location"), "/ "), "/")
	if len(parts) < 1 {
		return "", errors.New("Can not find Location from response header")
	}
	queueId := parts[len(parts)-1]
	logger.Debugf("Triggered job %s and got queue id %s", jobname, queueId)
	return queueId, nil
}

func (j *jenkins) GetJobIdAndStatus(queueid string, jobname string) (string, string, error) {
	var path = strings.Replace(
		strings.Replace(j.jobUrl, "{JobName}", jobname, 1), "{QueueId}", queueid, 1)
	data, err := j.doGetWithRetry(path)
	if err != nil {
		logger.Warningf("Error getting jobid by queueid: %s", err)
		return "", "", err
	}
	var body = string(data)
	re := regexp.MustCompile(`<id>[\d]*</id>`)
	var jobid = strings.Replace(strings.Replace(re.FindString(body), "<id>", "", 1), "</id>", "", 1)
	if jobid == "" {
		return "", "", errors.New("Cannot find the job, " +
			"the server might be busy. Wait few minutes and try again.")
	}
	re = regexp.MustCompile(`<result>[\w]*</result>`)
	var status = strings.Replace(strings.Replace(re.FindString(body), "<result>", "", 1), "</result>", "", 1)
	if status == "" {
		return jobid, "INPROGRESS", nil
	}
	return jobid, status, nil
}

// Return jobid, jobStatus, error.
func (j *jenkins) GetJobIdByQueueId(queueid string, jobname string) (string, string, error) {
	var path = strings.Replace(
		strings.Replace(j.jobUrl, "{JobName}", jobname, 1), "{QueueId}", queueid, 1)
	data, err := j.doGetWithRetry(path)
	if err != nil {
		logger.Warningf("Error getting jobid by queueid: %s", err)
		return "", "", err
	}
	var body = string(data)
	re := regexp.MustCompile(`<id>[\d]*</id>`)
	var jobid = strings.Replace(strings.Replace(re.FindString(body), "<id>", "", 1), "</id>", "", 1)
	if jobid == "" {
		return "", "", errors.New("Cannot find the job, " +
			"the server might be busy. Wait few minutes and try again.")
	}
	re = regexp.MustCompile(`<result>FAILURE</result>`)
	if re.FindString(body) != "" {
		return jobid, FAIL, nil
	}
	re = regexp.MustCompile(`<result>[\w]*</result>`)
	if re.FindString(body) == "" {
		return jobid, INPROGRESS, nil
	}
	result, err := j.getJobStatusByJobId(jobid, jobname)
	return jobid, result, err
}

func (j *jenkins) getJobStatusByJobId(jobid string, jobname string) (string, error) {
	var path = strings.Replace(
		strings.Replace(j.jobStatusUrl, "{JobName}", jobname, 1), "{JobId}", jobid, 1)
	data, err := j.doGetWithRetry(path)
	if err != nil {
		return "", err
	}
	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return "", err
	}
	return body["status"].(string), nil
}

// Send HTTP GET request to jenkins
// will return (response_body_bytes, nil) or (nil, error)
func (j *jenkins) doGet(path string, withCrumb bool) ([]byte, error) {
	url := j.baseUrl + path
	logger.Debugf("GET: %s", path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if withCrumb && j.crumbRequestField != "" && j.crumb != "" {
		req.Header.Add(j.crumbRequestField, j.crumb)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		if resp.StatusCode == http.StatusForbidden {
			return nil, StatusForbiddenErr
		}
		return nil, fmt.Errorf("Bad status: %s, path: %s", resp.Status, path)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Send http get request with retry ability
func (j *jenkins) doGetWithRetry(path string) ([]byte, error) {
	body, err := j.doGet(path, true)
	if err == StatusForbiddenErr {
		j.updateJenkinsCrumb()
		body, err = j.doGet(path, true)
	}
	return body, err
}

func (j *jenkins) doPost(path string, params map[string]string) (http.Header, []byte, error) {
	logger.Debugf("POST: %s", path)
	req, err := http.NewRequest("POST", j.baseUrl+path, nil)
	if err != nil {
		return nil, nil, err
	}
	q := req.URL.Query()
	if params != nil {
		for k, v := range params {
			q.Add(k, v)
		}
	}
	if j.crumbRequestField != "" && j.crumb != "" {
		req.Header.Set(j.crumbRequestField, j.crumb)
	}
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == http.StatusForbidden {
			return nil, nil, StatusForbiddenErr
		}
		return nil, nil, errors.New("Bad http response code")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp.Header, data, nil
}

func (j *jenkins) doPostWithRetry(path string, params map[string]string) (http.Header, []byte, error) {
	header, body, err := j.doPost(path, params)
	if err == StatusForbiddenErr {
		j.updateJenkinsCrumb()
		header, body, err = j.doPost(path, params)
	}
	return header, body, err
}

func (j *jenkins) updateJenkinsCrumb() {
	j.crumbRequestField, j.crumb, _ = j.getJenkinsCrumb()
}

func (j *jenkins) getJenkinsCrumb() (string, string, error) {
	logger.Info("Started getting Jenkins crumb.")
	data, err := j.doGet(j.crumbUrl, false)
	if err != nil {
		logger.Errorf("Error getting Jenkins crumb: %s", err)
		return "", "", err
	}
	var body map[string]interface{}
	if err = json.Unmarshal(data, &body); err != nil {
		logger.Errorf("Error unmarshaling response data: %s", err.Error())
		return "", "", err
	}
	logger.Info("Got Jenkins crumb successfully.")
	return body["crumbRequestField"].(string), body["crumb"].(string), nil
}
