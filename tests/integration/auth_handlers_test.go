package integration

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../../.env")
	os.Exit(m.Run())
}

type postgresContainer struct {
	testcontainers.Container
	DbUrl    string
	DbName   string
	Password string
	Port     string
}

func startPostgresContainer(ctx context.Context) (*postgresContainer, error) {
	dbName := "optio"
	port := "5432/tcp"
	password := "example"
	var dbUrl string
	dbURLFunc := func(host string, port nat.Port) string {
		url := fmt.Sprintf("postgres://postgres:%s@%s:%s/%s?sslmode=disable", password, host, port.Port(), dbName)
		dbUrl = url
		return url
	}
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{port},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_PASSWORD": password,
		},
		WaitingFor: wait.ForSQL(nat.Port(port), "postgres", dbURLFunc),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return nil, err
	}

	return &postgresContainer{
		Container: container,
		DbName:    dbName,
		DbUrl:     dbUrl,
		Port:      mappedPort.Port(),
		Password:  password,
	}, nil
}

const migrationDir = "../../sql/schema"

func gooseUp(t *testing.T, dir, dbUrl string) {
	t.Helper()
	cmd := exec.Command("goose", "-dir", dir, "postgres", dbUrl, "up")
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goose up: %v\n%s", err, out)
	}
}

// func gooseDown(t *testing.T, dir, dbUrl string) {
// 	t.Helper()
// 	cmd := exec.Command("goose", "-dir", dir, "postgres", dbUrl, "down-to", "0")
// 	cmd.Env = os.Environ()
// 	out, err := cmd.CombinedOutput()
// 	if err != nil {
// 		t.Fatalf("goose down: %v\n%s", err, out)
// 	}
// }

func startTestServer(t *testing.T, dbContainer *postgresContainer) (*httptest.Server, *db.DB) {
	t.Helper()

	// gooseDown(t, migrationDir, c.DbUrl)
	gooseUp(t, migrationDir, dbContainer.DbUrl)

	// dbName := os.Getenv("DB_NAME")
	// if dbName == "" {
	// 	dbName = "optio_test"
	// }
	// dbHost := os.Getenv("DB_HOST")
	// if dbHost == "" {
	// 	dbHost = "localhost"
	// }
	// dbPort := os.Getenv("DB_PORT")
	// if dbPort == "" {
	// 	dbPort = "5432"
	// }
	// dbUser := os.Getenv("DB_USER")
	// if dbUser == "" {
	// 	dbUser = "postgres"
	// }
	// dbPass := os.Getenv("DB_PASSWORD")
	// if dbPass == "" {
	// 	dbPass = "postgres"
	// }
	// sslMode := os.Getenv("DB_SSLMODE")
	// if sslMode == "" {
	// 	sslMode = "disable"
	// }

	// timezone := os.Getenv("DB_TIMEZONE")
	// if timezone == "" {
	// 	timezone = "UTC"
	// }

	cfg := db.Config{
		Host:     "localhost",
		Password: dbContainer.Password,
		Port:     dbContainer.Port,
		DBName:   dbContainer.DbName,
		// DBName:   dbName,
		// Host:     dbHost,
		// Port:     dbPort,
		User: "postgres",
		// Password: dbPass,
		// SSLMode:  sslMode,
		// TimeZone: timezone,
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
	refreshTTL := 30 * 24 * time.Hour
	refreshSvc := models.NewRefreshService(dbConn.Queries, refreshTTL)
	auth.Refresh = refreshSvc
	auth.RefreshTTL = refreshTTL
	auth.CookieDomain = ""

	router := mux.NewRouter()
	router.HandleFunc("/api/auth/register", auth.RegisterUser).Methods("POST")
	router.HandleFunc("/api/auth/login", auth.LoginUser).Methods("POST")

	server := httptest.NewUnstartedServer(router)
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	return server, dbConn
}

func TestRegister(t *testing.T) {
	dbContainer, err := startPostgresContainer(context.Background())
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer testcontainers.CleanupContainer(t, dbContainer)

	server, _ := startTestServer(t, dbContainer)
	base := server.URL

	res := postJSON(t, base+"/api/auth/register", `{"username":"test1", "email":"test1@example.com", "password":"test123"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("succesful register: want 200, got %d body:%s", res.Code, res.Body)
	}
	var ok handlers.AuthResponse
	mustJSON(t, res.Body, &ok)
	if ok.Token == "" || ok.User.Username != "test1" || ok.User.Email != "test1@example.com" {
		t.Fatalf("bad response: %+v", ok)
	}
	res = postJSON(t, base+"/api/auth/register", `{"username":"test2", "email":"test1@example.com", "password":"test123"}`)
	if res.Code != http.StatusConflict {
		t.Fatalf("duplicate email: want 409, got %d body:%s", res.Code, res.Body)
	}
	res = postJSON(t, base+"/api/auth/register", `{"username":"test1", "email":"test2@example.com", "password":"test123"}`)
	if res.Code != http.StatusConflict {
		t.Fatalf("duplicate username: want 409, got %d body:%s", res.Code, res.Body)
	}
	res = postJSON(t, base+"/api/auth/register", `{"username":"", "email":"", "password":""}`)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("missing field: want 400, got %d body:%s", res.Code, res.Body)
	}

	res = postRaw(t, base+"/api/auth/register", `{"username":"x","email":"x@example.com","password":"y"`, "application/json")
	if res.Code != http.StatusBadRequest {
		t.Fatalf("invalid JSON: want 400, got %d body=%s", res.Code, res.Body)
	}
}

func TestLogin(t *testing.T) {
	dbContainer, err := startPostgresContainer(context.Background())
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer testcontainers.CleanupContainer(t, dbContainer)

	server, _ := startTestServer(t, dbContainer)
	base := server.URL

	regBody := `{"username":"test3", "email":"test3@example.com","password":"test123"}`
	regRes := postJSON(t, base+"/api/auth/register", regBody)
	if regRes.Code != 200 {
		t.Fatalf("register: want 200, got %d body:%s", regRes.Code, regRes.Body)
	}

	loginBody := `{"identifier":"test3","password":"test123"}`
	loginRes := postJSON(t, base+"/api/auth/login", loginBody)
	if loginRes.Code != 200 {
		t.Fatalf("succesfull username login: want 200, got %d body: %s", loginRes.Code, loginRes.Body)
	}

	var login handlers.AuthResponse
	mustJSON(t, loginRes.Body, &login)
	if login.Token == "" || login.User.Username != "test3" {
		t.Fatalf("bad login response: %v", login)
	}

	loginBody2 := `{"identifier":"test3@example.com", "password":"test123"}`
	loginRes2 := postJSON(t, base+"/api/auth/login", loginBody2)
	if loginRes2.Code != http.StatusOK {
		t.Fatalf("succesfull username login: want 200, got %d body: %s", loginRes2.Code, loginRes2.Body)
	}

	badIdentifierBody := `{"identifier":"wrong", "password":"test123"}`
	badIdentifierRes := postJSON(t, base+"/api/auth/login", badIdentifierBody)
	if badIdentifierRes.Code != http.StatusUnauthorized {
		t.Fatalf("wrong username: want 401, got %d body:%s", badIdentifierRes.Code, badIdentifierRes.Body)
	}

	badPasswordBody := `{"identifier": "test3", "password":"wrong"}`
	badPasswordRes := postJSON(t, base+"/api/auth/login", badPasswordBody)
	if badPasswordRes.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password: want 401, got %d body:%s", badPasswordRes.Code, badPasswordRes.Body)
	}

	missingFieldBody := `{"identifier":"", "password":""}`
	missingFieldRes := postJSON(t, base+"/api/auth/login", missingFieldBody)
	if missingFieldRes.Code != http.StatusBadRequest {
		t.Fatalf("missing fields: want 400, got %d body:%s", missingFieldRes.Code, missingFieldRes.Body)
	}
}

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
