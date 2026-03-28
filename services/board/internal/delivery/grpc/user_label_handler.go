package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// UserLabelHandler группирует зависимости для операций с пользовательскими метками
type UserLabelHandler struct {
	create          *usecase.CreateUserLabelUseCase
	list            *usecase.ListUserLabelsUseCase
	update          *usecase.UpdateUserLabelUseCase
	delete          *usecase.DeleteUserLabelUseCase
	listAvailable   *usecase.ListAvailableLabelsUseCase
}

func NewUserLabelHandler(
	create *usecase.CreateUserLabelUseCase,
	list *usecase.ListUserLabelsUseCase,
	update *usecase.UpdateUserLabelUseCase,
	delete_ *usecase.DeleteUserLabelUseCase,
	listAvailable *usecase.ListAvailableLabelsUseCase,
) UserLabelHandler {
	return UserLabelHandler{
		create:        create,
		list:          list,
		update:        update,
		delete:        delete_,
		listAvailable: listAvailable,
	}
}

// CreateUserLabel создает новую пользовательскую метку
func (s *BoardServiceServer) CreateUserLabel(ctx context.Context, req *boardpb.CreateUserLabelRequest) (*boardpb.CreateUserLabelResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetColor() == "" {
		return nil, status.Error(codes.InvalidArgument, "color is required")
	}

	label, err := s.userLabels.create.Execute(ctx, req.GetUserId(), req.GetName(), req.GetColor())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateUserLabelResponse{
		Label: mapUserLabelToProto(label),
	}, nil
}

// ListUserLabels возвращает все пользовательские метки
func (s *BoardServiceServer) ListUserLabels(ctx context.Context, req *boardpb.ListUserLabelsRequest) (*boardpb.ListUserLabelsResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	labels, err := s.userLabels.list.Execute(ctx, req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListUserLabelsResponse{
		Labels: mapUserLabelsToProto(labels),
	}, nil
}

// UpdateUserLabel обновляет пользовательскую метку
func (s *BoardServiceServer) UpdateUserLabel(ctx context.Context, req *boardpb.UpdateUserLabelRequest) (*boardpb.UpdateUserLabelResponse, error) {
	if req.GetLabelId() == "" {
		return nil, status.Error(codes.InvalidArgument, "label_id is required")
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

	label, err := s.userLabels.update.Execute(ctx, req.GetLabelId(), req.GetUserId(), req.GetName(), req.GetColor())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateUserLabelResponse{
		Label: mapUserLabelToProto(label),
	}, nil
}

// DeleteUserLabel удаляет пользовательскую метку
func (s *BoardServiceServer) DeleteUserLabel(ctx context.Context, req *boardpb.DeleteUserLabelRequest) (*emptypb.Empty, error) {
	if req.GetLabelId() == "" {
		return nil, status.Error(codes.InvalidArgument, "label_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.userLabels.delete.Execute(ctx, req.GetLabelId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// ListAvailableLabels возвращает все доступные метки для доски
func (s *BoardServiceServer) ListAvailableLabels(ctx context.Context, req *boardpb.ListAvailableLabelsRequest) (*boardpb.ListAvailableLabelsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	result, err := s.userLabels.listAvailable.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListAvailableLabelsResponse{
		BoardLabels:        mapLabelsToProto(result.BoardLabels),
		UserLabels:         mapUserLabelsToProto(result.UserLabels),
		UseBoardLabelsOnly: result.UseBoardLabelsOnly,
	}, nil
}
