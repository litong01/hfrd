package resource

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"hfrd/api/utils"
	"hfrd/api/utils/jenkins"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func IcpNetworkPost(c *gin.Context) {
	thisLogger.Debugf("Start the ICP Network creation process")
	uid := c.Param("uid")
	reqId := uuid.New().String()[0:8] + "-i"
	rootPath := utils.GetValue("contentRepo").(string) + "/" + uid + "/" + reqId
	os.MkdirAll(rootPath, 0755)
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	files := form.File["config"]
	for _, file := range files {
		if err := c.SaveUploadedFile(file, rootPath+"/networkspec.yml"); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
	}

	var params = map[string]string{jenkins.REQUESTID: reqId, jenkins.UID: uid, jenkins.METHOD: jenkins.POST}
	thisLogger.Debug(utils.PrettyMapString(params))
	queueid, err := jks.TriggerJob(jenkins.NETWORK_ICP, params)
	if err != nil {
		c.String(http.StatusServiceUnavailable, err.Error())
		return
	}
	ioutil.WriteFile(rootPath+"/"+QUEUEID, []byte(queueid), 0664)
	c.JSON(http.StatusOK,
		gin.H{"reqId": reqId, "jobId": queueid,
			"baseUrl": utils.GetValue("jenkins.baseUrl").(string),
			"jobName": jenkins.NETWORK_ICP})
}

func IcpNetworkGet(c *gin.Context) {
	nets, err := GetItems(c, "*-i")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{})
	}
	c.JSON(http.StatusOK, nets)
}

func IcpNetworkDelete(c *gin.Context) {
	thisLogger.Debugf("Start the Network deletion process")
	uid := c.Param("uid")
	reqId := c.Query("requestid")
	var params = map[string]string{jenkins.REQUESTID: reqId, jenkins.UID: uid,
		jenkins.ACTION: "delete"}
	thisLogger.Debug(utils.PrettyMapString(params))
	queueid, err := jks.TriggerJob(jenkins.K8SNETWORK, params)
	if err != nil {
		c.String(http.StatusServiceUnavailable, err.Error())
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"reqId": reqId, "jobId": queueid,
			"baseUrl": utils.GetValue("jenkins.baseUrl").(string),
			"jobName": jenkins.K8SNETWORK})
}
