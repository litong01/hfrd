package auth

import (
	"github.com/gin-gonic/gin"
	"hfrd/api/resource/auth/interfaces"
	"hfrd/api/resource/auth/jwt"
	"hfrd/api/resource/auth/jwt/jwk"
	"hfrd/api/utils"
	"hfrd/api/utils/hfrdlogging"
	"strings"
)

const (
	JWT = "jwt"
)

var (
	logger      = hfrdlogging.MustGetLogger(hfrdlogging.MODULE_AUTH)
	nonAuth     = &NonAuth{}
	authEnabled bool
	authType    string
)

func init() {
	if enabled, ok := utils.GetValue("auth.enabled").(bool); ok {
		authEnabled = enabled
	}
	if auth, ok := utils.GetValue("auth.type").(string); ok {
		authType = auth
	}
}

func Auth() interfaces.Auth {
	if !authEnabled {
		logger.Info("Auth is disabled")
		return nonAuth
	}
	logger.Infof("Auth type from configuration: \"%s\"", authType)
	switch strings.ToLower(authType) {
	case JWT:
		return jwt.NewJwtAuth(jwk.NewIbmJwksManager())
	default:
		// Unsupported auth type
		logger.Warningf("Unsupported auth type: \"%s\"", authType)
		return nonAuth
	}
}

type NonAuth struct {
}

func (nonAuth *NonAuth) Handler() func(*gin.Context) {
	return func(*gin.Context) {}
}

func (NonAuth *NonAuth) Stop() {

}
