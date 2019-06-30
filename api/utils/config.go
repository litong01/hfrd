package utils

import (
	"encoding/json"
	"gopkg.in/op/go-logging.v1"
	"hfrd/api/utils/hfrdlogging"
	"io/ioutil"
	"strconv"
	"strings"
)

const defaultConfig = `
{
  "jenkins": {
    "baseUrl": "http://localhost:9090",
    "crumbUrl": "/crumbIssuer/api/json",
    "jobUrl": "/job/{JobName}/api/xml?tree=builds[id,result,queueId]&xpath=\/\/build[queueId={QueueId}]",
    "jobStatusUrl": "/job/{JobName}/{JobId}/artifact/workdir/results/jobStatus.json",
    "serviceGetByServiceId": "job/{JobName}}/{JobId}/artifact/workdir/results/package.tar",
    "buildUrl": "/job/{JobName}/buildWithParameters"
  },
  "log": {
    "level": "info"
  },
  "auth": {
    "enabled": false,
    "type": "jwt"
  },
  "couchUrl": "http://localhost:5984"
}
`

var config = getConfig()
var logger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_UTILS)

func init() {
	// Set log level from configuration
	logLevel := GetValue("log.level")
	if levelStr, ok := logLevel.(string); ok {
		if level, err := logging.LogLevel(levelStr); err == nil {
			hfrdlogging.SetLogLevel(level)
			logger.Infof("hfrdlogging level set to: %s", level.String())
		}
	}
}

func getConfig() interface{} {
	var config interface{}
	raw, err := ioutil.ReadFile("var/config.json")
	if err != nil {
		logger.Infof("Error reading config file: %s", err.Error())
		logger.Info("Will use default configurations")
		raw = []byte(defaultConfig)
	}

	json.Unmarshal(raw, &config)
	return config
}

func getValue(obj interface{}, key string) interface{} {
	md, _ := obj.(map[string]interface{})
	return md[key]
}

func GetValue(key string) interface{} {
	vals := strings.Split(key, ".")
	var md = config
	for i := 0; i < len(vals); i++ {
		if strings.Contains(vals[i], "[") {
			arr := strings.Split(strings.Replace(vals[i], "]", "[", 1), "[")
			idx, _ := strconv.Atoi(arr[1])
			el := getValue(md, arr[0]).([]interface{})
			md = el[idx]
		} else {
			md = getValue(md, vals[i])
		}
	}
	return md
}
