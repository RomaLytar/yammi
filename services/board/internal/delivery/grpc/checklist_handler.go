package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// ChecklistHandler группирует зависимости для операций с чеклистами
type ChecklistHandler struct {
	create     *usecase.CreateChecklistUseCase
	get        *usecase.GetChecklistsUseCase
	update     *usecase.UpdateChecklistUseCase
	delete     *usecase.DeleteChecklistUseCase
	createItem *usecase.CreateChecklistItemUseCase
	updateItem *usecase.UpdateChecklistItemUseCase
	deleteItem *usecase.DeleteChecklistItemUseCase
	toggleItem *usecase.ToggleChecklistItemUseCase
}

func NewChecklistHandler(
	create *usecase.CreateChecklistUseCase,
	get *usecase.GetChecklistsUseCase,
	update *usecase.UpdateChecklistUseCase,
	delete_ *usecase.DeleteChecklistUseCase,
	createItem *usecase.CreateChecklistItemUseCase,
	updateItem *usecase.UpdateChecklistItemUseCase,
	deleteItem *usecase.DeleteChecklistItemUseCase,
	toggleItem *usecase.ToggleChecklistItemUseCase,
) ChecklistHandler {
	return ChecklistHandler{
		create: create, get: get, update: update, delete: delete_,
		createItem: createItem, updateItem: updateItem, deleteItem: deleteItem, toggleItem: toggleItem,
	}
}

// CreateChecklist создает новый чеклист для карточки
func (s *BoardServiceServer) CreateChecklist(ctx context.Context, req *boardpb.CreateChecklistRequest) (*boardpb.CreateChecklistResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	checklist, err := s.checklists.create.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId(), req.GetTitle(), int(req.GetPosition()))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateChecklistResponse{
		Checklist: mapChecklistToProto(checklist),
	}, nil
}

// GetChecklists возвращает все чеклисты карточки
func (s *BoardServiceServer) GetChecklists(ctx context.Context, req *boardpb.GetChecklistsRequest) (*boardpb.GetChecklistsResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	checklists, err := s.checklists.get.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetChecklistsResponse{
		Checklists: mapChecklistsToProto(checklists),
	}, nil
}

// UpdateChecklist обновляет заголовок чеклиста
func (s *BoardServiceServer) UpdateChecklist(ctx context.Context, req *boardpb.UpdateChecklistRequest) (*boardpb.UpdateChecklistResponse, error) {
	if req.GetChecklistId() == "" {
		return nil, status.Error(codes.InvalidArgument, "checklist_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	checklist, err := s.checklists.update.Execute(ctx, req.GetChecklistId(), req.GetBoardId(), req.GetUserId(), req.GetTitle())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateChecklistResponse{
		Checklist: mapChecklistToProto(checklist),
	}, nil
}

// DeleteChecklist удаляет чеклист
func (s *BoardServiceServer) DeleteChecklist(ctx context.Context, req *boardpb.DeleteChecklistRequest) (*emptypb.Empty, error) {
	if req.GetChecklistId() == "" {
		return nil, status.Error(codes.InvalidArgument, "checklist_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.checklists.delete.Execute(ctx, req.GetChecklistId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// CreateChecklistItem создает новый элемент чеклиста
func (s *BoardServiceServer) CreateChecklistItem(ctx context.Context, req *boardpb.CreateChecklistItemRequest) (*boardpb.CreateChecklistItemResponse, error) {
	if req.GetChecklistId() == "" {
		return nil, status.Error(codes.InvalidArgument, "checklist_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	item, err := s.checklists.createItem.Execute(ctx, req.GetChecklistId(), req.GetBoardId(), req.GetUserId(), req.GetTitle(), int(req.GetPosition()))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateChecklistItemResponse{
		Item: mapChecklistItemToProto(item),
	}, nil
}

// UpdateChecklistItem обновляет заголовок элемента чеклиста
func (s *BoardServiceServer) UpdateChecklistItem(ctx context.Context, req *boardpb.UpdateChecklistItemRequest) (*boardpb.UpdateChecklistItemResponse, error) {
	if req.GetItemId() == "" {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	item, err := s.checklists.updateItem.Execute(ctx, req.GetItemId(), req.GetBoardId(), req.GetUserId(), req.GetTitle())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateChecklistItemResponse{
		Item: mapChecklistItemToProto(item),
	}, nil
}

// DeleteChecklistItem удаляет элемент чеклиста
func (s *BoardServiceServer) DeleteChecklistItem(ctx context.Context, req *boardpb.DeleteChecklistItemRequest) (*emptypb.Empty, error) {
	if req.GetItemId() == "" {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.checklists.deleteItem.Execute(ctx, req.GetItemId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// ToggleChecklistItem переключает состояние элемента чеклиста
func (s *BoardServiceServer) ToggleChecklistItem(ctx context.Context, req *boardpb.ToggleChecklistItemRequest) (*boardpb.ToggleChecklistItemResponse, error) {
	if req.GetItemId() == "" {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	isChecked, err := s.checklists.toggleItem.Execute(ctx, req.GetItemId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ToggleChecklistItemResponse{
		IsChecked: isChecked,
	}, nil
}
