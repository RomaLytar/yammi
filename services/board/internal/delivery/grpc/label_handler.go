package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// LabelHandler группирует зависимости для операций с метками
type LabelHandler struct {
	create         *usecase.CreateLabelUseCase
	list           *usecase.ListLabelsUseCase
	update         *usecase.UpdateLabelUseCase
	delete         *usecase.DeleteLabelUseCase
	addToCard      *usecase.AddLabelToCardUseCase
	removeFromCard *usecase.RemoveLabelFromCardUseCase
	getForCard     *usecase.GetCardLabelsUseCase
}

func NewLabelHandler(
	create *usecase.CreateLabelUseCase,
	list *usecase.ListLabelsUseCase,
	update *usecase.UpdateLabelUseCase,
	delete_ *usecase.DeleteLabelUseCase,
	addToCard *usecase.AddLabelToCardUseCase,
	removeFromCard *usecase.RemoveLabelFromCardUseCase,
	getForCard *usecase.GetCardLabelsUseCase,
) LabelHandler {
	return LabelHandler{create: create, list: list, update: update, delete: delete_, addToCard: addToCard, removeFromCard: removeFromCard, getForCard: getForCard}
}

// CreateLabel создает новую метку доски
func (s *BoardServiceServer) CreateLabel(ctx context.Context, req *boardpb.CreateLabelRequest) (*boardpb.CreateLabelResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetColor() == "" {
		return nil, status.Error(codes.InvalidArgument, "color is required")
	}

	label, err := s.labels.create.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetName(), req.GetColor())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateLabelResponse{
		Label: mapLabelToProto(label),
	}, nil
}

// ListLabels возвращает все метки доски
func (s *BoardServiceServer) ListLabels(ctx context.Context, req *boardpb.ListLabelsRequest) (*boardpb.ListLabelsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	labels, err := s.labels.list.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListLabelsResponse{
		Labels: mapLabelsToProto(labels),
	}, nil
}

// UpdateLabel обновляет метку
func (s *BoardServiceServer) UpdateLabel(ctx context.Context, req *boardpb.UpdateLabelRequest) (*boardpb.UpdateLabelResponse, error) {
	if req.GetLabelId() == "" {
		return nil, status.Error(codes.InvalidArgument, "label_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetColor() == "" {
		return nil, status.Error(codes.InvalidArgument, "color is required")
	}

	label, err := s.labels.update.Execute(ctx, req.GetLabelId(), req.GetBoardId(), req.GetUserId(), req.GetName(), req.GetColor())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateLabelResponse{
		Label: mapLabelToProto(label),
	}, nil
}

// DeleteLabel удаляет метку (только owner)
func (s *BoardServiceServer) DeleteLabel(ctx context.Context, req *boardpb.DeleteLabelRequest) (*emptypb.Empty, error) {
	if req.GetLabelId() == "" {
		return nil, status.Error(codes.InvalidArgument, "label_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.labels.delete.Execute(ctx, req.GetLabelId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// AddLabelToCard назначает метку на карточку
func (s *BoardServiceServer) AddLabelToCard(ctx context.Context, req *boardpb.AddLabelToCardRequest) (*emptypb.Empty, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetLabelId() == "" {
		return nil, status.Error(codes.InvalidArgument, "label_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.labels.addToCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetLabelId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// RemoveLabelFromCard снимает метку с карточки
func (s *BoardServiceServer) RemoveLabelFromCard(ctx context.Context, req *boardpb.RemoveLabelFromCardRequest) (*emptypb.Empty, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetLabelId() == "" {
		return nil, status.Error(codes.InvalidArgument, "label_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.labels.removeFromCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetLabelId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetCardLabels возвращает все метки карточки
func (s *BoardServiceServer) GetCardLabels(ctx context.Context, req *boardpb.GetCardLabelsRequest) (*boardpb.GetCardLabelsResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	labels, err := s.labels.getForCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetCardLabelsResponse{
		Labels: mapLabelsToProto(labels),
	}, nil
}
