package grpc

// NOTE: protoc needs to be run to regenerate Go code from board.proto before this file will compile.
// Run: cd services/board && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/proto/v1/board.proto

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// TemplateHandler группирует зависимости для операций с шаблонами
type TemplateHandler struct {
	createBoardTemplate     *usecase.CreateBoardTemplateUseCase
	listBoardTemplates      *usecase.ListBoardTemplatesUseCase
	deleteBoardTemplate     *usecase.DeleteBoardTemplateUseCase
	createBoardFromTemplate *usecase.CreateBoardFromTemplateUseCase
}

func NewTemplateHandler(
	createBoardTemplate *usecase.CreateBoardTemplateUseCase,
	listBoardTemplates *usecase.ListBoardTemplatesUseCase,
	deleteBoardTemplate *usecase.DeleteBoardTemplateUseCase,
	createBoardFromTemplate *usecase.CreateBoardFromTemplateUseCase,
) TemplateHandler {
	return TemplateHandler{
		createBoardTemplate:     createBoardTemplate,
		listBoardTemplates:      listBoardTemplates,
		deleteBoardTemplate:     deleteBoardTemplate,
		createBoardFromTemplate: createBoardFromTemplate,
	}
}

// ============================================================================
// Board Templates
// ============================================================================

// CreateBoardTemplate создает шаблон доски
func (s *BoardServiceServer) CreateBoardTemplate(ctx context.Context, req *boardpb.CreateBoardTemplateRequest) (*boardpb.CreateBoardTemplateResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// Конвертируем proto данные в domain
	var columnsData []domain.BoardColumnTemplateData
	for _, col := range req.GetColumnsData() {
		columnsData = append(columnsData, domain.BoardColumnTemplateData{
			Title:    col.GetTitle(),
			Position: int(col.GetPosition()),
		})
	}

	var labelsData []domain.LabelTemplateData
	for _, lbl := range req.GetLabelsData() {
		labelsData = append(labelsData, domain.LabelTemplateData{
			Name:  lbl.GetName(),
			Color: lbl.GetColor(),
		})
	}

	tmpl, err := s.templates.createBoardTemplate.Execute(ctx,
		req.GetUserId(), req.GetName(), req.GetDescription(),
		columnsData, labelsData)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateBoardTemplateResponse{
		Template: mapBoardTemplateToProto(tmpl),
	}, nil
}

// ListBoardTemplates возвращает шаблоны досок пользователя
func (s *BoardServiceServer) ListBoardTemplates(ctx context.Context, req *boardpb.ListBoardTemplatesRequest) (*boardpb.ListBoardTemplatesResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	templates, err := s.templates.listBoardTemplates.Execute(ctx, req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListBoardTemplatesResponse{
		Templates: mapBoardTemplatesToProto(templates),
	}, nil
}

// DeleteBoardTemplate удаляет шаблон доски
func (s *BoardServiceServer) DeleteBoardTemplate(ctx context.Context, req *boardpb.DeleteBoardTemplateRequest) (*emptypb.Empty, error) {
	if req.GetTemplateId() == "" {
		return nil, status.Error(codes.InvalidArgument, "template_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.templates.deleteBoardTemplate.Execute(ctx, req.GetTemplateId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// CreateBoardFromTemplate создает доску из шаблона
func (s *BoardServiceServer) CreateBoardFromTemplate(ctx context.Context, req *boardpb.CreateBoardFromTemplateRequest) (*boardpb.CreateBoardResponse, error) {
	if req.GetTemplateId() == "" {
		return nil, status.Error(codes.InvalidArgument, "template_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	board, err := s.templates.createBoardFromTemplate.Execute(ctx,
		req.GetTemplateId(), req.GetTitle(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateBoardResponse{
		Board: mapBoardToProto(board),
	}, nil
}
