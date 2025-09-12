package middleware

import (
	"context"

	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	secret    []byte
	issuer    string
	audience  string
	expiresIn time.Duration
}

func NewJWTManager(secret, issuer, audience string, expiresIn time.Duration) *JWTManager {
	return &JWTManager{
		secret:    []byte(secret),
		issuer:    issuer,
		audience:  audience,
		expiresIn: expiresIn,
	}
}

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func (m *JWTManager) GenerateJWT(userID uuid.UUID, username string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			Audience:  []string{m.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(m.secret)
}

func (m *JWTManager) ValidateJWT(tokenstring string) (*Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenstring, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrTokenUnverifiable
		}

		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	now := time.Now()
	leeway := time.Minute

	if claims.ExpiresAt != nil && now.After(claims.ExpiresAt.Time.Add(leeway)) {
		return nil, errors.New("token expired")
	}

	if claims.NotBefore != nil && now.Before(claims.NotBefore.Time.Add(-leeway)) {
		return nil, errors.New("token not valid yet")
	}

	return &claims, nil
}

type ctxKey int

const (
	ctxUserIDKey ctxKey = iota
	ctxUsernameKey
)

func UserIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ctxUserIDKey)
	id, ok := v.(uuid.UUID)

	return id, ok
}

func UsernameFromCtx(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxUsernameKey)
	s, ok := v.(string)
	return s, ok
}

func (m *JWTManager) JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}
		token := strings.TrimSpace(auth[len("Bearer "):])
		claims, err := m.ValidateJWT(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, ctxUsernameKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
