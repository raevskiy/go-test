package integrationtest

import (
	"cruder/internal/core"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteUserByUuid_Success(t *testing.T) {
	harryUsername := "tequila_sunset"
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")
	user, err := repositories.Users.GetByUsername(harryUsername)
	if user == nil || err != nil {
		t.Fatalf("user %s is expected to be present in the DB", harryUsername)
	}

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/"+uuidHarry.String(), nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNoContent)
	user, err = repositories.Users.GetByUsername(harryUsername)
	if user != nil || err == nil {
		t.Fatalf("user %s is expected to be absent in the DB", harryUsername)
	}
}

func TestDeleteUserByInvalidUuid_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/worst-uuid-ever", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusBadRequest)
	assertThatErrorMessageIsExpected(t, responseRecorder, "invalid UUID")
	users, _ := repositories.Users.GetAll()
	if len(users) != 2 {
		t.Fatalf("expected all users remain present in the DB")
	}
}

func TestDeleteUserByINonExistentUuid_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	randomUuid, _ := uuid.NewRandom()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/"+randomUuid.String(), nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNotFound)
	assertThatErrorMessageIsExpected(t, responseRecorder, "users not found")
	users, _ := repositories.Users.GetAll()
	if len(users) != 2 {
		t.Fatalf("expected all users remain present in the DB")
	}
}
