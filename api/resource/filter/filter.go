package filter

import (
	"fmt"
	"github.com/gin-gonic/gin"
	. "hfrd/api/resource"
	"hfrd/api/utils/hfrdlogging"
	"hfrd/api/utils/jenkins"
	"net/http"
)

var logger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_FILTER)

// check whether the end user provided a valid environment parameter
func CheckEnv(c *gin.Context) {
	var env string
	switch c.Request.Method {
	// we put env into post body. Need to handle it differently
	case http.MethodPost:
		var plan Plan
		plan.Env = DEFAULT_ENV
		err := c.BindJSON(&plan)
		if err != nil {
			logger.Warningf("Error binding post body to plan object: %s", err)
			c.AbortWithStatusJSON(http.StatusBadRequest,
				gin.H{"message": fmt.Sprintf("Unable to bind json: %s", err)})
			return
		}
		env = plan.Env
		c.Set("plan", plan)
	default:
		var exist bool
		env, exist = c.GetQuery(ENV)
		// user didn't provide env param, we will bypass the check and use default env
		if !exist {
			return
		}
	}
	// user provided an empty env. Reject it
	if env == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"message": fmt.Sprintf("%s is empty", ENV)})
		return
	}
	// check whether we support the env user provided
	if _, exist := SUPPORTED_ENV[env]; !exist {
		var supported []string
		for k := range SUPPORTED_ENV {
			supported = append(supported, k)
		}
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"message": fmt.Sprintf("Currently supported env %v", supported)})
		return
	}
}

// check whether request id exists as query param
func CheckReqId(c *gin.Context) {
	if requestid := c.Query(REQUEST_ID_REQ); requestid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"message": fmt.Sprintf("%s is required", REQUEST_ID_REQ)})
		return
	}
}

// check whether service id exists as path param
func CheckServiceId(c *gin.Context) {
	if serviceid := c.Param(jenkins.SERVICE_ID); serviceid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"message": fmt.Sprintf("%s is required", jenkins.SERVICE_ID)})
		return
	}
}

// ensure path param uid is not empty
func CheckUid(c *gin.Context) {
	if uid := c.Param("uid"); uid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"message": fmt.Sprintf("%s is empty", UID)})
		return
	}
}
