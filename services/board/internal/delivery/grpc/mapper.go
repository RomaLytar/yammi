package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

// ============================================================================
// Mappers (domain → proto)
// ============================================================================

func mapBoardToProto(b *domain.Board) *boardpb.Board {
	return &boardpb.Board{
		Id:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		OwnerId:     b.OwnerID,
		Version:     int32(b.Version),
		CreatedAt:   timestamppb.New(b.CreatedAt),
		UpdatedAt:   timestamppb.New(b.UpdatedAt),
	}
}

func mapBoardsToProto(boards []*domain.Board) []*boardpb.Board {
	result := make([]*boardpb.Board, len(boards))
	for i, b := range boards {
		result[i] = mapBoardToProto(b)
	}
	return result
}

func mapColumnToProto(c *domain.Column) *boardpb.Column {
	return &boardpb.Column{
		Id:        c.ID,
		BoardId:   c.BoardID,
		Title:     c.Title,
		Position:  int32(c.Position),
		Version:   0, // Column не имеет version в domain.Column, но есть в proto
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func mapColumnsToProto(columns []*domain.Column) []*boardpb.Column {
	result := make([]*boardpb.Column, len(columns))
	for i, c := range columns {
		result[i] = mapColumnToProto(c)
	}
	return result
}

func mapColumnsWithCountsToProto(columns []*domain.Column, counts map[string]int) []*boardpb.Column {
	result := make([]*boardpb.Column, len(columns))
	for i, c := range columns {
		col := mapColumnToProto(c)
		if counts != nil {
			col.CardCount = int32(counts[c.ID])
		}
		result[i] = col
	}
	return result
}

func mapCardToProto(c *domain.Card, boardID string) *boardpb.Card {
	assigneeID := ""
	if c.AssigneeID != nil {
		assigneeID = *c.AssigneeID
	}

	return &boardpb.Card{
		Id:          c.ID,
		ColumnId:    c.ColumnID,
		BoardId:     boardID,
		Title:       c.Title,
		Description: c.Description,
		Position:    c.Position, // lexorank string
		AssigneeId:  assigneeID,
		Version:     0, // Card не имеет version в domain.Card, но есть в proto
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(c.UpdatedAt),
		CreatorId:   c.CreatorID,
	}
}

func mapCardsToProto(cards []*domain.Card, boardID string) []*boardpb.Card {
	result := make([]*boardpb.Card, len(cards))
	for i, c := range cards {
		result[i] = mapCardToProto(c, boardID)
	}
	return result
}

func mapMemberToProto(m *domain.Member) *boardpb.BoardMember {
	return &boardpb.BoardMember{
		UserId:   m.UserID,
		Role:     m.Role.String(),
		Version:  0, // Member не имеет version в domain.Member, но есть в proto
		JoinedAt: timestamppb.New(m.JoinedAt),
	}
}

func mapMembersToProto(members []*domain.Member) []*boardpb.BoardMember {
	result := make([]*boardpb.BoardMember, len(members))
	for i, m := range members {
		result[i] = mapMemberToProto(m)
	}
	return result
}

func mapAttachmentToProto(a *domain.Attachment) *boardpb.Attachment {
	return &boardpb.Attachment{
		Id:         a.ID,
		CardId:     a.CardID,
		BoardId:    a.BoardID,
		FileName:   a.FileName,
		FileSize:   a.FileSize,
		MimeType:   a.MimeType,
		StorageKey: a.StorageKey,
		UploaderId: a.UploaderID,
		CreatedAt:  timestamppb.New(a.CreatedAt),
	}
}

func mapAttachmentsToProto(attachments []*domain.Attachment) []*boardpb.Attachment {
	result := make([]*boardpb.Attachment, len(attachments))
	for i, a := range attachments {
		result[i] = mapAttachmentToProto(a)
	}
	return result
}

func mapActivityToProto(a *domain.Activity) *boardpb.ActivityEntry {
	return &boardpb.ActivityEntry{
		Id:           a.ID,
		CardId:       a.CardID,
		BoardId:      a.BoardID,
		ActorId:      a.ActorID,
		ActivityType: string(a.Type),
		Description:  a.Description,
		Changes:      a.Changes,
		CreatedAt:    timestamppb.New(a.CreatedAt),
	}
}

func mapActivitiesToProto(activities []*domain.Activity) []*boardpb.ActivityEntry {
	result := make([]*boardpb.ActivityEntry, len(activities))
	for i, a := range activities {
		result[i] = mapActivityToProto(a)
	}
	return result
}

func mapLabelToProto(l *domain.Label) *boardpb.Label {
	return &boardpb.Label{
		Id:        l.ID,
		BoardId:   l.BoardID,
		Name:      l.Name,
		Color:     l.Color,
		CreatedAt: timestamppb.New(l.CreatedAt),
	}
}

func mapLabelsToProto(labels []*domain.Label) []*boardpb.Label {
	result := make([]*boardpb.Label, len(labels))
	for i, l := range labels {
		result[i] = mapLabelToProto(l)
	}
	return result
}

// ============================================================================
// Error Mapping (domain errors → gRPC codes)
// ============================================================================

func mapDomainError(err error) error {
	// NotFound errors
	if errors.Is(err, domain.ErrBoardNotFound) ||
		errors.Is(err, domain.ErrColumnNotFound) ||
		errors.Is(err, domain.ErrCardNotFound) ||
		errors.Is(err, domain.ErrMemberNotFound) ||
		errors.Is(err, domain.ErrAttachmentNotFound) ||
		errors.Is(err, domain.ErrLabelNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}

	// Permission errors
	if errors.Is(err, domain.ErrAccessDenied) ||
		errors.Is(err, domain.ErrNotOwner) {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	// InvalidArgument errors
	if errors.Is(err, domain.ErrEmptyTitle) ||
		errors.Is(err, domain.ErrEmptyColumnTitle) ||
		errors.Is(err, domain.ErrEmptyCardTitle) ||
		errors.Is(err, domain.ErrEmptyOwnerID) ||
		errors.Is(err, domain.ErrInvalidLexorank) ||
		errors.Is(err, domain.ErrInvalidRole) ||
		errors.Is(err, domain.ErrInvalidPosition) ||
		errors.Is(err, domain.ErrCardNotInColumn) ||
		errors.Is(err, domain.ErrAssigneeNotMember) ||
		errors.Is(err, domain.ErrFileTooLarge) ||
		errors.Is(err, domain.ErrEmptyFileName) ||
		errors.Is(err, domain.ErrEmptyLabelName) ||
		errors.Is(err, domain.ErrInvalidColor) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// ResourceExhausted errors
	if errors.Is(err, domain.ErrMaxAttachmentsReached) ||
		errors.Is(err, domain.ErrMaxLabelsReached) {
		return status.Error(codes.ResourceExhausted, err.Error())
	}

	// AlreadyExists errors
	if errors.Is(err, domain.ErrMemberExists) ||
		errors.Is(err, domain.ErrLabelExists) ||
		errors.Is(err, domain.ErrLabelAlreadyOnCard) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	// FailedPrecondition errors
	if errors.Is(err, domain.ErrCannotRemoveOwner) {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	// Aborted errors (optimistic locking conflict)
	if errors.Is(err, domain.ErrInvalidVersion) {
		return status.Error(codes.Aborted, err.Error())
	}

	// Default internal error
	return status.Error(codes.Internal, "internal server error")
}
