package models

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	t.Run("succeeds and verifies", func(t *testing.T) {
		hash, err := hashPassword("successful-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if hash == "" {
			t.Fatalf("expected a non-empty hash")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("successful-password")); err != nil {
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

func TestCheckPasswor(t *testing.T) {
	validHash, err := hashPassword("correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name   string
		hash   string
		pw     string
		wantOK bool
	}{
		{
			name:   "correct password",
			hash:   validHash,
			pw:     "correct-password",
			wantOK: true,
		},
		{
			name:   "wrong password",
			hash:   validHash,
			pw:     "wrong-password",
			wantOK: false,
		},
		{
			name:   "malformed hash returns false",
			hash:   "not-a-bcrypt-hash",
			pw:     "correct-password",
			wantOK: false,
		},
		{
			name:   "empty hash returns false",
			hash:   "",
			pw:     "correct-password",
			wantOK: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got := checkPassword(test.hash, test.pw)
			if got != test.wantOK {
				t.Fatalf("checkPassword(%q, %q) = %v, want %v", test.hash, test.pw, got, test.wantOK)
			}
		})
	}
}
