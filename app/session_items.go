package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Kam1217/optio/internal/database"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type SessionItemService struct {
	queries *database.Queries
}

type SourceType string

const (
	SourceCustom SourceType = "custom"
	SourceSteam  SourceType = "steam"
)

//ADD STEAM LATER

type SessionItemInput struct {
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	ImageURL      string          `json:"image_url"`
	SessionId     uuid.UUID       `json:"session_id"`
	AddedByUserID uuid.UUID       `json:"added_by_user_id"`
	SourceType    SourceType      `json:"source_type"`
	Metadata      json.RawMessage `json:"metadata"`
}

// type CustomMetadata struct {
// 	Foo string
// 	Bar string
// }

// type SteamMetadata struct {
// 	Baz int
// 	Yeet float32
// }

// func GetMetadata[T any](rawMetadata json.RawMessage) (*T, error) {
// 	var metadata *T
// 	err := json.Unmarshal(rawMetadata, metadata)
// 	return metadata, err
// }

// func example(sourceType SourceType, rawMetadata pqtype.NullRawMessage) {
// 	switch sourceType {
// 	case SourceCustom:
// 		customMetadata, err := GetMetadata[CustomMetadata](rawMetadata)
// 	case SourceSteam:
// 		steamMetadata, err := GetMetadata[SteamMetadata](rawMetadata)
// 	}
// }

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

func (si *SessionItemService) CreateNewSessionItem(ctx context.Context, itemInput SessionItemInput) (*database.SessionItem, error) {
	if itemInput.SessionId == uuid.Nil || itemInput.AddedByUserID == uuid.Nil {
		return nil, fmt.Errorf("missing ID")
	}
	//Make switch statement once steam / other is added - //DEFAULT UNSUPORTED SOURCE TYPE

	if itemInput.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	exists, err := si.CheckSessionItemExists(ctx, itemInput.SessionId)
	if err != nil {
		return nil, fmt.Errorf("error verifying if code exists: %w", err)
	}
	if !exists {
		item, err := si.queries.CreateSessionItem(ctx, database.CreateSessionItemParams{
			SessionID:       itemInput.SessionId,
			ItemTitle:       itemInput.Title,
			ItemDescription: sql.NullString{String: itemInput.Description, Valid: true},
			ImageUrl:        sql.NullString{String: itemInput.ImageURL, Valid: true},
			SourceType:      string(SourceCustom),
			SourceID:        sql.NullString{Valid: false},
			Metadata:        pqtype.NullRawMessage{RawMessage: itemInput.Metadata, Valid: true},
			AddedByUserID:   itemInput.AddedByUserID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create session item: %w", err)
		}
		return &item, nil
	}
	return nil, fmt.Errorf("item already exists")
}
