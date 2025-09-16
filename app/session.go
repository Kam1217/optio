package app

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/Kam1217/optio/internal/database"
	"github.com/google/uuid"
)

type SessionService struct {
	queries   *database.Queries
	InviteURL string
}

func NewSessionService(queries *database.Queries, inviteURL string) *SessionService {
	return &SessionService{queries: queries, InviteURL: inviteURL}
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

func (s *SessionService) generateInviteLink(sessionCode string) (string, error) {
	if s.InviteURL == "" {
		return "", fmt.Errorf("invite URL is not configured")
	}
	link, err := url.Parse(s.InviteURL)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}

	q := link.Query()
	q.Set("code", sessionCode)
	link.RawQuery = q.Encode()

	return link.String(), nil
}

func (s *SessionService) CreateNewSession(ctx context.Context, sessionName string, creatorID uuid.UUID) (*database.Session, string, error) {
	sessionCode, err := s.generateUniqueSessionCode(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("error generating session code: %w", err)
	}

	session, err := s.queries.CreateSession(ctx, database.CreateSessionParams{
		SessionCode:   sessionCode,
		SessionName:   sessionName,
		CreatorUserID: creatorID,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session in db: %w", err)
	}

	inviteLink, err := s.generateInviteLink(sessionCode)
	if err != nil {
		return nil, "", fmt.Errorf("error generating invite link: %w", err)
	}

	return &session, inviteLink, nil
}
