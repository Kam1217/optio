package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Kam1217/optio/internal/database"
)

type DB struct {
	*sql.DB
	Queries *database.Queries
}

type Config struct {
	DBName          string
	Host            string
	Port            string
	User            string
	Password        string
	SSLMode         string
	TimeZone        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
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
		DB:      dbConn,
		Queries: queries,
	}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Health() error {
	return db.DB.Ping()
}
