package integrationtest

import (
	"context"
	"cruder/internal/handler"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
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
	//Arrange
	gin.SetMode(gin.TestMode)
	db := prepareDb(t)
	router := handler.SetupAppLayersAndRouter(db)

	//Act
	req, _ := http.NewRequest("GET", "/api/v1/users/", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	//Assert
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, responseRecorder.Code)
	}
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

func TestGetUserByName_Success(t *testing.T) {
	//Arrange
	gin.SetMode(gin.TestMode)
	db := prepareDb(t)
	router := handler.SetupAppLayersAndRouter(db)

	//Act
	req, _ := http.NewRequest("GET", "/api/v1/users/username/kim", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	//Assert
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, responseRecorder.Code)
	}
	var user map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &user); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	assertThatUserNamesAreExpected(t, user, "kim", "Kim Kitsuragi")
}

func TestGetUserByName_Failure(t *testing.T) {
	//Arrange
	gin.SetMode(gin.TestMode)
	db := prepareDb(t)
	router := handler.SetupAppLayersAndRouter(db)

	//Act
	req, _ := http.NewRequest("GET", "/api/v1/users/username/klaasje", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, req)

	//Assert
	if responseRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, responseRecorder.Code)
	}
	var jsonError map[string]interface{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &jsonError); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	assertThatErrorMessageIsExpected(t, jsonError, "users not found")
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

	runMigrations(t, db, "../../migrations")
	insertTestData(t, db)

	return db
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

func insertTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec( `
        INSERT INTO users (username, email, full_name)
        VALUES
        ('tequila_sunset', 'harrier.dubois@rcm.org', 'Harrier Du Bois'),
        ('kim',   'kim.kitsuragi@rcm.com', 'Kim Kitsuragi');
    `)
	if err != nil {
		t.Fatalf("failed to insert data: %v", err)
	}
}

func assertThatUserNamesAreExpected(t *testing.T, user map[string]interface{}, expectedUserName string, expectedFullName string) {
	t.Helper()

	if user["username"] != expectedUserName || user["full_name"] != expectedFullName {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func assertThatErrorMessageIsExpected(t *testing.T, jsonError map[string]interface{}, expectedMessage string) {
	t.Helper()

	if jsonError["error"] != expectedMessage {
		t.Fatalf("unexpected error message: %+v", jsonError)
	}
}

