package integrationtest

import (
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
	db, _, _ := prepareDb(t)
	_, router := core.SetupAppLayers(db)

	req, _ := http.NewRequest("GET", "/api/v1/users/", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusOK)
	var users []map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &users); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(users) != 3 + 2 {
		t.Fatalf("expected 3 users from migration and 2 users added in this test, got %d", len(users))
	}
	assertThatUserNamesAreExpected(t, users[0], "asmith", "Alice Smith")
	assertThatUserNamesAreExpected(t, users[4], "kim", "Kim Kitsuragi")
}

func TestGetUserByUsername_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDb(t)
	_, router := core.SetupAppLayers(db)

	req, _ := http.NewRequest("GET", "/api/v1/users/username/kim", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusOK)
	var user map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &user); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	assertThatUserNamesAreExpected(t, user, "kim", "Kim Kitsuragi")
}

func TestGetUserByNonExistentUsername_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDb(t)
	_, router := core.SetupAppLayers(db)

	req, _ := http.NewRequest("GET", "/api/v1/users/username/klaasje", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNotFound)
	assertThatErrorMessageIsExpected(t, responseRecorder, "users not found")
}

func TestDeleteUserByUuid_Success(t *testing.T) {
	harryUsername := "tequila_sunset";
	gin.SetMode(gin.TestMode)
	db, uuidHarry, _ := prepareDb(t)
	repositories, router := core.SetupAppLayers(db)
	user, err := repositories.Users.GetByUsername(harryUsername)
	if user == nil || err != nil {
		t.Fatalf("user %s is expected to be present in the DB", harryUsername)
	}

	req, _ := http.NewRequest("DELETE", "/api/v1/users/" + uuidHarry.String(), nil)
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
	db, _, _ := prepareDb(t)
	repositories, router := core.SetupAppLayers(db)

	req, _ := http.NewRequest("DELETE", "/api/v1/users/worst-uuid-ever", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusBadRequest)
	assertThatErrorMessageIsExpected(t, responseRecorder, "invalid UUID")
	users, _ := repositories.Users.GetAll()
	if len(users) != 3 + 2 {
		t.Fatalf("expected all users remain present in the DB")
	}
}

func TestDeleteUserByINonExistentUuid_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _, _ := prepareDb(t)
	repositories, router := core.SetupAppLayers(db)

	randomUuid, _ := uuid.NewRandom()
	req, _ := http.NewRequest("DELETE", "/api/v1/users/" + randomUuid.String(), nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	assertThatResponseCodeIsExpected(t, responseRecorder, http.StatusNotFound)
	assertThatErrorMessageIsExpected(t, responseRecorder, "users not found")
	users, _ := repositories.Users.GetAll()
	if len(users) != 3 + 2 {
		t.Fatalf("expected all users remain present in the DB")
	}
}

func prepareDb(t *testing.T) (*sql.DB, uuid.UUID, uuid.UUID) {
	t.Helper()

	dataSourceName := startPostgresContainer(t)
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	runMigrations(t, db, "../../migrations")
	uuidHarry, uuidKim := insertTestData(t, db)

	return db, uuidHarry, uuidKim
}

func startPostgresContainer(t *testing.T) string {
	t.Helper()

	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("testdb"),
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
		"postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port.Port())

	return dataSourceName
}

func runMigrations(t *testing.T, db *sql.DB, migrationsDir string) {
	t.Helper()

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set dialect: %v", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
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
        ($2, 'kim',   'kim.kitsuragi@rcm.com', 'Kim Kitsuragi');
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

func assertThatUserNamesAreExpected(t *testing.T, user map[string]interface{}, expectedUserName string, expectedFullName string) {
	t.Helper()

	if user["username"] != expectedUserName || user["full_name"] != expectedFullName {
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

