package interfaces

import "github.com/gin-gonic/gin"

type Auth interface {
	// Return gin auth handler
	Handler() (func(*gin.Context))
	// Release resources and stop component
	Stop()
}
