package models

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
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
	if err != nil {
		return "", err
	}
	return plain, nil
}

func (r *RefreshService) RotateRefreshToken(ctx context.Context, oldPlain string, userPasswordChangedAt *time.Time, ua, ip string) (newPlain string, userID uuid.UUID, err error) {

}

func hashRefresh(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
