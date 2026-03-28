package usecase

import (
	"context"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mocks
// ============================================================================

// MockBoardTemplateRepository - мок для BoardTemplateRepository
type MockBoardTemplateRepository struct {
	mock.Mock
}

func (m *MockBoardTemplateRepository) Create(ctx context.Context, t *domain.BoardTemplate) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockBoardTemplateRepository) GetByID(ctx context.Context, id string) (*domain.BoardTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BoardTemplate), args.Error(1)
}

func (m *MockBoardTemplateRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.BoardTemplate, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BoardTemplate), args.Error(1)
}

func (m *MockBoardTemplateRepository) Update(ctx context.Context, t *domain.BoardTemplate) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockBoardTemplateRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ============================================================================
// CreateBoardTemplate Tests
// ============================================================================

func TestCreateBoardTemplate_Success(t *testing.T) {
	boardTmplRepo := new(MockBoardTemplateRepository)

	boardTmplRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.BoardTemplate")).Return(nil)

	uc := NewCreateBoardTemplateUseCase(boardTmplRepo)
	tmpl, err := uc.Execute(context.Background(), "user-123", "Project Board", "desc",
		[]domain.BoardColumnTemplateData{{Title: "To Do", Position: 0}},
		[]domain.LabelTemplateData{{Name: "Bug", Color: "#ef4444"}})

	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
	assert.Equal(t, "Project Board", tmpl.Name)
	assert.Equal(t, "user-123", tmpl.UserID)

	boardTmplRepo.AssertExpectations(t)
}

func TestCreateBoardTemplate_EmptyName(t *testing.T) {
	boardTmplRepo := new(MockBoardTemplateRepository)

	uc := NewCreateBoardTemplateUseCase(boardTmplRepo)
	tmpl, err := uc.Execute(context.Background(), "user-123", "", "desc", nil, nil)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyTemplateName, err)
	assert.Nil(t, tmpl)
}

// ============================================================================
// ListBoardTemplates Tests
// ============================================================================

func TestListBoardTemplates_Success(t *testing.T) {
	boardTmplRepo := new(MockBoardTemplateRepository)

	boardTmplRepo.On("ListByUserID", mock.Anything, "user-123").
		Return([]*domain.BoardTemplate{
			{ID: "tmpl-1", UserID: "user-123", Name: "Project"},
		}, nil)

	uc := NewListBoardTemplatesUseCase(boardTmplRepo)
	templates, err := uc.Execute(context.Background(), "user-123")

	assert.NoError(t, err)
	assert.Len(t, templates, 1)

	boardTmplRepo.AssertExpectations(t)
}

// ============================================================================
// DeleteBoardTemplate Tests
// ============================================================================

func TestDeleteBoardTemplate_OwnerCanDelete(t *testing.T) {
	boardTmplRepo := new(MockBoardTemplateRepository)

	boardTmplRepo.On("GetByID", mock.Anything, "tmpl-123").
		Return(&domain.BoardTemplate{ID: "tmpl-123", UserID: "user-123"}, nil)
	boardTmplRepo.On("Delete", mock.Anything, "tmpl-123").Return(nil)

	uc := NewDeleteBoardTemplateUseCase(boardTmplRepo)
	err := uc.Execute(context.Background(), "tmpl-123", "user-123")

	assert.NoError(t, err)

	boardTmplRepo.AssertExpectations(t)
}

func TestDeleteBoardTemplate_NotOwner_Denied(t *testing.T) {
	boardTmplRepo := new(MockBoardTemplateRepository)

	boardTmplRepo.On("GetByID", mock.Anything, "tmpl-123").
		Return(&domain.BoardTemplate{ID: "tmpl-123", UserID: "user-123"}, nil)

	uc := NewDeleteBoardTemplateUseCase(boardTmplRepo)
	err := uc.Execute(context.Background(), "tmpl-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)

	boardTmplRepo.AssertExpectations(t)
}

// ============================================================================
// CreateBoardFromTemplate Tests
// ============================================================================

func TestCreateBoardFromTemplate_Success(t *testing.T) {
	boardTmplRepo := new(MockBoardTemplateRepository)
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)
	columnRepo := new(MockColumnRepository)
	labelRepo := new(MockLabelRepository)
	publisher := new(MockEventPublisher)

	boardTmplRepo.On("GetByID", mock.Anything, "tmpl-123").
		Return(&domain.BoardTemplate{
			ID:          "tmpl-123",
			UserID:      "user-123",
			Name:        "Project Board",
			Description: "Standard project board",
			ColumnsData: []domain.BoardColumnTemplateData{
				{Title: "To Do", Position: 0},
				{Title: "Done", Position: 1},
			},
			LabelsData: []domain.LabelTemplateData{
				{Name: "Bug", Color: "#ef4444"},
			},
		}, nil)
	boardRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Board")).Return(nil)
	columnRepo.On("BatchCreate", mock.Anything, mock.AnythingOfType("[]*domain.Column")).Return(nil)
	labelRepo.On("BatchCreate", mock.Anything, mock.AnythingOfType("[]*domain.Label")).Return(nil)
	publisher.On("PublishBoardCreated", mock.Anything, mock.Anything).Return(nil).Maybe()
	publisher.On("PublishMemberAdded", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewCreateBoardFromTemplateUseCase(boardTmplRepo, boardRepo, memberRepo, columnRepo, labelRepo, publisher)
	board, err := uc.Execute(context.Background(), "tmpl-123", "My Board", "user-123")

	assert.NoError(t, err)
	assert.NotNil(t, board)
	assert.Equal(t, "My Board", board.Title)
	assert.Equal(t, "Standard project board", board.Description)
	assert.Equal(t, "user-123", board.OwnerID)

	boardTmplRepo.AssertExpectations(t)
	boardRepo.AssertExpectations(t)
	columnRepo.AssertExpectations(t)
	labelRepo.AssertExpectations(t)
}

func TestCreateBoardFromTemplate_TemplateNotFound(t *testing.T) {
	boardTmplRepo := new(MockBoardTemplateRepository)
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)
	columnRepo := new(MockColumnRepository)
	labelRepo := new(MockLabelRepository)
	publisher := new(MockEventPublisher)

	boardTmplRepo.On("GetByID", mock.Anything, "tmpl-999").
		Return(nil, domain.ErrTemplateNotFound)

	uc := NewCreateBoardFromTemplateUseCase(boardTmplRepo, boardRepo, memberRepo, columnRepo, labelRepo, publisher)
	board, err := uc.Execute(context.Background(), "tmpl-999", "Board", "user-123")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrTemplateNotFound, err)
	assert.Nil(t, board)

	boardTmplRepo.AssertExpectations(t)
}
