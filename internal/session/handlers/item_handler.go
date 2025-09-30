package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Kam1217/optio/app"
	"github.com/Kam1217/optio/internal/auth/middleware"
	"github.com/google/uuid"
)

type ItemHandler struct {
	itemService *app.SessionItemService
}

func NewItemHandler(i *app.SessionItemService) *ItemHandler {
	return &ItemHandler{itemService: i}
}

type CreateItemRequest struct {
	ItemInput app.SessionItemInput
}

type CreateItemResponse struct {
	ItemID          uuid.UUID `json:"item_id"`
	SessionID       uuid.UUID `json:"session_id"`
	ItemTitle       string    `json:"item_title"`
	ItemDescription string    `json:"item_description"`
	ImageURL        string    `json:"image_url"`
	SourceType      string    `json:"source_type"`
}

func (ih *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Error(w, "Unauthrised: Creator ID not found", http.StatusUnauthorized)
		return
	}

	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.ItemInput.Title == "" {
		http.Error(w, "Item title required", http.StatusBadRequest)
		return
	}

	item, err := ih.itemService.CreateNewSessionItem(r.Context(), req.ItemInput)
	if err != nil {
		http.Error(w, "Failed to create item", http.StatusInternalServerError)
		return
	}

	response := CreateItemResponse{
		ItemID:          item.ID,
		SessionID:       item.SessionID,
		ItemTitle:       item.ItemTitle,
		ItemDescription: item.ItemDescription.String,
		ImageURL:        item.ImageUrl.String,
		SourceType:      item.SourceType,
	}
	ih.respondWithJSON(w, response, http.StatusCreated)
}

func (ih *ItemHandler) respondWithJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
