package integration

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/Kam1217/optio/db"
)

// helper funtion - migrate up/ down using goose
const migrationDir = "./sql/schema"

func gooseUp(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("goose", "-dir", dir, "up")
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goose up: %v\n%s", err, out)
	}
}

func gooseDown(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("goose", "-dir", dir, "down-to", "0")
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goose down: %v\n%s", err, out)
	}
}

// Start DB
func startTestServer(t *testing.T) (*http.Server, *db.DB) {
	t.Helper()

	gooseDown(t, migrationDir)
	gooseUp(t, migrationDir)

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "optio_test"
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "postgres"
	}
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	timezone := os.Getenv("DB_TIMEZONE")
	if timezone == "" {
		timezone = "UTC"
	}

	cfg := db.Config{
		DBName:   dbName,
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPass,
		SSLMode:  sslMode,
		TimeZone: timezone,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	dbConn, err := db.Connect(ctx, cfg)
	if err != nil {
		t.Fatalf("db connection: %v", err)
	}
	t.Cleanup(func() { _ = dbConn.Close() })

	t.Cleanup(func() {
		_, _ = dbConn.DB.ExecContext(context.Background(), `TRUNCATE TABLE users RESTART IDENTITY`)
	})

}

//Register - t.run success, duplicate email/username, missing(email, username, password), bad JSON, user exists
