package middleware

import (
	"net/http"
	"net/http/httptest"
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

	t.Run("expired token", func(t *testing.T) {
		m := NewJWTManager("supersecret", "tester", "client", -time.Hour)
		uid := uuid.New()
		token, _ := m.GenerateJWT(uid, "username")
		if _, err := m.ValidateJWT(token); err == nil {
			t.Fatalf("expected expired error")
		}
	})

	t.Run("not valid yet", func(t *testing.T) {
		m := newMgr()
		uid := uuid.New()
		now := time.Now()
		claims := &Claims{
			UserID:   uid,
			Username: "username",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    m.issuer,
				Subject:   uid.String(),
				Audience:  []string{m.audience},
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				NotBefore: jwt.NewNumericDate(now.Add(10 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(now),
				ID:        uuid.NewString(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(m.secret)
		if err != nil {
			t.Fatalf("sign: %v", err)
		}

		if _, err := m.ValidateJWT(signed); err == nil {
			t.Fatalf("expected not valid yet error")
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		m := newMgr()
		if _, err := m.ValidateJWT("not a JWT"); err == nil {
			t.Fatalf("expected parse error")
		}
	})
}

func TestJwtMiddleware(t *testing.T) {
	m := newMgr()
	uid := uuid.New()
	username := "username"
	token, err := m.GenerateJWT(uid, username)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	var gotID uuid.UUID
	var gotUser string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := UserIDFromCtx(r.Context()); ok {
			gotID = id
		}
		if u, ok := UsernameFromCtx(r.Context()); ok {
			gotUser = u
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := m.JWTMiddleware(next)

	t.Run("success with bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d want 200", w.Code)
		}
		if gotID != uid || gotUser != username {
			t.Fatalf("context values not set: id=%v user=%q", gotID, gotUser)
		}
	})

	t.Run("missing auth header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %v", w.Code)
		}
	})

	t.Run("invalid scheme", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "invalid ")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %v", w.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", " Bearer invalid ")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %v", w.Code)
		}
	})

	t.Run("case-insensitive bearer and extra spaces handled", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "bEaReR   "+token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d want 200", w.Code)
		}
		if gotID != uid || gotUser != username {
			t.Fatalf("context values not set after mixed case bearer: id=%v user=%q", gotID, gotUser)
		}
	})
}
