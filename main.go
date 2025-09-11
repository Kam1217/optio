package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Kam1217/optio/db"
	"github.com/Kam1217/optio/internal/auth/handlers"
	"github.com/Kam1217/optio/internal/auth/middleware"
	"github.com/Kam1217/optio/internal/auth/models"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	jwtMgr := middleware.NewJWTManager(jwtSecret, "optio", "optio-api", 15*time.Minute)

	dbConfig := db.Config{
		DBName:          os.Getenv("DB_NAME"),
		Host:            os.Getenv("DB_HOST"),
		Port:            os.Getenv("DB_PORT"),
		User:            os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		SSLMode:         os.Getenv("DB_SSLMODE"),
		TimeZone:        os.Getenv("DB_TIMEZONE"),
		MaxOpenConns:    atoiEnv("DB_MAXOPENCONNS", 20),
		MaxIdleConns:    atoiEnv("DB_MAXIDLECONNS", 10),
		ConnMaxLifetime: durEnv("DB_CONNMAXLIFETIME", "1h"),
		ConnMaxIdleTime: durEnv("DB_CONNMAXIDLETIME", "10m"),
	}

	dbConn, err := db.Connect(context.Background(), dbConfig)
	if err != nil {
		log.Fatalf("DB connect: %vv", err)
	}
	defer dbConn.Close()
	log.Printf("Succesfully connected to the database")

	userService := models.NewUserService(dbConn.Queries)
	authHandler := handlers.NewAuthHandler(dbConn.DB, userService, jwtMgr)

	router := setUpRouts(authHandler, jwtMgr)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Listen and serve: %v", err)
	}
}

func setUpRouts(authHandler *handlers.AuthHandler, jwtMgr *middleware.JWTManager) *mux.Router {
	router := mux.NewRouter()
	router.Use(corsMiddleware)

	router.HandleFunc("/api/auth/register", authHandler.RegisterUser).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.LoginUser).Methods("POST")
	router.Handle("/api/auth/profile", jwtMgr.JWTMiddleware(http.HandlerFunc(authHandler.Profile))).Methods("GET")
	router.HandleFunc("/api/auth/refresh", authHandler.RefreshSession).Methods("POST")
	router.HandleFunc("/api/auth/logout", authHandler.Logout).Methods("POST")

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "Server is running"}`))
	}).Methods("GET")

	fs := http.FileServer(http.Dir("./assets"))
	router.PathPrefix("/").Handler(fs)

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Endpoint not found", http.StatusNotFound)
	})

	return router
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func atoiEnv(key string, def int) int {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return valueInt
}

func durEnv(key, _ string) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(0)
	}
	valueTime, err := time.ParseDuration(value)
	if err != nil {
		return time.Duration(0)
	}
	return valueTime
}
