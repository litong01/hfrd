package main

import (
	"github.com/gin-gonic/gin"
	"hfrd/api/metadata"
	"hfrd/api/router"
	"hfrd/api/utils/hfrdlogging"
)

var logger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_MAIN)

func init() {
	logger.Infof("Starting hfrd api server:\n%s", metadata.GetVersion())
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := router.Router(true)
	if err := router.Run(); err != nil {
		logger.Errorf("Error running hfrdserver: %s", err)
	}
}
