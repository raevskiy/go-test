package integrationtest

import (
	"cruder/internal/core"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

const apiHeaderKey = "X-API-Key"

func TestGetUserByUsernameWithCorrectXApiKey_Success(t *testing.T) {
	key := "Les Cles de Fort Boyard"
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, key)

	req, _ := http.NewRequest(
		http.MethodGet, "/api/v1/users/username/kim",
		nil)
	req.Header.Set(apiHeaderKey, key)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusOK)
	var user map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &user); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	assertThatUserFieldsAreExpected(t, user,
		"kim", "Kim Kitsuragi", "kim.kitsuragi@rcm.org")
}

func TestGetUserByUsernameWithMissingXApiKey_Failure(t *testing.T) {
	key := "Les Cles de Fort Boyard"
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, key)

	req, _ := http.NewRequest(
		http.MethodGet, "/api/v1/users/username/kim",
		nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusUnauthorized)
	assertThatErrorMessageIsExpected(t, responseRecorder, "X-API-Key header is missing")
}

func TestGetUserByUsernameWithWrongXApiKey_Failure(t *testing.T) {
	key := "Les Cles de Fort Boyard"
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, key)

	req, _ := http.NewRequest(
		http.MethodGet, "/api/v1/users/username/kim",
		nil)
	req.Header.Set(apiHeaderKey, "Telegram end-to-end encryption keys, 2 pcs")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusForbidden)
	assertThatErrorMessageIsExpected(t, responseRecorder, "Invalid API key")
}
