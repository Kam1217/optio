package db

import (
	"database/sql"

	"github.com/Kam1217/optio/internal/database"
)

//Setup DB here

type DB struct {
	*sql.DB
	db *database.Queries
}

type Config struct {
	DBName   string
	Host     string
	Port     string
	User     string
	Password string
}

