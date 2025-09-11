package models

import (
	"context"
	"time"

	"github.com/Kam1217/optio/internal/auth/middleware"
	"github.com/Kam1217/optio/internal/database"
	"github.com/google/uuid"
)

type RefreshService struct {
	queries *database.Queries
	ttl     time.Duration
}

func NewRefreshService(q *database.Queries, ttl time.Duration) *RefreshService {
	return &RefreshService{queries: q, ttl: ttl}
}

func (r *RefreshService) IssueRefreshToken(ctx context.Context, userID uuid.UUID, ua, ip string) (plain string, err error) {
	tokenHash, plain, err := middleware.MakeRefreshToken()
	if err != nil {
		return "", err
	}
	now := time.Now()
	_, err = r.queries.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		IssuedAt:  now,
		ExpiresAt: now.Add(r.ttl),
		UserAgent: ua,
		Ip:        ip,
	})
	return plain, nil
}
