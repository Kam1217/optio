package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Kam1217/optio/internal/auth/middleware"
	"github.com/Kam1217/optio/internal/auth/models"
	"github.com/Kam1217/optio/internal/database"
	"github.com/google/uuid"
)

// This is where the authentication logic goes -login, register etc.
type AuthHandler struct {
	DB          *sql.DB
	UserService *models.UserService
}

func NewAuthHandler(db *sql.DB, userService *models.UserService) *AuthHandler {
	return &AuthHandler{
		DB:          db,
		UserService: userService,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *AuthHandler) toUserResponse(user *database.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "Username, password and email cannot be empty", http.StatusBadRequest)
		return
	}

	exists, err := h.UserService.UserExists(context.Background(), req.Username, req.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "user with this email or username already exists", http.StatusConflict)
		return
	}

	user, err := h.UserService.CreateUser(context.Background(), req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	token, err := middleware.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	userResponse := h.toUserResponse(user)
	response := AuthResponse{
		Token: token,
		User:  userResponse,
	}

	h.respondWithJSON(w, response, http.StatusOK)
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.ValidateUserCredentials(context.Background(), req.Username, req.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
	}

	token, err := middleware.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	userResponse := h.toUserResponse(user)
	response := AuthResponse{
		Token: token,
		User:  userResponse,
	}

	h.respondWithJSON(w, response, http.StatusOK)
}

//Profile func
//- retrievs the user profile

func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
