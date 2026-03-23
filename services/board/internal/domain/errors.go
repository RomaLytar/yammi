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
	ErrAssigneeNotMember = errors.New("assignee is not a board member")
)

// Member errors
var (
	ErrMemberNotFound    = errors.New("member not found")
	ErrMemberExists      = errors.New("member already exists")
	ErrCannotRemoveOwner = errors.New("cannot remove owner from board")
	ErrInvalidRole       = errors.New("invalid role")
)

// Attachment errors
var (
	ErrAttachmentNotFound    = errors.New("attachment not found")
	ErrFileTooLarge          = errors.New("file size exceeds maximum allowed (50 MB)")
	ErrMaxAttachmentsReached = errors.New("maximum number of attachments per card reached")
	ErrEmptyFileName         = errors.New("file name cannot be empty")
)

// Label errors
var (
	ErrLabelNotFound      = errors.New("label not found")
	ErrLabelExists        = errors.New("label with this name already exists")
	ErrEmptyLabelName     = errors.New("label name cannot be empty")
	ErrInvalidColor       = errors.New("invalid color format")
	ErrMaxLabelsReached   = errors.New("maximum labels per board reached")
	ErrLabelAlreadyOnCard = errors.New("label already assigned to card")
)

// Activity errors
var (
	ErrEmptyActorID      = errors.New("actor ID cannot be empty")
	ErrInvalidActivityType = errors.New("invalid activity type")
)
