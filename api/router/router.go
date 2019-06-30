package router

import (
	res "hfrd/api/resource"
	"hfrd/api/resource/auth"
	"hfrd/api/resource/filter"
	"hfrd/api/utils"
	"hfrd/api/utils/hfrdlogging"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var routerLogger = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_ROUTER)

func Router(logger bool) *gin.Engine {
	router := gin.New()
	if logger {
		router.Use(gin.Logger())
	}
	// Enable CORS for swagger UI
	var allowOrigins []string
	if rawOrigins, ok := utils.GetValue("allowOrigins").([]interface{}); ok {
		for _, rawOrigin := range rawOrigins {
			if strOrigin, ok := rawOrigin.(string); ok {
				if strings.HasPrefix(strOrigin, "http://") || strings.HasPrefix(strOrigin, "https") {
					allowOrigins = append(allowOrigins, strOrigin)
				} else {
					routerLogger.Warningf("bad origin %s, should start with http:// or https://",
						strOrigin)
				}
			}
		}
	}
	routerLogger.Debugf("allow orgins: %v", allowOrigins)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Access-Control-Allow-Headers"},
		ExposeHeaders:    []string{"Content-Length", "Request-Id"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return false
		},
		MaxAge: 12 * time.Hour,
	}))
	router.Use(gin.Recovery())
	router.Static("/static", "./static/content")
	router.LoadHTMLGlob("./static/templates/*")
	router.Static("/swagger", "./static/swagger-ui/dist")

	authInstance := auth.Auth()
	defer authInstance.Stop()

	service := router.Group("/v1/:uid/service", filter.CheckUid, authInstance.Handler())
	{
		service.GET("", filter.CheckReqId, filter.CheckEnv, res.ServiceGetByRequestId)
		service.POST("", filter.CheckEnv, res.ServicePost)
		service.DELETE("", filter.CheckReqId, filter.CheckEnv, res.ServiceDeleteByRequestId)
		service.GET("/:serviceid", filter.CheckEnv, res.ServiceGetByServiceId)
		service.DELETE("/:serviceid", filter.CheckEnv, res.ServiceDeleteByServiceId)
	}

	router.GET("/v1/:uid/services", filter.CheckUid, authInstance.Handler(), res.GetAllServices)
	router.GET("/v1/:uid/pending", filter.CheckUid, authInstance.Handler(), res.GetPendingCreate)
	router.GET("/v1/:uid/certsPending", filter.CheckUid, authInstance.Handler(), res.GetCertsPendingJobs)
	router.GET("/v1/:uid/certsService", filter.CheckUid, authInstance.Handler(), res.GetAllCertsJob4IBP)

	test := router.Group("/v1/:uid/test", filter.CheckUid, authInstance.Handler())
	{
		test.POST("", res.TestPost)
		test.GET("", filter.CheckReqId, res.TestGet)
	}

	ui := router.Group("/v1/:uid/ui", filter.CheckUid, authInstance.Handler())
	{
		ui.GET("/network", res.NetworkUI)
		ui.GET("/moduletest", res.ModuleTestUI)
		ui.GET("/ibp", res.IbpUI)
		ui.GET("/icp", res.IcpUI)
		ui.GET("/ibpcerts", res.IbpCerts)
	}

	network := router.Group("/v1/:uid/network", filter.CheckUid, authInstance.Handler())
	{
		network.POST("", res.NetworkPost)
		network.GET("", res.NetworkGet)
		network.DELETE("", filter.CheckReqId, res.NetworkDelete)
	}

	// APIs for icp network management
	icpNetwork := router.Group("/v1/:uid/icpnet", filter.CheckUid, authInstance.Handler())
	{
		icpNetwork.POST("", res.IcpNetworkPost)
		icpNetwork.GET("", res.IcpNetworkGet)
		icpNetwork.DELETE("", filter.CheckReqId, res.IcpNetworkDelete)
	}

	moduletest := router.Group("/v1/:uid/moduletest", filter.CheckUid, authInstance.Handler())
	{
		moduletest.POST("", filter.ValidateCertsTar, filter.ValidateTestPlan, res.ModuleTestPost)
		moduletest.GET("", res.ModuleTestGet)
		moduletest.DELETE("", filter.CheckReqId, res.ModuleTestDelete)
	}

	mainui := router.Group("/", authInstance.Handler())
	{
		mainui.GET("", res.MainUI)
	}

	ibpcerts := router.Group("/v1/:uid/ibpcerts", filter.CheckUid, authInstance.Handler())
	{
		ibpcerts.DELETE("", filter.CheckReqId, res.CertsJobDelete)
		ibpcerts.POST("", res.IbpCertsPost)
	}

	addorg := router.Group("/v1/:uid/addorg", filter.CheckUid, authInstance.Handler())
	{
		addorg.POST("", res.AddOrgPost)
	}

	router.GET("/v1/:uid/job/:jobname/:queueid", filter.CheckUid, authInstance.Handler(), res.GetJobStatus)
	return router
}
