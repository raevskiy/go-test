package integrationtest

import (
	"bytes"
	"context"
	"cruder/internal/core"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetAllUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	_, router := core.SetupAppLayers(db)

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
	_, router := core.SetupAppLayers(db)

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
	_, router := core.SetupAppLayers(db)

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
	_, router := core.SetupAppLayers(db)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/username/klaasje", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNotFound)
	assertThatErrorMessageIsExpected(t, responseRecorder, "users not found")
}

func TestDeleteUserByUuid_Success(t *testing.T) {
	harryUsername := "tequila_sunset"
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db)
	user, err := repositories.Users.GetByUsername(harryUsername)
	if user == nil || err != nil {
		t.Fatalf("user %s is expected to be present in the DB", harryUsername)
	}

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/" + uuidHarry.String(), nil)
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
	repositories, router := core.SetupAppLayers(db)

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
	repositories, router := core.SetupAppLayers(db)

	randomUuid, _ := uuid.NewRandom()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/users/" + randomUuid.String(), nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNotFound)
	assertThatErrorMessageIsExpected(t, responseRecorder, "users not found")
	users, _ := repositories.Users.GetAll()
	if len(users) != 2 {
		t.Fatalf("expected all users remain present in the DB")
	}
}

func TestPatchUserByUuid_Success(t *testing.T) {
	harryUsername := "tequila_sunset"
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db)

	body := `{"full_name": {"value": "Raphaël Ambrosius Costeau"}}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/" + uuidHarry.String(), bytes.NewBuffer([]byte(body)))
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
	repositories, router := core.SetupAppLayers(db)

	body := `{"full_name": null}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/" + uuidHarry.String(), bytes.NewBuffer([]byte(body)))
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
	repositories, router := core.SetupAppLayers(db)

	body := `{"full_name": {"value": null}}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/" + uuidHarry.String(), bytes.NewBuffer([]byte(body)))
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
	repositories, router := core.SetupAppLayers(db)

	body := `{"full_name": {}}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/" + uuidHarry.String(), bytes.NewBuffer([]byte(body)))
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
	repositories, router := core.SetupAppLayers(db)

	body := `{"username": "kim"}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/" + uuidHarry.String(), bytes.NewBuffer([]byte(body)))
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
	repositories, router := core.SetupAppLayers(db)

	body := `{"email": "kim.kitsuragi@rcm.org"}`
	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/" + uuidHarry.String(), bytes.NewBuffer([]byte(body)))
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
	_, router := core.SetupAppLayers(db)
	body := `{"username": "klaasje"}`
	randomUuid, _ := uuid.NewRandom()

	req, _ := http.NewRequest(
		http.MethodPatch, "/api/v1/users/" + randomUuid.String(), bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNotFound)
	assertThatErrorMessageIsExpected(t, responseRecorder, "users not found")
}

func TestCreateUser_Success(t *testing.T) {
	klaasjeUserName := "klaasje"
	klaasjeFullName := "Klaasje Amandou"
	klaasjeEmail := "klaasje.amandou@noname.com"
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDbWithTestData(t)
	repositories, router := core.SetupAppLayers(db)

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
	if (err != nil) {
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
	repositories, router := core.SetupAppLayers(db)

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
	if (err != nil) {
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
	_, router := core.SetupAppLayers(db)

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
	_, router := core.SetupAppLayers(db)

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
	_, router := core.SetupAppLayers(db)

	body := `{"username": "klaasje", "email": "kim.kitsuragi@rcm.org"}`
	req, _ := http.NewRequest(
		http.MethodPost, "/api/v1/users", bytes.NewBuffer([]byte(body)))
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusConflict)
	assertThatErrorMessageIsExpected(t, responseRecorder, "the email is already in use")
}

func prepareDbWithTestData(t *testing.T) (*sql.DB, uuid.UUID, uuid.UUID) {
	t.Helper()

	db := prepareDb(t)
	uuidHarry, uuidKim := insertTestData(t, db)

	return db, uuidHarry, uuidKim
}

func prepareDb(t *testing.T) *sql.DB {
	t.Helper()

	dataSourceName := startPostgresContainer(t)
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	runMigrations(t, db, "../../migrations", "../../migrations_test")

	return db
}

func startPostgresContainer(t *testing.T) string {
	t.Helper()

	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("crudertestdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		postgresContainer.Terminate(ctx)
	})

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}
	port, err := postgresContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}
	dataSourceName := fmt.Sprintf(
		"postgres://postgres:postgres@%s:%s/crudertestdb?sslmode=disable", host, port.Port())

	return dataSourceName
}

func runMigrations(t *testing.T, db *sql.DB, migrationsDir string, testMigrationsDir string) {
	t.Helper()

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set dialect: %v", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	if err := goose.Up(db, testMigrationsDir); err != nil {
		t.Fatalf("failed to run integration test migrations: %v", err)
	}
}

func insertTestData(t *testing.T, db *sql.DB) (uuid.UUID, uuid.UUID) {
	t.Helper()

	uuidHarry, errHarry := uuid.NewRandom()
	uuidKim, errKim := uuid.NewRandom()
	if errHarry != nil || errKim != nil {
		t.Fatalf("failed to generate UUIDs")
	}
	_, err := db.ExecContext(context.Background(), `
        INSERT INTO users (uuid, username, email, full_name)
        VALUES
        ($1, 'tequila_sunset', 'harrier.dubois@rcm.org', 'Harrier Du Bois'),
        ($2, 'kim',   'kim.kitsuragi@rcm.org', 'Kim Kitsuragi');
    `, uuidHarry, uuidKim)
	if err != nil {
		t.Fatalf("failed to insert data: %v", err)
	}

	return uuidHarry, uuidKim
}

func assertThatResponseCodeIsExpected(t *testing.T, responseRecorder *httptest.ResponseRecorder, expectedCode int) {
	t.Helper()

	if responseRecorder.Code != expectedCode {
		t.Fatalf("expected %d, got %d", expectedCode, responseRecorder.Code)
	}
}

func assertThatUserFieldsAreExpected(
	t *testing.T,
	user map[string]interface{},
	expectedUserName string,
	expectedFullName string,
	expectedEmail string) {
	t.Helper()

	if user["username"] != expectedUserName ||
		user["full_name"] != expectedFullName ||
		user["email"] != expectedEmail {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func assertThatUsernameAndEmailAreExpected(
	t *testing.T,
	user map[string]interface{},
	expectedUserName string,
	expectedEmail string) {
	t.Helper()

	if user["username"] != expectedUserName ||
		user["full_name"] != nil ||
		user["email"] != expectedEmail {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func assertThatErrorMessageIsExpected(t *testing.T, responseRecorder *httptest.ResponseRecorder, expectedMessage string) {
	t.Helper()

	var jsonError map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &jsonError); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if jsonError["error"] != expectedMessage {
		t.Fatalf("unexpected error message: %+v", jsonError)
	}
}

