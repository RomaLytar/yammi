package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type GetUnreadCountUseCase struct {
	unreadCounter UnreadCounter
}

func NewGetUnreadCountUseCase(unreadCounter UnreadCounter) *GetUnreadCountUseCase {
	return &GetUnreadCountUseCase{unreadCounter: unreadCounter}
}

func (uc *GetUnreadCountUseCase) Execute(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, domain.ErrEmptyUserID
	}
	return uc.unreadCounter.Get(ctx, userID)
}
