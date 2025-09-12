package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

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
	plain, tokenHash := MakeRefreshToken()

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
	hash := hashRefresh(oldPlain)
	now := time.Now()
	_ = hashRefresh(oldPlain)
	refreshToken, err := r.queries.GetActiveRefreshTokenByHash(ctx, hash)
	if err != nil {
		return "", uuid.Nil, err
	}

	if userPasswordChangedAt != nil && userPasswordChangedAt.After(refreshToken.IssuedAt) {
		_ = r.queries.RevokeRefreshTokenByID(ctx, refreshToken.ID)
		return "", uuid.Nil, fmt.Errorf("invalid refresh token")
	}

	_ = r.queries.RevokeRefreshTokenByID(ctx, refreshToken.ID)
	newPlain, newHash := MakeRefreshToken()

	_, err = r.queries.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		UserID:    refreshToken.UserID,
		TokenHash: newHash,
		IssuedAt:  now,
		ExpiresAt: now.Add(r.ttl),
		UserAgent: ua,
		Ip:        ip,
	})
	if err != nil {
		return "", uuid.Nil, err
	}
	return newPlain, refreshToken.UserID, nil
}

func MakeRefreshToken() (plain, hash string) {
	token := make([]byte, 32)
	rand.Read(token)
	plain = base64.RawURLEncoding.EncodeToString(token)

	sum := sha256.Sum256([]byte(plain))
	hash = base64.RawURLEncoding.EncodeToString(sum[:])
	return
}

func hashRefresh(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
