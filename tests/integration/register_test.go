package integration

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Kam1217/optio/db"
	"github.com/Kam1217/optio/internal/auth/handlers"
	"github.com/Kam1217/optio/internal/auth/middleware"
	"github.com/Kam1217/optio/internal/auth/models"
	"github.com/gorilla/mux"
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
func startTestServer(t *testing.T) (*httptest.Server, *db.DB) {
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

	secretJWT := os.Getenv("JWT_SECRET")
	if secretJWT == "" {
		secretJWT = "testsecret"
	}
	jwtMgr := middleware.NewJWTManager(secretJWT, "optio", "optio-api", 15*time.Minute)

	user := models.NewUserService(dbConn.Queries)
	auth := handlers.NewAuthHandler(dbConn.DB, user, jwtMgr)

	router := mux.NewRouter()
	router.HandleFunc("/api/auth/register", auth.RegisterUser).Methods("POST")

	server := httptest.NewUnstartedServer(router)
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	return server, dbConn
}

// Register - t.run success, duplicate email/username, missing(email, username, password), bad JSON, user exists

//helpers

type httpRes struct {
	Code int
	Body string
}

func postJSON(t *testing.T, url, body string) httpRes {
	return postRaw(t, url, body, "application/json")
}

func postRaw(t *testing.T, url, body, contentType string) httpRes {
	t.Helper()
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c := &http.Client{Timeout: 5 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return httpRes{Code: resp.StatusCode, Body: string(b)}
}

func mustJSON(t *testing.T, s string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(s), v); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, s)
	}
}
