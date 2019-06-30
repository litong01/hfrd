package jwt

import (
	"strings"
	"errors"
	"encoding/base64"
	"encoding/json"
	"bytes"
	"time"
	"hfrd/api/resource/auth/jwt/jwk"
	"github.com/gin-gonic/gin"
	"net/http"
	"hfrd/api/resource/auth/interfaces"
)

const (
	ISSUER = "https://iam.ng.bluemix.net/oidc/token"
	ALG = "RS256"
)

var (
	empty IbmJwt = IbmJwt{}

	MALFORMED = errors.New("Malformed JWT token")
	DECODE_FAILURE = errors.New("Decode JWT token failure")
	UNMARSHAL_FAILURE = errors.New("Unmarshal JWT token failure")
	UNEXPECTED_ISSUER = errors.New("The issuer in JWT is not supported")
	EXPIRED_TOKEN = errors.New("Expired token")
	ISSUE_AT_ERROR = errors.New("Token used before issued")
	UNSUPPORTED_ALG = errors.New("Unsupported algorithm")
)

func NewJwtAuth(manager jwk.IbmJwksManager) interfaces.Auth {
	manager.Init()
	return &JwtAuth{manager: manager}
}

// JWTAuth implements auth.Auth interface
type JwtAuth struct {
	manager jwk.IbmJwksManager
}

func (j *JwtAuth) Handler() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Parse Authorization header
		authorization := strings.TrimSpace(c.GetHeader("Authorization"))
		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No authorization header provided"})
			return
		}
		tokens := strings.Split(authorization, " ")
		if len(tokens) != 2 || tokens[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": MALFORMED.Error()})
			return
		}

		// Decode and verify JWT header, payload
		jwt, err := decodeJWT(tokens[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}

		// Verify JWT signature with JWKS
		parts := strings.Split(tokens[1], ".")
		// No need to check len(parts) because it's already checked in decodeJWT
		if err := j.manager.VerifySig(jwt.IbmHeader.Kid, parts[0] + "." + parts[1], parts[2]); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}
	}
}

func (j *JwtAuth) Stop() {
	j.manager.Stop()
}

func decodeJWT(tokenRawString string) (IbmJwt, error) {
	parts := strings.Split(tokenRawString, ".")
	if len(parts) != 3 {
		return empty, MALFORMED
	}
	// Decode the header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return empty, DECODE_FAILURE
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return empty, DECODE_FAILURE
	}

	var header IbmHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return empty, UNMARSHAL_FAILURE
	}
	var payload IbmPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return empty, UNMARSHAL_FAILURE
	}
	// Verify issuer
	if !verifiedIssuer(payload.Iss, ISSUER) {
		return empty, UNEXPECTED_ISSUER
	}
	// Verify Algorithm
	if !verifiedAlg(header.Alg) {
		return empty, UNSUPPORTED_ALG
	}
	// Verify expire time
	if expired(payload.Exp) {
		return empty, EXPIRED_TOKEN
	}
	// Verify JWT "issue at" time
	if before(payload.Iat) {
		return empty, ISSUE_AT_ERROR
	}
	return IbmJwt{header, payload, tokenRawString}, nil
}

func verifiedIssuer(actual, expected string) bool {
	if bytes.Compare([]byte(actual), []byte(expected)) != 0 {
		return false
	}
	return true
}

func expired(exp int64) bool {
	return time.Now().Unix() > exp
}

func before(iat int64) bool {
	return time.Now().Unix() < iat
}

func verifiedAlg(alg string) bool {
	if bytes.Compare([]byte(alg), []byte(ALG)) != 0 {
		return false
	}
	return true
}
