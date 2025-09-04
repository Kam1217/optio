package main

import (
	"log"
	"net/http"
	"os"

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

	dbConfig := db.Config{
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	db, err := db.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Printf("Succesfully connected to the database")

	userService := models.NewUserService(db.Queries)

	authHandler := handlers.NewAuthHandler(db.DB, userService)

	router := setUpRouts(authHandler)

	port := os.Getenv("PORT")

	log.Printf("Server starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  POST /api/auth/register - Register new user")
	log.Printf("  POST /api/auth/login    - Login user")
	log.Printf("  GET  /api/auth/profile  - Get user profile (protected)")
	log.Printf("  GET  /api/users         - List users (protected)")
	log.Printf("  GET  /health            - Health check")

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func setUpRouts(authHandler *handlers.AuthHandler) *mux.Router {
	router := mux.NewRouter()
	router.Use(corsMiddleware)

	router.HandleFunc("/api/auth/register", authHandler.RegisterUser).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.LoginUser).Methods("POST")

	router.HandleFunc("/api/auth/profile", middleware.JWTMiddleware(authHandler.Profile)).Methods("GET")

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "Server is running"}`))
	}).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Endpoint not found"}`))
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

//REFRESH TOKEN
//SEND EMAIL TO VERIFY
//BE ABLE TO LOGIN VIA EMAIL OR USERNAME
//CLEAN UP
//GOOGLE LOGIN
//STEAM LOGIN