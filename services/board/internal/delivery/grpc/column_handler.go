package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
)

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
