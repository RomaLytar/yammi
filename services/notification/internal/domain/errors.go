package domain

import "errors"

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrEmptyUserID          = errors.New("user id is required")
	ErrEmptyTitle           = errors.New("title is required")
	ErrEmptyType            = errors.New("notification type is required")
)
