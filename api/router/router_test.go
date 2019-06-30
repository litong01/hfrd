package router

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/http"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/gin-gonic/gin"
	res "hfrd/api/resource"
	"hfrd/api/utils/jenkins"
)

// Refer to https://github.com/gin-gonic/gin#testing

var router *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)
	router = Router(false)
}

func TestCheckUid(t *testing.T) {
	// Reset mock jenkins
	res.SetJenkins(jenkins.NewMockJenkins())

	// No uid provided
	w := httptest.NewRecorder()
	body := map[string]string{res.ENV: res.BX_STAGING}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/v1//service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// uid provided
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/testuser/service?requestid=123", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCheckRequestId(t *testing.T) {
	// Reset mock jenkins
	res.SetJenkins(jenkins.NewMockJenkins())

	// no requestid provided
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/testuser/service", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ------------------POST /v1/:uid/service------------------
// ************************GOOD PATHS************************
func TestServicePost(t *testing.T)  {
	// Set mock jenkins
	res.SetJenkins(jenkins.NewMockJenkins())

	// Empty body defaults to BX_STAGING, STARTER
	w := httptest.NewRecorder()
	body := map[string]string{}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/v1/testuser/service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.NotEmpty(t, w.Header().Get(res.REQUEST_ID_RES))

	w = httptest.NewRecorder()
	body = map[string]string{res.ENV: res.BX_STAGING}
	bodyBytes, _ = json.Marshal(body)
	req, _ = http.NewRequest("POST", "/v1/testuser/service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.NotEmpty(t, w.Header().Get(res.REQUEST_ID_RES))

	w = httptest.NewRecorder()
	body = map[string]string{res.ENV: res.BX_STAGING, res.PLAN_NAME: res.ENTERPRISE}
	bodyBytes, _ = json.Marshal(body)
	req, _ = http.NewRequest("POST", "/v1/testuser/service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.NotEmpty(t, w.Header().Get(res.REQUEST_ID_RES))

	w = httptest.NewRecorder()
	body = map[string]string{res.ENV: res.BX_PROD}
	bodyBytes, _ = json.Marshal(body)
	req, _ = http.NewRequest("POST", "/v1/testuser/service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.NotEmpty(t, w.Header().Get(res.REQUEST_ID_RES))

	w = httptest.NewRecorder()
	body = map[string]string{res.ENV: res.CM}
	bodyBytes, _ = json.Marshal(body)
	req, _ = http.NewRequest("POST", "/v1/testuser/service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.NotEmpty(t, w.Header().Get(res.REQUEST_ID_RES))

	w = httptest.NewRecorder()
	body = map[string]string{res.ENV: res.CM, res.PLAN_NAME: res.ENTERPRISE}
	bodyBytes, _ = json.Marshal(body)
	req, _ = http.NewRequest("POST", "/v1/testuser/service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.NotEmpty(t, w.Header().Get(res.REQUEST_ID_RES))
}

// ************************BAD PATHS************************
func TestServicePostBad(t *testing.T)  {
	// Reset mock jenkins
	res.SetJenkins(jenkins.NewMockJenkins())

	// Provided a bad env
	w := httptest.NewRecorder()
	body := map[string]string{res.ENV: "NotDefinedEnvironment"}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/v1/testuser/service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Bad plan name
	w = httptest.NewRecorder()
	body = map[string]string{res.ENV: res.BX_PROD, res.PLAN_NAME: "UndefinedPlanName"}
	bodyBytes, _ = json.Marshal(body)
	req, _ = http.NewRequest("POST", "/v1/testuser/service", bytes.NewBuffer(bodyBytes))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
