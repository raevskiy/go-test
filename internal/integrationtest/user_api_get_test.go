package integrationtest

import (
	"cruder/internal/core"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusOK)
	var users []map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &users); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users added in this test, got %d", len(users))
	}
	assertThatUserFieldsAreExpected(t, users[0],
		"tequila_sunset", "Harrier Du Bois", "harrier.dubois@rcm.org")
	assertThatUserFieldsAreExpected(t, users[1],
		"kim", "Kim Kitsuragi", "kim.kitsuragi@rcm.org")
}

func TestGetAllUsersOnEmptyDb_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := prepareDb(t)
	_, router := core.SetupAppLayers(db, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusOK)
	var users []map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &users); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected no users at all, got %d", len(users))
	}
}

func TestGetUserByUsername_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/username/kim", nil)
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

func TestGetUserByNonExistentUsername_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, "")

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/username/klaasje", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNotFound)
	assertThatErrorMessageIsExpected(t, responseRecorder, "users not found")
}
