package resource

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/op/go-logging.v1"
	"hfrd/api/utils"
	"hfrd/api/utils/hfrdlogging"
	"hfrd/api/utils/jenkins"
	"io/ioutil"
	"net/http"
	"os"
)

var moduleLogger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_MODULETEST)

func ModuleTestPost(c *gin.Context) {
	//hfrdlogging.SetLogLevel(logging.DEBUG)
	fmt.Printf("log level of module logger:%s\n", logging.GetLevel(hfrdlogging.MODULE_MODULETEST))
	moduleLogger.Debugf("Start the module test process")
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	uid := c.Param("uid")
	var action string
	if c.Query("rerun") != "1" {
		action = "create"
	} else {
		action = form.Value["requestid"][0]
	}
	reqId := uuid.New().String()[0:8] + "-t"
	rootPath := utils.GetValue("contentRepo").(string) + "/" + uid + "/" + reqId
	os.MkdirAll(rootPath, 0755)
	if action == "create" {
		files := form.File["cert"]
		for _, file := range files {
			if err := c.SaveUploadedFile(file, rootPath+"/allcerts.tgz"); err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
				return
			}
			certVersion := c.GetString(CERT_VERSION_KEY)
			if certVersion == "" || certVersion == CERT_V1 {
				logger.Debug("Cert V1 detected")
			} else if certVersion == CERT_V2 {
				logger.Debug("Cert V2 detected")
				// TODO: convert the IBP 2.0 input into the format that hfrd test modules can read
				// TODO: The proposed IBP 2.0 tar folder structure is like below. May change later
				/*
					connection-profiles/
					├── Org1
					│   ├── connection.json
					│   └── identity.json
					└── Org2
					    ├── connection.json
					    └── identity.json
				*/
				// untar IBP 2.0 tar gz file
				if err := unTarGz(rootPath+"/allcerts.tgz", rootPath); err != nil {
					logger.Debugf("Error unTarGz allcerts.tgz: %s", err)
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				// convert IBP 2.0 SaaS connection profile and identity json
				if err := convertIBP2SaaS(rootPath+"/connection-profiles", rootPath+"/allcerts.tgz"); err != nil {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				// TODO: Should we clean up the test instance directory $rootPath?
			}
		}
		files = form.File["kubeconfig"]
		for _, file := range files {
			if err := c.SaveUploadedFile(file, rootPath+"/kubeconfig.zip"); err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
				return
			}
		}
		files = form.File["plan"]
		for _, file := range files {
			if err := c.SaveUploadedFile(file, rootPath+"/testplan.yml"); err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
				return
			}
		}

		files = form.File["chaincode"]
		for _, file := range files {
			if err := c.SaveUploadedFile(file, rootPath+"/chaincode.tgz"); err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
				return
			}
		}
	} else {
		files := form.File["testplan-"+action]
		for _, file := range files {
			if file.Filename != "" {
				if err := c.SaveUploadedFile(file, rootPath+"/testplan.yml"); err != nil {
					c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
					return
				}
			}

		}
	}
	var params = map[string]string{jenkins.REQUESTID: reqId, jenkins.UID: uid,
		jenkins.APACHE_BASE: apacheBase, jenkins.ACTION: action}
	moduleLogger.Debug(utils.PrettyMapString(params))
	queueid, err := jks.TriggerJob(jenkins.MODULETEST, params)
	if err != nil {
		c.String(http.StatusServiceUnavailable, err.Error())
		return
	}
	ioutil.WriteFile(rootPath+"/"+QUEUEID, []byte(queueid), 0664)
	c.JSON(http.StatusOK,
		gin.H{"reqId": reqId, "baseUrl": baseUrl,
			"jobId": queueid, "jobName": jenkins.MODULETEST,
			"metricsUrl": fmt.Sprintf("%s/%s/%s/metrics.html", apacheBase, uid, reqId)})
}

func ModuleTestGet(c *gin.Context) {
	tests, err := GetItems(c, "*-t")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{})
	}
	c.JSON(http.StatusOK, tests)
}

func ModuleTestDelete(c *gin.Context) {
	thisLogger.Debugf("Start the test deletion process")
	uid := c.Param("uid")
	reqId := c.Query("requestid")
	var params = map[string]string{jenkins.REQUESTID: reqId, jenkins.UID: uid,
		jenkins.ACTION: "delete"}
	thisLogger.Debug(utils.PrettyMapString(params))
	queueid, err := jks.TriggerJob(jenkins.MODULETEST, params)
	if err != nil {
		c.String(http.StatusServiceUnavailable, err.Error())
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"reqId": reqId, "baseUrl": baseUrl,
			"jobId": queueid, "jobName": jenkins.MODULETEST,
			"metricsUrl": fmt.Sprintf("%s/%s/%s/metrics.html", apacheBase, uid, reqId)})
}
