package middleware

import (
	"testing"
	"time"

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
}
