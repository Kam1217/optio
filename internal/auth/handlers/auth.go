package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Kam1217/optio/internal/auth/models"
	"github.com/Kam1217/optio/internal/database"
)

//This is where the authentication logic goes -login, register etc.

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
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// Register func
// -decode body to struct
// -Handle empty username, password, email
// - Check if user exists
// - generate JWT
// -Register - create user
func (a *AuthResponse) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "Username, password and email cannot be empty", http.StatusBadRequest)
		return
	}

	//Check if user exists - db func

	//Create user - db func

	//Generate JWT - need db

}

//Login func

//Profile func
//- retrievs the user profile
