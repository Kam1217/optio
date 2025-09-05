package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Kam1217/optio/internal/auth/middleware"
	"github.com/Kam1217/optio/internal/auth/models"
	"github.com/Kam1217/optio/internal/database"
	"github.com/google/uuid"
)

type AuthHandler struct {
	DB          *sql.DB
	UserService *models.UserService
	JWT         *middleware.JWTManager
}

func NewAuthHandler(db *sql.DB, userService *models.UserService, jwtManager *middleware.JWTManager) *AuthHandler {
	return &AuthHandler{
		DB:          db,
		UserService: userService,
		JWT:         jwtManager,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

func (h *AuthHandler) toUserResponseFromCreate(u *database.CreateUserRow) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}


func (h *AuthHandler) toUserFromLogin(user *database.GetUserForLoginRow) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

func (h *AuthHandler) toUserGetUserByIDRow(user *database.GetUserByIDRow) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "Username, password and email cannot be empty", http.StatusBadRequest)
		return
	}

	exists, err := h.UserService.UserExists(ctx, req.Username, req.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "User with this email or username already exists", http.StatusConflict)
		return
	}

	user, err := h.UserService.CreateUser(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	token, err := h.JWT.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}
	response := AuthResponse{
		Token: token,
		User:  h.toUserResponseFromCreate(user),
	}

	h.respondWithJSON(w, response, http.StatusOK)
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Identifier == "" || req.Password == "" {
		http.Error(w, "identifier and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.ValidateUserCredentials(ctx, req.Identifier, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentails) || errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	token, err := h.JWT.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: token,
		User:  h.toUserFromLogin(user),
	}

	h.respondWithJSON(w, response, http.StatusOK)
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromCtx(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	h.respondWithJSON(w, h.toUserGetUserByIDRow(user), http.StatusOK)
}

func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
