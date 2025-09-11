package models

import (
	"time"

	"github.com/Kam1217/optio/internal/database"
)

type RefreshService struct {
	queries *database.Queries
	ttl     time.Duration
}

func NewRefreshService(q *database.Queries, ttl time.Duration) *RefreshService {
	return &RefreshService{queries: q, ttl: ttl}
}

