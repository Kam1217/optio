package app

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Kam1217/optio/internal/database"
)

type SessionService struct {
	queries *database.Queries
}

func NewSessionService(queries *database.Queries) *SessionService {
	return &SessionService{queries: queries}
}

func (s *SessionService) CheckSessionCodeExists(ctx context.Context, code string) (bool, error) {
	_, err := s.queries.GetActiveSessionByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("database error checking session code: %w", err)
	}

	return true, nil
}

func (s *SessionService) generateUniqueSessionCode(ctx context.Context) (string, error) {
	for {
		code := rand.Text()
		exists, err := s.CheckSessionCodeExists(ctx, code)
		if err != nil {
			return "", fmt.Errorf("error verifying if code exists: %w", err)
		}
		if !exists {
			return code, nil
		}
	}
}

//Function that generates a session code - use crypto/rand to make secure (no one brute forcing into session)
//Must be unique
//Create invite link with code to send to users
// Extract code from invite link to enter session

//Func:

//Generate random code
//Check if already exists - if does retry - (GetSessionByCode)
