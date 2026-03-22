package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListAttachmentsUseCase struct {
	attachmentRepo AttachmentRepository
	memberRepo     MembershipRepository
}

func NewListAttachmentsUseCase(attachmentRepo AttachmentRepository, memberRepo MembershipRepository) *ListAttachmentsUseCase {
	return &ListAttachmentsUseCase{
		attachmentRepo: attachmentRepo,
		memberRepo:     memberRepo,
	}
}

// Execute возвращает список вложений карточки
func (uc *ListAttachmentsUseCase) Execute(ctx context.Context, cardID, boardID, userID string) ([]*domain.Attachment, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем вложения
	return uc.attachmentRepo.ListByCardID(ctx, cardID, boardID)
}
