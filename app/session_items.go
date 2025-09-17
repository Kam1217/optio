package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Kam1217/optio/internal/database"
	"github.com/google/uuid"
)

//Check if item already exists in session (by ID) - prevents duplicates
//Create session item that can be added to sessions

type SessionItemService struct {
	queries *database.Queries
}

func NewSessionItemService(queries *database.Queries) *SessionItemService {
	return &SessionItemService{queries: queries}
}

func (si *SessionItemService) CheckSessionItemExists(ctx context.Context, itemID uuid.UUID) (bool, error) {
	_, err := si.queries.GetSessionItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("database error checking session item: %w", err)
	}
	return true, nil
}

func (si *SessionItemService) CreateNewSessionItem(ctx context.Context, sessionID uuid.UUID, itemName, itemDescription, itemImaege string, creatorID uuid.UUID) (*database.SessionItem, error) {
	return nil, nil
}
