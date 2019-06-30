package resource

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"hfrd/api/utils/jenkins"
	"hfrd/api/utils/couch"
)

func deleteService(c *gin.Context, env, serviceId string) {
	var params = map[string]string{ENV: env, jenkins.METHOD: jenkins.DELETE, jenkins.SERVICE_ID: serviceId}
	var jobName string
	switch env {
	case CM:
		jobName = jenkins.NETWORK_CM
	default:
		jobName = jenkins.NETWORK
	}
	var queueid, err = jks.TriggerJob(jobName, params)
	if err != nil {
		logger.Warningf("Can not trigger a job to delete the service: %s", err)
		c.String(http.StatusBadRequest, "Can not trigger a job to delete the service")
		return
	}
	// Job triggered in Jenkins server
	// start a go routine to store the job into DB and poll status of job
	job := couch.Job{QueueId: queueid, Name: jobName, ID: jobName + couch.SEPARATOR + queueid,
		Method: jenkins.DELETE, User: c.Param("uid"), ServiceId: serviceId}
	go saveJob(&job)
	c.Header(REQUEST_ID_RES, queueid)
	c.String(http.StatusAccepted, "Your request has been accepted in "+env+" environment.")
}

func serveJobTarByQueueId(c *gin.Context, queueId, jobName string) {
	var jobid, result, err = jks.GetJobIdByQueueId(queueId, jobName)
	if jobid == "" {
		c.String(http.StatusNotFound, "The job with requestid:" + queueId +
			" and jobname:" + jobName + " can not be found.")
		return
	}
	if err != nil {
		logger.Warningf("Can not get job status with requestid: %s. jobname: %s. status: \"%s\". Error: %s",
			queueId, jobName, result, err)
		c.String(http.StatusNotFound, fmt.Sprintf("Can not get job status with requestid: %s, jobname: %s",
			queueId, jobName))
		return
	}
	if result == jenkins.SUCCESS {
		serveJobTarByJobId(c, jobid, jobName)
		return
	}
	data, err := jks.GetConsoleText(jobName, jobid)
	if err != nil {
		logger.Warningf("Can not get console text: %s", err)
		c.String(http.StatusBadRequest, "Can not get console text")
		return
	}
	if result == jenkins.FAIL {
		c.String(http.StatusBadRequest, string(data))
	} else {
		c.String(http.StatusAccepted, string(data))
	}
}

func serveJobTarByJobId(c *gin.Context, jobId, jobName string) {
	data, err := jks.ServeTar(jobName, jobId)
	if err != nil {
		logger.Warningf("Can not get the package tar for the jobid: %s. jobname: %s. Error: %s",
			jobId, jobName, err)
		c.JSON(http.StatusOK, gin.H{"message:": "Status: Success. No package tar"})
	} else {
		sendBinary(c, data)
	}
}

func sendBinary(c *gin.Context, data []byte) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=package.tar")
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/gzip", data)
}

func contains(array []string, target string) bool {
	for _, v := range array {
		if v == target {
			return true
		}
	}
	return false
}
