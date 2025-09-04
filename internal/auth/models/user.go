package models

import (
	"context"

	"github.com/Kam1217/optio/internal/database"
	"github.com/google/uuid"
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

func (s *UserService) CreateUser(ctx context.Context, username, email, password string) (*database.User, error) {
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.queries.CreateUser(ctx, database.CreateUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*database.User, error) {
	user, err := s.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*database.User, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*database.User, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// TODO: Validate user credentials

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]database.ListUsersRow, error) {
	users, err := s.queries.ListUsers(ctx, database.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.queries.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: passwordHash,
	})
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.queries.DeleteUser(ctx, userID); err != nil{
		return err
	}
	return nil
}
