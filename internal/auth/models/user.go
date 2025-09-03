package models

import (
	"time"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ToDo - User Model with Database
type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"_"`
	Email    string    `json:"email"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

//TODO:Check is user exists

//TODO:Create User - insert into database

//TODO:Get User by username

//TODO:Get User by ID


