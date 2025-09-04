package db

import (
	"database/sql"
	"fmt"

	"github.com/Kam1217/optio/internal/database"
)

//Setup DB here

type DB struct {
	*sql.DB
	dbQueries *database.Queries
}

type Config struct {
	DBName   string
	Host     string
	Port     string
	User     string
	Password string
}

func Connect(cfg Config) (*DB, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&TimeZone=UTC", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}
	if err := dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := database.New(dbConn)

	return &DB{
		DB:        dbConn,
		dbQueries: queries,
	}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

