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

func startPostgresContainer(t *testing.T) (testcontainers.Container, string) {
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

	dsn := fmt.Sprintf(
		"postgres://postgres:postgres@%s:%s/testdb?sslmode=disable",
		host,
		port.Port(),
	)

	return postgresContainer, dsn
}

func runMigrations(t *testing.T, db *sql.DB, dsn string, migrationsDir string) {
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

func TestGetAllUsers_Integration(t *testing.T) {
	//Arrange
	gin.SetMode(gin.TestMode)

	container, dsn := startPostgresContainer(t)
	defer container.Terminate(context.Background())
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	runMigrations(t, db, dsn, "../../migrations")

	insertTestData(t, db)
	router := handler.SetupAppLayersAndRouter(db)

	//Act
	req, _ := http.NewRequest("GET", "/api/v1/users/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	//Assert
	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(users) != 3 + 2 {
		t.Fatalf("expected 3 users from migration and 2 users added in this test, got %d", len(users))
	}
	expectoUser(t, users[0], "asmith", "Alice Smith")
	expectoUser(t, users[4], "kim", "Kim Kitsuragi")
}

func expectoUser(t *testing.T, user map[string]interface{}, expectedUserName string, expectedFullName string) {
	if user["username"] != expectedUserName || user["full_name"] != expectedFullName {
		t.Fatalf("unexpected user: %+v", user)
	}
}
