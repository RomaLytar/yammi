package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	commentpb "github.com/RomaLytar/yammi/services/comment/api/proto/v1"
)

// CreateComment создает новый комментарий
func (s *CommentServiceServer) CreateComment(ctx context.Context, req *commentpb.CreateCommentRequest) (*commentpb.CreateCommentResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetContent() == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	var parentID *string
	if req.GetParentId() != "" {
		pid := req.GetParentId()
		parentID = &pid
	}

	comment, err := s.createComment.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId(), req.GetContent(), parentID)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &commentpb.CreateCommentResponse{
		Comment: mapCommentToProto(comment),
	}, nil
}

// ListComments возвращает список комментариев для карточки
func (s *CommentServiceServer) ListComments(ctx context.Context, req *commentpb.ListCommentsRequest) (*commentpb.ListCommentsResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	comments, nextCursor, err := s.listComments.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId(), int(req.GetLimit()), req.GetCursor())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &commentpb.ListCommentsResponse{
		Comments:   mapCommentsToProto(comments),
		NextCursor: nextCursor,
	}, nil
}

// UpdateComment обновляет текст комментария
func (s *CommentServiceServer) UpdateComment(ctx context.Context, req *commentpb.UpdateCommentRequest) (*commentpb.UpdateCommentResponse, error) {
	if req.GetCommentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "comment_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetContent() == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	comment, err := s.updateComment.Execute(ctx, req.GetCommentId(), req.GetBoardId(), req.GetUserId(), req.GetContent())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &commentpb.UpdateCommentResponse{
		Comment: mapCommentToProto(comment),
	}, nil
}

// DeleteComment удаляет комментарий
func (s *CommentServiceServer) DeleteComment(ctx context.Context, req *commentpb.DeleteCommentRequest) (*emptypb.Empty, error) {
	if req.GetCommentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "comment_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.deleteComment.Execute(ctx, req.GetCommentId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetCommentCount возвращает количество комментариев к карточке
func (s *CommentServiceServer) GetCommentCount(ctx context.Context, req *commentpb.GetCommentCountRequest) (*commentpb.GetCommentCountResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	count, err := s.getCommentCount.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &commentpb.GetCommentCountResponse{
		Count: int32(count),
	}, nil
}
