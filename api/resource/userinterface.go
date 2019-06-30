package resource

import (
	"fmt"
	"net/http"
	"net/url"

	"hfrd/api/utils"
	"hfrd/api/utils/hfrdlogging"

	"github.com/gin-gonic/gin"
)

var (
	interLogger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_UI)
	baseUrl     = utils.GetValue("jenkins.baseUrl").(string)
	apacheBase  = utils.GetValue("apacheBaseUrl").(string)
)

func init() {
	// remove username and password credentials from jenkins baseurl
	rawUrl, err := url.Parse(baseUrl)
	if err != nil {
		baseUrl = ""
	} else {
		rawUrl.User = nil
		baseUrl = rawUrl.String()
	}
	interLogger.Debugf("jenkins baseUrl is %s", baseUrl)
}

func NetworkUI(c *gin.Context) {
	nets, err := GetList(c, "*-n")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	c.HTML(http.StatusOK, "networkUI.html",
		gin.H{"uid": c.Param("uid"), "nets": nets,
			"baseUrl": baseUrl})
}

func ModuleTestUI(c *gin.Context) {
	tests, err := GetList(c, "*-t")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	c.HTML(http.StatusOK, "moduleTestUI.html",
		gin.H{"uid": c.Param("uid"), "tests": tests,
			"baseUrl": baseUrl})
}

func IbpUI(c *gin.Context) {
	uid := c.Param("uid")
	c.HTML(http.StatusOK, "ibpUI.html", gin.H{"uid": uid, "jenkinsBase": baseUrl, "apacheBase": apacheBase})
}

func IcpUI(c *gin.Context) {
	nets, err := GetList(c, "*-i")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	c.HTML(http.StatusOK, "icpUI.html",
		gin.H{"uid": c.Param("uid"), "nets": nets,
			"baseUrl": baseUrl})
}

func MainUI(c *gin.Context) {
	c.HTML(http.StatusOK, "mainUI.html", gin.H{})
}

func IbpCerts(c *gin.Context) {
	uid := c.Param("uid")
	c.HTML(http.StatusOK, "ibpCerts.html", gin.H{"uid": uid, "jenkinsBase": baseUrl, "apacheBase": apacheBase})
}
