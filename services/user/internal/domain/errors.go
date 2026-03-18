package domain

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmptyName    = errors.New("name is required")
	ErrEmailExists  = errors.New("user with this email already exists")
)
