package integrationtest

import (
	"bytes"
	"cruder/internal/core"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPatchUserByUuid_Success(t *testing.T) {
	harryUsername := "tequila_sunset"
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	body := `{"full_name": {"value": "Raphaël Ambrosius Costeau"}}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/"+uuidHarry.String(), bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNoContent)
	user, _ := repositories.Users.GetByUsername(harryUsername)
	if user.FullName.String != "Raphaël Ambrosius Costeau" {
		t.Fatalf("user %s has an unxpected full name %s", harryUsername, user.FullName.String)
	}
}

func TestPatchUserByUuidWithNullFullName_Success(t *testing.T) {
	harryUsername := "tequila_sunset"
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	body := `{"full_name": null}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/"+uuidHarry.String(), bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNoContent)
	user, _ := repositories.Users.GetByUsername(harryUsername)
	if !user.FullName.Valid {
		t.Fatalf("user %s has an unxpected NULL full name", harryUsername)
	}
}

func TestPatchUserByUuidWithExplicitlyErasedFullName_Success(t *testing.T) {
	harryUsername := "tequila_sunset"
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	body := `{"full_name": {"value": null}}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/"+uuidHarry.String(), bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNoContent)
	user, _ := repositories.Users.GetByUsername(harryUsername)
	if user.FullName.Valid {
		t.Fatalf("user %s has an unxpected full name %s", harryUsername, user.FullName.String)
	}
}

func TestPatchUserByUuidWithImplicitlyErasedFullName_Success(t *testing.T) {
	harryUsername := "tequila_sunset"
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	body := `{"full_name": {}}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/"+uuidHarry.String(), bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNoContent)
	user, _ := repositories.Users.GetByUsername(harryUsername)
	if user.FullName.Valid {
		t.Fatalf("user %s has an unxpected full name %s", harryUsername, user.FullName.String)
	}
}

func TestPatchUserByUuidWithExistingUsername_Failure(t *testing.T) {
	kimUsername := "kim"
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	body := `{"username": "kim"}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/"+uuidHarry.String(), bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusConflict)
	assertThatErrorMessageIsExpected(t, responseRecorder, "the username is already taken")
	user, _ := repositories.Users.GetByUsername(kimUsername)
	if user.FullName.String != "Kim Kitsuragi" {
		t.Fatalf("user %s has an unxpected full name %s", kimUsername, user.FullName.String)
	}
}

func TestPatchUserByUuidWithExistingEmail_Failure(t *testing.T) {
	harryUsername := "tequila_sunset"
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	body := `{"email": "kim.kitsuragi@rcm.org"}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/"+uuidHarry.String(), bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusConflict)
	assertThatErrorMessageIsExpected(t, responseRecorder, "the email is already in use")
	user, _ := repositories.Users.GetByUsername(harryUsername)
	if user.Email != "harrier.dubois@rcm.org" {
		t.Fatalf("user %s has an unxpected email %s", harryUsername, user.Email)
	}
}

func TestPatchUserByUuidWithNonExistentUuid_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, "")
	body := `{"username": "klaasje"}`
	randomUuid, _ := uuid.NewRandom()

	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/"+randomUuid.String(), bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNotFound)
	assertThatErrorMessageIsExpected(t, responseRecorder, "users not found")
}
