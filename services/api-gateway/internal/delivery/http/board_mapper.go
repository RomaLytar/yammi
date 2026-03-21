package http

import (
	"net/http"
	"strconv"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
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

func mapColumnFromProto(pb *boardpb.Column) columnResponse {
	return columnResponse{
		ID:        pb.Id,
		BoardID:   pb.BoardId,
		Title:     pb.Title,
		Position:  pb.Position,
		Version:   pb.Version,
		CreatedAt: pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
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
