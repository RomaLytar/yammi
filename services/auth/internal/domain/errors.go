package domain

import "errors"

var (
	ErrEmptyEmail      = errors.New("email is required")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrEmptyPassword   = errors.New("password is required")
	ErrWeakPassword    = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong = errors.New("password must be at most 72 characters")
	ErrEmptyName       = errors.New("name is required")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailExists     = errors.New("email already exists")
	ErrTokenNotFound   = errors.New("refresh token not found")
	ErrTokenRevoked    = errors.New("refresh token is revoked")
	ErrTokenExpired    = errors.New("refresh token has expired")
)
