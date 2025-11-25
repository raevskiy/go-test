package integrationtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http/httptest"
	"testing"
	"time"
)

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
