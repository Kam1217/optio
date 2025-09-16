package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Kam1217/optio/app"
	"github.com/Kam1217/optio/internal/auth/middleware"
	"github.com/google/uuid"
)

type SessionHandler struct {
	sessionService *app.SessionService
}

func NewSessionHandler(s *app.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: s}
}

type CreateSessionRequest struct {
	SessionName string `json:"session_name"`
}

type CreateSessionResponse struct {
	SessionID   uuid.UUID `json:"session_id"`
	SessionCode string    `json:"session_code"`
	SessionName string    `json:"session_name"`
	InviteLink  string    `json:"invite_link"`
}

func (sh *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	creatorID, ok := middleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Error(w, "Unathurised: Creator ID not found", http.StatusUnauthorized)
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.SessionName == "" {
		http.Error(w, "Session name is required", http.StatusBadRequest)
		return
	}

	session, inviteLink, err := sh.sessionService.CreateNewSession(r.Context(), req.SessionName, creatorID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := CreateSessionResponse{
		SessionID:   session.ID,
		SessionCode: session.SessionCode,
		SessionName: session.SessionName,
		InviteLink:  inviteLink,
	}

	sh.respondWithJSON(w, response, http.StatusCreated)
}

func (sh *SessionHandler) respondWithJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
