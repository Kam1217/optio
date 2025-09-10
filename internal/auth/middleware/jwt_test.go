package middleware

import (
	"slices"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func newMgr() *JWTManager {
	return NewJWTManager("supersecret", "tester", "client", time.Minute*15)
}

func TestGenerateAndValidateJWT_Success(t *testing.T) {
	m := newMgr()
	uid := uuid.New()
	username := "testusername"

	token, err := m.GenerateJWT(uid, username)
	if err != nil {
		t.Fatalf("Generate JWT error: %v", err)
	}
	if token == "" {
		t.Fatalf("expected token")
	}

	claims, err := m.ValidateJWT(token)
	if err != nil {
		t.Fatalf("validate JWT error: %v", err)
	}
	if claims.UserID.String() != uid.String() {
		t.Fatalf("userID mismatch: got %v, want %v", claims.UserID, uid)
	}
	if claims.Username != username {
		t.Fatalf("username mismatch: got %v, want %v", claims.Username, username)
	}
	if claims.Issuer != "tester" {
		t.Fatalf("issuer mismatch: got %v, want tester", claims.Issuer)
	}
	found := slices.Contains(claims.Audience, "client")
	if !found {
		t.Fatalf("audience mismatch: got %v, want client", claims.Audience)
	}
	if claims.Subject != uid.String() {
		t.Fatalf("subject mismatch")
	}
}

func TestValidateJWT_fail(t *testing.T) {
	t.Run("wrong secret", func(t *testing.T) {
		uid := uuid.New()
		correct := newMgr()
		wrong := NewJWTManager("wrongsecret", "tester", "client", time.Minute*15)

		token, _ := correct.GenerateJWT(uid, "username")
		if _, err := wrong.ValidateJWT(token); err == nil {
			t.Fatalf("expected error with wrong secret")
		}
	})

	t.Run("wrong signing method", func(t *testing.T) {
		m := newMgr()
		uid := uuid.New()
		claims := &Claims{
			UserID:   uid,
			Username: "username",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    m.issuer,
				Subject:   uid.String(),
				Audience:  []string{m.audience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ID:        uuid.NewString(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signed, _ := token.SignedString([]byte("irrelevant"))
		if _, err := m.ValidateJWT(signed); err == nil {
			t.Fatalf("expected error for wrong signing method")
		}
	})
}
