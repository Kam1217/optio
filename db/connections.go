package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
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

func Connect(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = "UTC"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s", url.QueryEscape(cfg.User), url.QueryEscape(cfg.Password), cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode, cfg.TimeZone)

	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		dbConn.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		dbConn.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		dbConn.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		dbConn.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	if err := dbConn.PingContext(ctx); err != nil {
		_ = dbConn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &DB{
		DB:      dbConn,
		Queries: database.New(dbConn),
	}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Health(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}
