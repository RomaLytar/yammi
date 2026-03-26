package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// AttachmentHandler группирует зависимости для операций с вложениями
type AttachmentHandler struct {
	upload      *usecase.UploadAttachmentUseCase
	confirm     *usecase.ConfirmUploadUseCase
	downloadURL *usecase.GetDownloadURLUseCase
	list        *usecase.ListAttachmentsUseCase
	delete      *usecase.DeleteAttachmentUseCase
}

func NewAttachmentHandler(
	upload *usecase.UploadAttachmentUseCase,
	confirm *usecase.ConfirmUploadUseCase,
	downloadURL *usecase.GetDownloadURLUseCase,
	list *usecase.ListAttachmentsUseCase,
	delete_ *usecase.DeleteAttachmentUseCase,
) AttachmentHandler {
	return AttachmentHandler{upload: upload, confirm: confirm, downloadURL: downloadURL, list: list, delete: delete_}
}

// CreateUploadURL создает метаданные вложения и возвращает pre-signed URL для загрузки
func (s *BoardServiceServer) CreateUploadURL(ctx context.Context, req *boardpb.CreateUploadURLRequest) (*boardpb.CreateUploadURLResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetFileName() == "" {
		return nil, status.Error(codes.InvalidArgument, "file_name is required")
	}
	if req.GetFileSize() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "file_size must be positive")
	}

	attachment, uploadURL, err := s.attachments.upload.Execute(
		ctx,
		req.GetCardId(),
		req.GetBoardId(),
		req.GetUserId(),
		req.GetFileName(),
		req.GetContentType(),
		req.GetFileSize(),
	)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateUploadURLResponse{
		Attachment: mapAttachmentToProto(attachment),
		UploadUrl:  uploadURL,
	}, nil
}

// ConfirmUpload подтверждает загрузку файла (проверяет существование в хранилище)
func (s *BoardServiceServer) ConfirmUpload(ctx context.Context, req *boardpb.ConfirmUploadRequest) (*boardpb.ConfirmUploadResponse, error) {
	if req.GetAttachmentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "attachment_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	attachment, err := s.attachments.confirm.Execute(ctx, req.GetAttachmentId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ConfirmUploadResponse{
		Attachment: mapAttachmentToProto(attachment),
	}, nil
}

// GetDownloadURL генерирует pre-signed URL для скачивания
func (s *BoardServiceServer) GetDownloadURL(ctx context.Context, req *boardpb.GetDownloadURLRequest) (*boardpb.GetDownloadURLResponse, error) {
	if req.GetAttachmentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "attachment_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	downloadURL, err := s.attachments.downloadURL.Execute(ctx, req.GetAttachmentId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetDownloadURLResponse{
		DownloadUrl: downloadURL,
	}, nil
}

// ListAttachments возвращает список вложений карточки
func (s *BoardServiceServer) ListAttachments(ctx context.Context, req *boardpb.ListAttachmentsRequest) (*boardpb.ListAttachmentsResponse, error) {
	if req.GetCardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "card_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	attachments, err := s.attachments.list.Execute(ctx, req.GetCardId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListAttachmentsResponse{
		Attachments: mapAttachmentsToProto(attachments),
	}, nil
}

// DeleteAttachment удаляет вложение
func (s *BoardServiceServer) DeleteAttachment(ctx context.Context, req *boardpb.DeleteAttachmentRequest) (*emptypb.Empty, error) {
	if req.GetAttachmentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "attachment_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.attachments.delete.Execute(ctx, req.GetAttachmentId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}
