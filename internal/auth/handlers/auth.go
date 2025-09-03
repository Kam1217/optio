package handlers

import "github.com/Kam1217/optio/internal/auth/models"

//This is where the authentication logic goes -login, register etc.

//TODO: DATABASE - sits here

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email string `json:"email"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User models.User `json:"user"`
}

//Register func 
//- Check if user existsd
//- generate JWT
//-Register - create user 


//Login func

//Profile func 
//- retrievs the user profile