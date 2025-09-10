package models

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	t.Run("succeeds and verifies", func(t *testing.T) {
		hash, err := hashPassword("test123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hash == "" {
			t.Fatalf("expected a non-empty hash")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("test123")); err != nil {
			t.Fatalf("hashed password did not verify: %v", err)
		}
	})

	t.Run("different hashes from same input", func(t *testing.T) {
		hash1, err := hashPassword("same-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hash2, err := hashPassword("same-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hash1 == hash2 {
			t.Fatalf("expected different hashes but got the same")
		}
	})
}
