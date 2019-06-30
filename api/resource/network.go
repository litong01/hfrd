package resource

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"hfrd/api/utils"
	"hfrd/api/utils/hfrdlogging"
	"hfrd/api/utils/jenkins"

	"hfrd/api/utils/couch"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var thisLogger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_NETWORK)

func NetworkPost(c *gin.Context) {
	thisLogger.Debugf("Start the Network creation process")
	uid := c.Param("uid")
	reqId := uuid.New().String()[0:8] + "-n"
	rootPath := utils.GetValue("contentRepo").(string) + "/" + uid + "/" + reqId
	os.MkdirAll(rootPath, 0755)
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	files := form.File["kubeconfig"]
	for _, file := range files {
		if err := c.SaveUploadedFile(file, rootPath+"/kubeconfig.zip"); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
	}
	files = form.File["config"]
	for _, file := range files {
		if err := c.SaveUploadedFile(file, rootPath+"/networkspec.yml"); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
	}

	var params = map[string]string{jenkins.REQUESTID: reqId, jenkins.UID: uid}
	thisLogger.Debug(utils.PrettyMapString(params))
	queueid, err := jks.TriggerJob(jenkins.K8SNETWORK, params)
	if err != nil {
		c.String(http.StatusServiceUnavailable, err.Error())
		return
	}
	ioutil.WriteFile(rootPath+"/"+QUEUEID, []byte(queueid), 0664)
	c.JSON(http.StatusOK,
		gin.H{"reqId": reqId, "jobId": queueid,
			"baseUrl": utils.GetValue("jenkins.baseUrl").(string),
			"jobName": jenkins.K8SNETWORK})
}

func NetworkGet(c *gin.Context) {
	nets, err := GetItems(c, "*-n")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{})
	}
	c.JSON(http.StatusOK, nets)
}

func NetworkDelete(c *gin.Context) {
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

func AddOrgPost(c *gin.Context) {
	thisLogger.Debugf("Start the Add Org process")
	uid := c.Param("uid")

	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	networkId := form.Value["network-id"][0]
	orgName := form.Value["orgName-"+networkId][0]
	channels := form.Value["channels-"+networkId][0]
	rootPath := utils.GetValue("contentRepo").(string) + "/" + uid + "/" + networkId
	var fileName string
	for _, file := range form.File["orgcerts-"+networkId] {
		if err := c.SaveUploadedFile(file, rootPath+"/"+file.Filename); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
		fileName = file.Filename
	}

	thisLogger.Debugf("uid: %s , networkId : %s channels: %s orgName: %s certsFileName: %s", uid, networkId, channels, orgName, fileName)

	if orgName == "" || channels == "" || fileName == "" {
		c.String(http.StatusBadRequest, fmt.Sprintf("No enough input artifacts. Make sure you have provided all required artifacts"))
		return
	}

	var params = map[string]string{jenkins.REQUESTID: networkId, jenkins.UID: uid, jenkins.ORGNAME: orgName, jenkins.ORGCERTSFILE: fileName, jenkins.CHANNELS: channels}
	thisLogger.Debug(utils.PrettyMapString(params))
	//queueid, err := jks.TriggerJob(jenkins.ADDORG, params)
	//if err != nil {
	//	c.String(http.StatusServiceUnavailable, err.Error())
	//	return
	//}
	//ioutil.WriteFile(rootPath+"/"+QUEUEID, []byte(queueid), 0664)
	//c.JSON(http.StatusOK,
	//	gin.H{"reqId": networkId, "jobId": queueid,
	//		"baseUrl": utils.GetValue("jenkins.baseUrl").(string),
	//		"jobName": jenkins.ADDORG})
}

func IbpCertsPost(c *gin.Context) {
	thisLogger.Debugf("Start the Ibp certs generation process")
	uid := c.Param("uid")
	reqId := uuid.New().String()[0:8] + "-x"
	rootPath := utils.GetValue("contentRepo").(string) + "/" + uid + "/" + reqId
	os.MkdirAll(rootPath, 0755)
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	files := form.File["service_config"]
	for _, file := range files {
		if err := c.SaveUploadedFile(file, rootPath+"/service_config.json"); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
	}
	var jobName string
	var params = map[string]string{jenkins.REQUESTID: reqId, jenkins.UID: uid, "contentrepo": contentRepo, "service_config": "service_config.json"}
	thisLogger.Debug(utils.PrettyMapString(params))
	jobName = jenkins.IBPCERTS
	queueid, err := jks.TriggerJob(jobName, params)
	if err != nil {
		c.String(http.StatusServiceUnavailable, err.Error())
		return
	}
	ioutil.WriteFile(rootPath+"/"+QUEUEID, []byte(queueid), 0664)
	//apacheBaseUrl := utils.GetValue("apacheBaseUrl").(string)
	job := couch.Job{ServiceId: reqId, QueueId: queueid, Name: jobName, ID: jobName + couch.SEPARATOR + queueid,
		Method: jenkins.POST, User: c.Param("uid"), PlanName: "N/A", Env: "N/A", Location: "N/A"}
	go saveJob(&job)
	c.Header(REQUEST_ID_RES, queueid)
	c.String(http.StatusAccepted, "Your request has been accepted ")

}

func CertsJobDelete(c *gin.Context) {
	uid := c.Param("uid")
	reqId := c.Query("requestid")
	version := c.Query("version")
	thisLogger.Debug("received parameters uid=" + uid + " \nreqId=" + reqId + "\nversion=" + version)
	_, err := couch.DeleteCertsJobById(reqId, version)
	if err != nil {
		thisLogger.Error("Delete certs job error occured: " + err.Error())
	}
	c.JSON(http.StatusOK,
		gin.H{"reqId": reqId,
			"jobName": jenkins.IBPCERTS})
}
