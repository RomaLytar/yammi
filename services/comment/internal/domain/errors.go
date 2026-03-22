package domain

import "errors"

// Comment errors
var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrEmptyText       = errors.New("comment text cannot be empty")
	ErrContentTooLong  = errors.New("comment text cannot exceed 10000 characters")
	ErrEmptyCardID     = errors.New("card ID cannot be empty")
	ErrEmptyBoardID    = errors.New("board ID cannot be empty")
	ErrEmptyAuthorID   = errors.New("author ID cannot be empty")
	ErrAccessDenied    = errors.New("access denied")
	ErrNotAuthor       = errors.New("only author can perform this action")
	ErrParentNotFound  = errors.New("parent comment not found")
	ErrNestedReply     = errors.New("replies to replies are not allowed")
)
