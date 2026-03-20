package domain

import "errors"

// Board errors
var (
	ErrBoardNotFound    = errors.New("board not found")
	ErrEmptyTitle       = errors.New("board title cannot be empty")
	ErrEmptyOwnerID     = errors.New("owner ID cannot be empty")
	ErrAccessDenied     = errors.New("access denied")
	ErrNotOwner         = errors.New("only owner can perform this action")
	ErrInvalidVersion   = errors.New("invalid version for optimistic locking")
)

// Column errors
var (
	ErrColumnNotFound = errors.New("column not found")
	ErrEmptyColumnTitle = errors.New("column title cannot be empty")
	ErrInvalidPosition = errors.New("invalid position")
)

// Card errors
var (
	ErrCardNotFound      = errors.New("card not found")
	ErrEmptyCardTitle    = errors.New("card title cannot be empty")
	ErrInvalidLexorank   = errors.New("invalid lexorank position")
	ErrCardNotInColumn   = errors.New("card does not belong to this column")
)

// Member errors
var (
	ErrMemberNotFound    = errors.New("member not found")
	ErrMemberExists      = errors.New("member already exists")
	ErrCannotRemoveOwner = errors.New("cannot remove owner from board")
	ErrInvalidRole       = errors.New("invalid role")
)
