package resource

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"encoding/json"
	"fmt"
	"hfrd/api/utils/jenkins"
)

func TestGet(c *gin.Context) {
	requestid := c.Query(REQUEST_ID_REQ)
	serveJobTarByQueueId(c, requestid, jenkins.TEST)
}

func TestPost(c *gin.Context) {
	var test Test
	// Set sslcerts default value to true
	test.Sslcerts = true
	err := c.BindJSON(&test)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Unable to bind json: %s", err))
		return
	}
	testBytes, err := json.Marshal(test)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Unable to marshal json: %s", err))
		return
	}

	if test.Hash == "" || test.Startcmd == "" || test.Url == ""  {
		c.String(http.StatusBadRequest, "please provide required field(s)")
		return
	}
  
	logger.Debugf("testconfig: %s", string(testBytes))
	var params = map[string]string{jenkins.TESTCONFIG: string(testBytes)}
	queueid, err := jks.TriggerJob(jenkins.TEST, params)
	if err != nil {
		c.String(http.StatusServiceUnavailable, err.Error())
		return
	}
	c.Header(REQUEST_ID_RES, queueid)
	c.String(http.StatusAccepted, "Your request has been accepted.")
}
