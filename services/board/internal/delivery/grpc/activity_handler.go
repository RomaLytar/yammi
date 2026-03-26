package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
)

// GetCardActivity возвращает журнал активности карточки
func (s *BoardServiceServer) GetCardActivity(ctx context.Context, req *boardpb.GetCardActivityRequest) (*boardpb.GetCardActivityResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}

	activities, nextCursor, err := s.cards.activity.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId(), limit, req.GetCursor())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetCardActivityResponse{
		Entries:    mapActivitiesToProto(activities),
		NextCursor: nextCursor,
	}, nil
}
