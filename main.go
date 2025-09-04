package main

import (
	"log"
	"os"

	"github.com/Kam1217/optio/db"
	_ "github.com/lib/pq"
)

func main() {
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
}

//Setup mux ports helper func
