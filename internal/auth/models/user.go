package models

import (
	"context"

	"github.com/Kam1217/optio/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	queries *database.Queries
}

func NewUserService(queries *database.Queries) *UserService {
	return &UserService{queries: queries}
}
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func CheckPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) UserExists(ctx context.Context, username, email string) (bool, error) {
	usernameExists, err := s.queries.UserExistsByUsername(ctx, username)
	if err != nil {
		return false, err
	}

	emailExists, err := s.queries.UserExistsByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	return usernameExists || emailExists, nil
}


//TODO:Check is user exists

//TODO:Create User - insert into database

//TODO:Get User by username

//TODO:Get User by ID
