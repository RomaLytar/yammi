package http

import (
	commentpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/comment"
)

const timeFormat = "2006-01-02T15:04:05Z07:00"

func mapCommentFromProto(pb *commentpb.Comment) commentResponse {
	resp := commentResponse{
		ID:         pb.Id,
		CardID:     pb.CardId,
		BoardID:    pb.BoardId,
		AuthorID:   pb.AuthorId,
		ParentID:   pb.ParentId,
		Content:    pb.Content,
		ReplyCount: pb.ReplyCount,
	}
	if pb.CreatedAt != nil {
		resp.CreatedAt = pb.CreatedAt.AsTime().Format(timeFormat)
	}
	if pb.UpdatedAt != nil {
		resp.UpdatedAt = pb.UpdatedAt.AsTime().Format(timeFormat)
	}
	return resp
}

func mapCommentsFromProto(pbs []*commentpb.Comment) []commentResponse {
	result := make([]commentResponse, len(pbs))
	for i, pb := range pbs {
		result[i] = mapCommentFromProto(pb)
	}
	return result
}
