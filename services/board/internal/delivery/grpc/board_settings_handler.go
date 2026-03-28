package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// BoardSettingsHandler группирует зависимости для операций с настройками доски
type BoardSettingsHandler struct {
	get    *usecase.GetBoardSettingsUseCase
	update *usecase.UpdateBoardSettingsUseCase
}

func NewBoardSettingsHandler(
	get *usecase.GetBoardSettingsUseCase,
	update *usecase.UpdateBoardSettingsUseCase,
) BoardSettingsHandler {
	return BoardSettingsHandler{get: get, update: update}
}

// GetBoardSettings возвращает настройки доски
func (s *BoardServiceServer) GetBoardSettings(ctx context.Context, req *boardpb.GetBoardSettingsRequest) (*boardpb.GetBoardSettingsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	settings, err := s.boardSettings.get.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetBoardSettingsResponse{
		Settings: mapBoardSettingsToProto(settings),
	}, nil
}

// UpdateBoardSettings обновляет настройки доски (только owner)
func (s *BoardServiceServer) UpdateBoardSettings(ctx context.Context, req *boardpb.UpdateBoardSettingsRequest) (*boardpb.UpdateBoardSettingsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	settings, err := s.boardSettings.update.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetUseBoardLabelsOnly())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateBoardSettingsResponse{
		Settings: mapBoardSettingsToProto(settings),
	}, nil
}
