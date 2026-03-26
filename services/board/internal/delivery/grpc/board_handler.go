package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// BoardCoreHandler группирует зависимости для операций с досками
type BoardCoreHandler struct {
	create *usecase.CreateBoardUseCase
	get    *usecase.GetBoardUseCase
	list   *usecase.ListBoardsUseCase
	update *usecase.UpdateBoardUseCase
	delete *usecase.DeleteBoardUseCase
}

func NewBoardCoreHandler(
	create *usecase.CreateBoardUseCase,
	get *usecase.GetBoardUseCase,
	list *usecase.ListBoardsUseCase,
	update *usecase.UpdateBoardUseCase,
	delete *usecase.DeleteBoardUseCase,
) BoardCoreHandler {
	return BoardCoreHandler{create: create, get: get, list: list, update: update, delete: delete}
}

// CreateBoard создает новую доску
func (s *BoardServiceServer) CreateBoard(ctx context.Context, req *boardpb.CreateBoardRequest) (*boardpb.CreateBoardResponse, error) {
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetOwnerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "owner_id is required")
	}

	board, err := s.boards.create.Execute(ctx, req.GetTitle(), req.GetDescription(), req.GetOwnerId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateBoardResponse{
		Board: mapBoardToProto(board),
	}, nil
}

// GetBoard возвращает доску со всеми колонками и участниками
func (s *BoardServiceServer) GetBoard(ctx context.Context, req *boardpb.GetBoardRequest) (*boardpb.GetBoardResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// 1. Загружаем доску (включает проверку IsMember)
	board, err := s.boards.get.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	// 2. Доступ проверен — загружаем columns и members без повторного IsMember
	columns, err := s.columns.get.ExecuteAuthorized(ctx, req.GetBoardId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	members, err := s.members.list.ExecuteAuthorized(ctx, req.GetBoardId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	// 3. Подсчёт карточек по колонкам (один запрос)
	cardCounts, _ := s.cards.repo.CountByBoard(ctx, req.GetBoardId())

	return &boardpb.GetBoardResponse{
		Board:   mapBoardToProto(board),
		Columns: mapColumnsWithCountsToProto(columns, cardCounts),
		Members: mapMembersToProto(members),
	}, nil
}

// ListBoards возвращает список досок пользователя с cursor-based pagination
func (s *BoardServiceServer) ListBoards(ctx context.Context, req *boardpb.ListBoardsRequest) (*boardpb.ListBoardsResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20 // default limit
	}
	if limit > 100 {
		limit = 100 // max limit
	}

	boards, nextCursor, err := s.boards.list.Execute(ctx, req.GetUserId(), limit, req.GetCursor(), req.GetOwnerOnly(), req.GetSearch(), req.GetSortBy())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListBoardsResponse{
		Boards:     mapBoardsToProto(boards),
		NextCursor: nextCursor,
	}, nil
}

// UpdateBoard обновляет метаданные доски
func (s *BoardServiceServer) UpdateBoard(ctx context.Context, req *boardpb.UpdateBoardRequest) (*boardpb.UpdateBoardResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	board, err := s.boards.update.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetTitle(), req.GetDescription(), int(req.GetVersion()))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateBoardResponse{
		Board: mapBoardToProto(board),
	}, nil
}

// DeleteBoard удаляет одну или несколько досок (batch)
func (s *BoardServiceServer) DeleteBoard(ctx context.Context, req *boardpb.DeleteBoardRequest) (*emptypb.Empty, error) {
	if len(req.GetBoardIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "board_ids is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.boards.delete.Execute(ctx, req.GetBoardIds(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}
