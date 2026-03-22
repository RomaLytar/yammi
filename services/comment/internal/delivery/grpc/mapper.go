package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commentpb "github.com/RomaLytar/yammi/services/comment/api/proto/v1"
	"github.com/RomaLytar/yammi/services/comment/internal/domain"
)

// ============================================================================
// Mappers (domain -> proto)
// ============================================================================

func mapCommentToProto(c *domain.Comment) *commentpb.Comment {
	parentID := ""
	if c.ParentID != nil {
		parentID = *c.ParentID
	}

	return &commentpb.Comment{
		Id:         c.ID,
		CardId:     c.CardID,
		BoardId:    c.BoardID,
		AuthorId:   c.AuthorID,
		ParentId:   parentID,
		Content:    c.Content,
		ReplyCount: int32(c.ReplyCount),
		CreatedAt:  timestamppb.New(c.CreatedAt),
		UpdatedAt:  timestamppb.New(c.UpdatedAt),
	}
}

func mapCommentsToProto(comments []*domain.Comment) []*commentpb.Comment {
	result := make([]*commentpb.Comment, len(comments))
	for i, c := range comments {
		result[i] = mapCommentToProto(c)
	}
	return result
}

// ============================================================================
// Error Mapping (domain errors -> gRPC codes)
// ============================================================================

func mapDomainError(err error) error {
	// NotFound errors
	if errors.Is(err, domain.ErrCommentNotFound) ||
		errors.Is(err, domain.ErrParentNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}

	// Permission errors
	if errors.Is(err, domain.ErrAccessDenied) ||
		errors.Is(err, domain.ErrNotAuthor) {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	// InvalidArgument errors
	if errors.Is(err, domain.ErrEmptyText) ||
		errors.Is(err, domain.ErrContentTooLong) ||
		errors.Is(err, domain.ErrEmptyCardID) ||
		errors.Is(err, domain.ErrEmptyBoardID) ||
		errors.Is(err, domain.ErrEmptyAuthorID) ||
		errors.Is(err, domain.ErrNestedReply) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Default internal error
	return status.Error(codes.Internal, "internal server error")
}
