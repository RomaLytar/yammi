package http

import (
	"net/http"
	"strconv"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
	userpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/user"
)

func mapBoardFromProto(pb *boardpb.Board) boardResponse {
	return boardResponse{
		ID:          pb.Id,
		Title:       pb.Title,
		Description: pb.Description,
		OwnerID:     pb.OwnerId,
		Version:     pb.Version,
		CreatedAt:   pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapBoardsFromProto(pbs []*boardpb.Board) []boardResponse {
	boards := make([]boardResponse, len(pbs))
	for i, pb := range pbs {
		boards[i] = mapBoardFromProto(pb)
	}
	return boards
}

func mapBoardsWithOwners(pbs []*boardpb.Board, owners map[string]*userpb.UserInfo) []boardResponse {
	boards := make([]boardResponse, len(pbs))
	for i, pb := range pbs {
		b := mapBoardFromProto(pb)
		if owner, ok := owners[pb.OwnerId]; ok {
			b.OwnerName = owner.Name
			b.OwnerAvatarURL = owner.AvatarUrl
		}
		boards[i] = b
	}
	return boards
}

func mapColumnFromProto(pb *boardpb.Column) columnResponse {
	return columnResponse{
		ID:        pb.Id,
		BoardID:   pb.BoardId,
		Title:     pb.Title,
		Position:  pb.Position,
		Version:   pb.Version,
		CreatedAt: pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		CardCount: pb.CardCount,
	}
}

func mapColumnsFromProto(pbs []*boardpb.Column) []columnResponse {
	columns := make([]columnResponse, len(pbs))
	for i, pb := range pbs {
		columns[i] = mapColumnFromProto(pb)
	}
	return columns
}

func mapCardFromProto(pb *boardpb.Card) cardResponse {
	return cardResponse{
		ID:          pb.Id,
		ColumnID:    pb.ColumnId,
		BoardID:     pb.BoardId,
		Title:       pb.Title,
		Description: pb.Description,
		Position:    pb.Position,
		AssigneeID:  pb.AssigneeId,
		CreatorID:   pb.CreatorId,
		Version:     pb.Version,
		CreatedAt:   pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapCardsFromProto(pbs []*boardpb.Card) []cardResponse {
	cards := make([]cardResponse, len(pbs))
	for i, pb := range pbs {
		cards[i] = mapCardFromProto(pb)
	}
	return cards
}

func mapMemberFromProto(pb *boardpb.BoardMember) memberResponse {
	return memberResponse{
		UserID:   pb.UserId,
		Role:     pb.Role,
		Version:  pb.Version,
		JoinedAt: pb.JoinedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapMembersFromProto(pbs []*boardpb.BoardMember) []memberResponse {
	members := make([]memberResponse, len(pbs))
	for i, pb := range pbs {
		members[i] = mapMemberFromProto(pb)
	}
	return members
}

func mapMembersWithProfiles(pbs []*boardpb.BoardMember, profiles map[string]*userpb.UserInfo) []memberResponse {
	members := make([]memberResponse, len(pbs))
	for i, pb := range pbs {
		m := mapMemberFromProto(pb)
		if profile, ok := profiles[pb.UserId]; ok {
			m.Name = profile.Name
			m.Email = profile.Email
			m.AvatarURL = profile.AvatarUrl
		}
		members[i] = m
	}
	return members
}

func mapAttachmentFromProto(pb *boardpb.Attachment) attachmentResponse {
	return attachmentResponse{
		ID:         pb.Id,
		CardID:     pb.CardId,
		BoardID:    pb.BoardId,
		FileName:   pb.FileName,
		FileSize:   pb.FileSize,
		MimeType:   pb.MimeType,
		UploaderID: pb.UploaderId,
		CreatedAt:  pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapAttachmentsFromProto(pbs []*boardpb.Attachment) []attachmentResponse {
	attachments := make([]attachmentResponse, len(pbs))
	for i, pb := range pbs {
		attachments[i] = mapAttachmentFromProto(pb)
	}
	return attachments
}

func parseIntQueryParam(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return n
}
