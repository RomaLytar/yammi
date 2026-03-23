package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

// CreateCustomField создает новое определение кастомного поля (только owner)
func (s *BoardServiceServer) CreateCustomField(ctx context.Context, req *boardpb.CreateCustomFieldRequest) (*boardpb.CreateCustomFieldResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetFieldType() == "" {
		return nil, status.Error(codes.InvalidArgument, "field_type is required")
	}

	def, err := s.createCustomField.Execute(ctx,
		req.GetBoardId(),
		req.GetUserId(),
		req.GetName(),
		domain.FieldType(req.GetFieldType()),
		req.GetOptions(),
		int(req.GetPosition()),
		req.GetRequired(),
	)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateCustomFieldResponse{
		Definition: mapCustomFieldDefToProto(def),
	}, nil
}

// ListCustomFields возвращает все определения кастомных полей доски
func (s *BoardServiceServer) ListCustomFields(ctx context.Context, req *boardpb.ListCustomFieldsRequest) (*boardpb.ListCustomFieldsResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	defs, err := s.listCustomFields.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListCustomFieldsResponse{
		Definitions: mapCustomFieldDefsToProto(defs),
	}, nil
}

// UpdateCustomField обновляет определение кастомного поля (только owner)
func (s *BoardServiceServer) UpdateCustomField(ctx context.Context, req *boardpb.UpdateCustomFieldRequest) (*boardpb.UpdateCustomFieldResponse, error) {
	if req.GetFieldId() == "" {
		return nil, status.Error(codes.InvalidArgument, "field_id is required")
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

	def, err := s.updateCustomField.Execute(ctx,
		req.GetFieldId(),
		req.GetBoardId(),
		req.GetUserId(),
		req.GetName(),
		req.GetOptions(),
		req.GetRequired(),
	)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateCustomFieldResponse{
		Definition: mapCustomFieldDefToProto(def),
	}, nil
}

// DeleteCustomField удаляет определение кастомного поля (только owner)
func (s *BoardServiceServer) DeleteCustomField(ctx context.Context, req *boardpb.DeleteCustomFieldRequest) (*emptypb.Empty, error) {
	if req.GetFieldId() == "" {
		return nil, status.Error(codes.InvalidArgument, "field_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.deleteCustomField.Execute(ctx, req.GetFieldId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// SetCustomFieldValue устанавливает значение кастомного поля для карточки
func (s *BoardServiceServer) SetCustomFieldValue(ctx context.Context, req *boardpb.SetCustomFieldValueRequest) (*boardpb.SetCustomFieldValueResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetFieldId() == "" {
		return nil, status.Error(codes.InvalidArgument, "field_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var valueText *string
	var valueNumber *float64
	var datePtr *time.Time

	if req.GetHasText() {
		txt := req.GetValueText()
		valueText = &txt
	}
	if req.GetHasNumber() {
		n := req.GetValueNumber()
		valueNumber = &n
	}
	if req.GetHasDate() && req.GetValueDate() != nil {
		t := req.GetValueDate().AsTime()
		datePtr = &t
	}

	value, err := s.setCustomFieldValue.Execute(ctx,
		req.GetCardId(),
		req.GetBoardId(),
		req.GetFieldId(),
		req.GetUserId(),
		valueText,
		valueNumber,
		datePtr,
	)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.SetCustomFieldValueResponse{
		Value: mapCustomFieldValueToProto(value),
	}, nil
}

// GetCardCustomFields возвращает все значения кастомных полей карточки
func (s *BoardServiceServer) GetCardCustomFields(ctx context.Context, req *boardpb.GetCardCustomFieldsRequest) (*boardpb.GetCardCustomFieldsResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	values, err := s.getCardCustomFields.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetCardCustomFieldsResponse{
		Values: mapCustomFieldValuesToProto(values),
	}, nil
}
