package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

var ErrInvalidCredentails = errors.New("invalid credentials")

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hashedPassword), nil
}

func checkPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func (s *UserService) UserExists(ctx context.Context, username, email string) (bool, error) {
	exists, err := s.queries.UserExistsByUsernameOrEmail(ctx, database.UserExistsByUsernameOrEmailParams{
		Username: username,
		Email:    email,
	})
	if err != nil {
		return false, fmt.Errorf("users exists: %w", err)
	}

	return exists, nil
}

func (s *UserService) CreateUser(ctx context.Context, username, email, password string) (*database.CreateUserRow, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	user, err := s.queries.CreateUser(ctx, database.CreateUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*database.GetUserByIDRow, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*database.GetUserByUsernameRow, error) {
	user, err := s.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*database.GetUserByEmailRow, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

func (s *UserService) ValidateUserCredentials(ctx context.Context, identifier, password string) (*database.GetUserForLoginRow, error) {
	user, err := s.queries.GetUserForLogin(ctx, identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentails
		}
		return nil, fmt.Errorf("get user for login: %w", err)
	}
	if !checkPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentails
	}

	return &user, nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]database.ListUsersRow, error) {
	users, err := s.queries.ListUsers(ctx, database.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (s *UserService) UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.queries.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: passwordHash,
	}); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	return nil
}

func (s *UserService) UpdateUsername(ctx context.Context, userID uuid.UUID, username string) error {
	if err := s.queries.UpdateUsername(ctx, database.UpdateUsernameParams{
		ID:       userID,
		Username: username,
	}); err != nil {
		return fmt.Errorf("update username: %w", err)
	}

	return nil
}

func (s *UserService) UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error {
	if err := s.queries.UpdateEmail(ctx, database.UpdateEmailParams{
		ID:    userID,
		Email: email,
	}); err != nil {
		return fmt.Errorf("update email: %w", err)
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.queries.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
