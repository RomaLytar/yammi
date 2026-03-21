package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// BoardServiceServer реализует gRPC сервер Board Service
type BoardServiceServer struct {
	boardpb.UnimplementedBoardServiceServer

	// Board Use Cases
	createBoard *usecase.CreateBoardUseCase
	getBoard    *usecase.GetBoardUseCase
	listBoards  *usecase.ListBoardsUseCase
	updateBoard *usecase.UpdateBoardUseCase
	deleteBoard *usecase.DeleteBoardUseCase

	// Column Use Cases
	addColumn      *usecase.AddColumnUseCase
	getColumns     *usecase.GetColumnsUseCase
	updateColumn   *usecase.UpdateColumnUseCase
	deleteColumn   *usecase.DeleteColumnUseCase
	reorderColumns *usecase.ReorderColumnsUseCase

	// Card Use Cases
	createCard *usecase.CreateCardUseCase
	getCard    *usecase.GetCardUseCase
	getCards   *usecase.GetCardsUseCase
	updateCard *usecase.UpdateCardUseCase
	moveCard   *usecase.MoveCardUseCase
	deleteCard *usecase.DeleteCardUseCase

	// Member Use Cases
	addMember    *usecase.AddMemberUseCase
	removeMember *usecase.RemoveMemberUseCase
	listMembers  *usecase.ListMembersUseCase
}

// NewBoardServiceServer создает новый gRPC сервер с внедренными use cases
func NewBoardServiceServer(
	createBoard *usecase.CreateBoardUseCase,
	getBoard *usecase.GetBoardUseCase,
	listBoards *usecase.ListBoardsUseCase,
	updateBoard *usecase.UpdateBoardUseCase,
	deleteBoard *usecase.DeleteBoardUseCase,
	addColumn *usecase.AddColumnUseCase,
	getColumns *usecase.GetColumnsUseCase,
	updateColumn *usecase.UpdateColumnUseCase,
	deleteColumn *usecase.DeleteColumnUseCase,
	reorderColumns *usecase.ReorderColumnsUseCase,
	createCard *usecase.CreateCardUseCase,
	getCard *usecase.GetCardUseCase,
	getCards *usecase.GetCardsUseCase,
	updateCard *usecase.UpdateCardUseCase,
	moveCard *usecase.MoveCardUseCase,
	deleteCard *usecase.DeleteCardUseCase,
	addMember *usecase.AddMemberUseCase,
	removeMember *usecase.RemoveMemberUseCase,
	listMembers *usecase.ListMembersUseCase,
) *BoardServiceServer {
	return &BoardServiceServer{
		createBoard:    createBoard,
		getBoard:       getBoard,
		listBoards:     listBoards,
		updateBoard:    updateBoard,
		deleteBoard:    deleteBoard,
		addColumn:      addColumn,
		getColumns:     getColumns,
		updateColumn:   updateColumn,
		deleteColumn:   deleteColumn,
		reorderColumns: reorderColumns,
		createCard:     createCard,
		getCard:        getCard,
		getCards:       getCards,
		updateCard:     updateCard,
		moveCard:       moveCard,
		deleteCard:     deleteCard,
		addMember:      addMember,
		removeMember:   removeMember,
		listMembers:    listMembers,
	}
}

// ============================================================================
// Board Operations
// ============================================================================

// CreateBoard создает новую доску
func (s *BoardServiceServer) CreateBoard(ctx context.Context, req *boardpb.CreateBoardRequest) (*boardpb.CreateBoardResponse, error) {
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetOwnerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "owner_id is required")
	}

	board, err := s.createBoard.Execute(ctx, req.GetTitle(), req.GetDescription(), req.GetOwnerId())
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

	board, err := s.getBoard.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	// Загружаем columns и members отдельно (granular API)
	columns, err := s.getColumns.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	members, err := s.listMembers.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetBoardResponse{
		Board:   mapBoardToProto(board),
		Columns: mapColumnsToProto(columns),
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

	boards, nextCursor, err := s.listBoards.Execute(ctx, req.GetUserId(), limit, req.GetCursor(), req.GetOwnerOnly(), req.GetSearch(), req.GetSortBy())
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

	board, err := s.updateBoard.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetTitle(), req.GetDescription(), int(req.GetVersion()))
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

	err := s.deleteBoard.Execute(ctx, req.GetBoardIds(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// ============================================================================
// Column Operations
// ============================================================================

// AddColumn добавляет колонку в доску
func (s *BoardServiceServer) AddColumn(ctx context.Context, req *boardpb.AddColumnRequest) (*boardpb.AddColumnResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	column, err := s.addColumn.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetTitle(), int(req.GetPosition()))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.AddColumnResponse{
		Column: mapColumnToProto(column),
	}, nil
}

// GetColumns возвращает все колонки доски
func (s *BoardServiceServer) GetColumns(ctx context.Context, req *boardpb.GetColumnsRequest) (*boardpb.GetColumnsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	columns, err := s.getColumns.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetColumnsResponse{
		Columns: mapColumnsToProto(columns),
	}, nil
}

// UpdateColumn обновляет заголовок колонки
func (s *BoardServiceServer) UpdateColumn(ctx context.Context, req *boardpb.UpdateColumnRequest) (*boardpb.UpdateColumnResponse, error) {
	if req.GetColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "column_id is required")
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

	column, err := s.updateColumn.Execute(ctx, req.GetColumnId(), req.GetBoardId(), req.GetUserId(), req.GetTitle(), int(req.GetVersion()))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateColumnResponse{
		Column: mapColumnToProto(column),
	}, nil
}

// DeleteColumn удаляет колонку
func (s *BoardServiceServer) DeleteColumn(ctx context.Context, req *boardpb.DeleteColumnRequest) (*emptypb.Empty, error) {
	if req.GetColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "column_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.deleteColumn.Execute(ctx, req.GetColumnId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// ReorderColumns изменяет порядок колонок
func (s *BoardServiceServer) ReorderColumns(ctx context.Context, req *boardpb.ReorderColumnsRequest) (*boardpb.ReorderColumnsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if len(req.GetPositions()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "positions are required")
	}

	// Преобразуем позиции в map для use case
	positions := make(map[string]int)
	for _, pos := range req.GetPositions() {
		positions[pos.GetColumnId()] = int(pos.GetPosition())
	}

	columns, err := s.reorderColumns.Execute(ctx, req.GetBoardId(), req.GetUserId(), positions, int(req.GetVersion()))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ReorderColumnsResponse{
		Columns: mapColumnsToProto(columns),
	}, nil
}

// ============================================================================
// Card Operations
// ============================================================================

// CreateCard создает новую карточку
func (s *BoardServiceServer) CreateCard(ctx context.Context, req *boardpb.CreateCardRequest) (*boardpb.CreateCardResponse, error) {
	if req.GetColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "column_id is required")
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
	if req.GetPosition() == "" {
		return nil, status.Error(codes.InvalidArgument, "position (lexorank) is required")
	}

	// CreateCardRequest не имеет assignee_id в proto (line 153-160), передаем nil
	card, err := s.createCard.Execute(ctx, req.GetColumnId(), req.GetBoardId(), req.GetUserId(), req.GetTitle(), req.GetDescription(), req.GetPosition(), nil)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateCardResponse{
		Card: mapCardToProto(card, req.GetBoardId()),
	}, nil
}

// GetCard возвращает карточку по ID
func (s *BoardServiceServer) GetCard(ctx context.Context, req *boardpb.GetCardRequest) (*boardpb.GetCardResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	card, err := s.getCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetCardResponse{
		Card: mapCardToProto(card, req.GetBoardId()),
	}, nil
}

// GetCards возвращает все карточки колонки
func (s *BoardServiceServer) GetCards(ctx context.Context, req *boardpb.GetCardsRequest) (*boardpb.GetCardsResponse, error) {
	if req.GetColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "column_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	cards, err := s.getCards.Execute(ctx, req.GetColumnId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetCardsResponse{
		Cards: mapCardsToProto(cards, req.GetBoardId()),
	}, nil
}

// UpdateCard обновляет метаданные карточки
func (s *BoardServiceServer) UpdateCard(ctx context.Context, req *boardpb.UpdateCardRequest) (*boardpb.UpdateCardResponse, error) {
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

	// Преобразуем assignee_id в *string
	var assigneeID *string
	if req.GetAssigneeId() != "" {
		assigneeID = stringPtr(req.GetAssigneeId())
	}

	card, err := s.updateCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId(), req.GetTitle(), req.GetDescription(), assigneeID, int(req.GetVersion()))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateCardResponse{
		Card: mapCardToProto(card, req.GetBoardId()),
	}, nil
}

// MoveCard перемещает карточку между колонками (с lexorank позицией)
func (s *BoardServiceServer) MoveCard(ctx context.Context, req *boardpb.MoveCardRequest) (*boardpb.MoveCardResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetFromColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "from_column_id is required")
	}
	if req.GetToColumnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "to_column_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetPosition() == "" {
		return nil, status.Error(codes.InvalidArgument, "position (lexorank) is required")
	}

	// Передаём lexorank позицию напрямую (вычисляется на фронтенде)
	card, err := s.moveCard.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetFromColumnId(), req.GetToColumnId(), req.GetUserId(), req.GetPosition())
	if err != nil {
		return nil, mapDomainError(err)
	}

	// Возвращаем обновленный список карточек в целевой колонке (для фронта)
	cardsInColumn, err := s.getCards.Execute(ctx, req.GetToColumnId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.MoveCardResponse{
		Card:          mapCardToProto(card, req.GetBoardId()),
		CardsInColumn: mapCardsToProto(cardsInColumn, req.GetBoardId()),
	}, nil
}

// DeleteCard удаляет одну или несколько карточек (batch)
func (s *BoardServiceServer) DeleteCard(ctx context.Context, req *boardpb.DeleteCardRequest) (*emptypb.Empty, error) {
	if len(req.GetCardIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "card_ids is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.deleteCard.Execute(ctx, req.GetCardIds(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// ============================================================================
// Member Operations
// ============================================================================

// AddMember добавляет участника в доску (только owner)
func (s *BoardServiceServer) AddMember(ctx context.Context, req *boardpb.AddMemberRequest) (*boardpb.AddMemberResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetMemberUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "member_user_id is required")
	}
	if req.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}

	role := domain.Role(req.GetRole())
	if !role.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	err := s.addMember.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetMemberUserId(), role)
	if err != nil {
		return nil, mapDomainError(err)
	}

	// AddMemberUseCase не возвращает member, загружаем его отдельно
	members, err := s.listMembers.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	// Находим добавленного member
	var addedMember *domain.Member
	for _, m := range members {
		if m.UserID == req.GetMemberUserId() {
			addedMember = m
			break
		}
	}

	if addedMember == nil {
		return nil, status.Error(codes.Internal, "member was added but not found")
	}

	return &boardpb.AddMemberResponse{
		Member: mapMemberToProto(addedMember),
	}, nil
}

// RemoveMember удаляет участника из доски (только owner)
func (s *BoardServiceServer) RemoveMember(ctx context.Context, req *boardpb.RemoveMemberRequest) (*emptypb.Empty, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetMemberUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "member_user_id is required")
	}

	err := s.removeMember.Execute(ctx, req.GetBoardId(), req.GetUserId(), req.GetMemberUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// ListMembers возвращает список участников доски
func (s *BoardServiceServer) ListMembers(ctx context.Context, req *boardpb.ListMembersRequest) (*boardpb.ListMembersResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	members, err := s.listMembers.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListMembersResponse{
		Members: mapMembersToProto(members),
	}, nil
}

// ============================================================================
// Mappers (domain → proto)
// ============================================================================

func mapBoardToProto(b *domain.Board) *boardpb.Board {
	return &boardpb.Board{
		Id:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		OwnerId:     b.OwnerID,
		Version:     int32(b.Version),
		CreatedAt:   timestamppb.New(b.CreatedAt),
		UpdatedAt:   timestamppb.New(b.UpdatedAt),
	}
}

func mapBoardsToProto(boards []*domain.Board) []*boardpb.Board {
	result := make([]*boardpb.Board, len(boards))
	for i, b := range boards {
		result[i] = mapBoardToProto(b)
	}
	return result
}

func mapColumnToProto(c *domain.Column) *boardpb.Column {
	return &boardpb.Column{
		Id:        c.ID,
		BoardId:   c.BoardID,
		Title:     c.Title,
		Position:  int32(c.Position),
		Version:   0, // Column не имеет version в domain.Column, но есть в proto
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.CreatedAt), // domain.Column не имеет UpdatedAt
	}
}

func mapColumnsToProto(columns []*domain.Column) []*boardpb.Column {
	result := make([]*boardpb.Column, len(columns))
	for i, c := range columns {
		result[i] = mapColumnToProto(c)
	}
	return result
}

func mapCardToProto(c *domain.Card, boardID string) *boardpb.Card {
	assigneeID := ""
	if c.AssigneeID != nil {
		assigneeID = *c.AssigneeID
	}

	return &boardpb.Card{
		Id:          c.ID,
		ColumnId:    c.ColumnID,
		BoardId:     boardID,
		Title:       c.Title,
		Description: c.Description,
		Position:    c.Position, // lexorank string
		AssigneeId:  assigneeID,
		Version:     0, // Card не имеет version в domain.Card, но есть в proto
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(c.UpdatedAt),
		CreatorId:   c.CreatorID,
	}
}

func mapCardsToProto(cards []*domain.Card, boardID string) []*boardpb.Card {
	result := make([]*boardpb.Card, len(cards))
	for i, c := range cards {
		result[i] = mapCardToProto(c, boardID)
	}
	return result
}

func mapMemberToProto(m *domain.Member) *boardpb.BoardMember {
	return &boardpb.BoardMember{
		UserId:   m.UserID,
		Role:     m.Role.String(),
		Version:  0, // Member не имеет version в domain.Member, но есть в proto
		JoinedAt: timestamppb.New(m.JoinedAt),
	}
}

func mapMembersToProto(members []*domain.Member) []*boardpb.BoardMember {
	result := make([]*boardpb.BoardMember, len(members))
	for i, m := range members {
		result[i] = mapMemberToProto(m)
	}
	return result
}

// ============================================================================
// Error Mapping (domain errors → gRPC codes)
// ============================================================================

func mapDomainError(err error) error {
	// NotFound errors
	if errors.Is(err, domain.ErrBoardNotFound) ||
		errors.Is(err, domain.ErrColumnNotFound) ||
		errors.Is(err, domain.ErrCardNotFound) ||
		errors.Is(err, domain.ErrMemberNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}

	// Permission errors
	if errors.Is(err, domain.ErrAccessDenied) ||
		errors.Is(err, domain.ErrNotOwner) {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	// InvalidArgument errors
	if errors.Is(err, domain.ErrEmptyTitle) ||
		errors.Is(err, domain.ErrEmptyColumnTitle) ||
		errors.Is(err, domain.ErrEmptyCardTitle) ||
		errors.Is(err, domain.ErrEmptyOwnerID) ||
		errors.Is(err, domain.ErrInvalidLexorank) ||
		errors.Is(err, domain.ErrInvalidRole) ||
		errors.Is(err, domain.ErrInvalidPosition) ||
		errors.Is(err, domain.ErrCardNotInColumn) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// AlreadyExists errors
	if errors.Is(err, domain.ErrMemberExists) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	// FailedPrecondition errors
	if errors.Is(err, domain.ErrCannotRemoveOwner) {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	// Aborted errors (optimistic locking conflict)
	if errors.Is(err, domain.ErrInvalidVersion) {
		return status.Error(codes.Aborted, err.Error())
	}

	// Default internal error
	return status.Error(codes.Internal, "internal server error")
}

// ============================================================================
// Helper Functions
// ============================================================================

func stringPtr(s string) *string {
	return &s
}
