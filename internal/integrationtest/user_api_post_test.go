package integrationtest

import (
	"bytes"
	"cruder/internal/core"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateUser_Success(t *testing.T) {
	klaasjeUserName := "klaasje"
	klaasjeFullName := "Klaasje Amandou"
	klaasjeEmail := "klaasje.amandou@noname.com"
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	body := `{"username": "klaasje", "full_name": "Klaasje Amandou", "email": "klaasje.amandou@noname.com"}`
	req, _ := http.NewRequest(
		http.MethodPost, "/api/v1/users", bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusCreated)
	var userResponse map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &userResponse); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	assertThatUserFieldsAreExpected(t, userResponse,
		klaasjeUserName, klaasjeFullName, klaasjeEmail)
	user, err := repositories.Users.GetByUsername(klaasjeUserName)
	if err != nil {
		t.Fatalf("user %s cannot be obtained from the DB", klaasjeUserName)
	}
	if user.FullName.String != klaasjeFullName || user.Email != klaasjeEmail {
		t.Fatalf("user %s has an unxpected fields", klaasjeUserName)
	}
}

func TestCreateUserWithoutFullName_Success(t *testing.T) {
	klaasjeUserName := "klaasje"
	klaasjeEmail := "klaasje.amandou@noname.com"
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db, "")

	body := `{"username": "klaasje", "email": "klaasje.amandou@noname.com"}`
	req, _ := http.NewRequest(
		http.MethodPost, "/api/v1/users", bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusCreated)
	var userResponse map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &userResponse); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	assertThatUsernameAndEmailAreExpected(t, userResponse,
		klaasjeUserName, klaasjeEmail)
	user, err := repositories.Users.GetByUsername(klaasjeUserName)
	if err != nil {
		t.Fatalf("user %s cannot be obtained from the DB", klaasjeUserName)
	}
	if user.FullName.Valid || user.Email != klaasjeEmail {
		t.Fatalf("user %s has an unxpected fields", klaasjeUserName)
	}
}

func TestCreateUserOnSlightlyDifferentUrl_Success(t *testing.T) {
	klaasjeUserName := "klaasje"
	klaasjeEmail := "klaasje.amandou@noname.com"
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, "")

	body := `{"username": "klaasje", "email": "klaasje.amandou@noname.com"}`
	req, _ := http.NewRequest(
		http.MethodPost, "/api/v1/users/", bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusCreated)
	var userResponse map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &userResponse); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	assertThatUsernameAndEmailAreExpected(t, userResponse,
		klaasjeUserName, klaasjeEmail)
}

func TestCreateUserWithExistingUsername_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, "")

	body := `{"username": "kim", "email": "klaasje.amandou@noname.com"}`
	req, _ := http.NewRequest(
		http.MethodPost, "/api/v1/users", bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusConflict)
	assertThatErrorMessageIsExpected(t, responseRecorder, "the username is already taken")
}

func TestCreateUserWithExistingEmail_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db, "")

	body := `{"username": "klaasje", "email": "kim.kitsuragi@rcm.org"}`
	req, _ := http.NewRequest(
		http.MethodPost, "/api/v1/users", bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusConflict)
	assertThatErrorMessageIsExpected(t, responseRecorder, "the email is already in use")
}
