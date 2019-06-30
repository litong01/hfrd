package resource

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetJobStatus(c *gin.Context) {
	jobName := c.Param("jobname")
	queueId := c.Param("queueid")
	if len(jobName) == 0 || len(queueId) == 0 {
		c.String(http.StatusBadRequest, "No jobname or queueid provided")
		return
	}
	jobId, status, err := jks.GetJobIdAndStatus(queueId, jobName)
	if err != nil {
		c.String(http.StatusNotFound, fmt.Sprintf("Unable to get job status: %s", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobId": jobId, "status": status})
}
