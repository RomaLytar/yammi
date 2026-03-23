package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCustomFieldRepository - мок для CustomFieldRepository
type MockCustomFieldRepository struct {
	mock.Mock
}

func (m *MockCustomFieldRepository) CreateDefinition(ctx context.Context, def *domain.CustomFieldDefinition) error {
	args := m.Called(ctx, def)
	return args.Error(0)
}

func (m *MockCustomFieldRepository) GetDefinitionByID(ctx context.Context, defID string) (*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, defID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomFieldDefinition), args.Error(1)
}

func (m *MockCustomFieldRepository) ListDefinitionsByBoardID(ctx context.Context, boardID string) ([]*domain.CustomFieldDefinition, error) {
	args := m.Called(ctx, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CustomFieldDefinition), args.Error(1)
}

func (m *MockCustomFieldRepository) UpdateDefinition(ctx context.Context, def *domain.CustomFieldDefinition) error {
	args := m.Called(ctx, def)
	return args.Error(0)
}

func (m *MockCustomFieldRepository) DeleteDefinition(ctx context.Context, defID string) error {
	args := m.Called(ctx, defID)
	return args.Error(0)
}

func (m *MockCustomFieldRepository) CountDefinitionsByBoardID(ctx context.Context, boardID string) (int, error) {
	args := m.Called(ctx, boardID)
	return args.Int(0), args.Error(1)
}

func (m *MockCustomFieldRepository) SetValue(ctx context.Context, value *domain.CustomFieldValue) error {
	args := m.Called(ctx, value)
	return args.Error(0)
}

func (m *MockCustomFieldRepository) GetCardValues(ctx context.Context, cardID, boardID string) ([]*domain.CustomFieldValue, error) {
	args := m.Called(ctx, cardID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.CustomFieldValue), args.Error(1)
}

func (m *MockCustomFieldRepository) DeleteValue(ctx context.Context, cardID, boardID, fieldID string) error {
	args := m.Called(ctx, cardID, boardID, fieldID)
	return args.Error(0)
}

func TestCreateCustomField_Success(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	cfRepo.On("CountDefinitionsByBoardID", mock.Anything, "board-123").Return(5, nil)
	cfRepo.On("CreateDefinition", mock.Anything, mock.AnythingOfType("*domain.CustomFieldDefinition")).Return(nil)
	publisher.On("PublishCustomFieldCreated", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewCreateCustomFieldUseCase(cfRepo, memberRepo, publisher)
	def, err := uc.Execute(context.Background(), "board-123", "user-123", "Sprint", domain.FieldTypeText, nil, 0, false)

	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.Equal(t, "Sprint", def.Name)
	assert.Equal(t, domain.FieldTypeText, def.FieldType)
	assert.Equal(t, "board-123", def.BoardID)
	assert.NotEmpty(t, def.ID)

	cfRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestCreateCustomField_NonMember_Denied(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewCreateCustomFieldUseCase(cfRepo, memberRepo, publisher)
	def, err := uc.Execute(context.Background(), "board-123", "user-999", "Sprint", domain.FieldTypeText, nil, 0, false)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, def)

	memberRepo.AssertExpectations(t)
}

func TestCreateCustomField_NotOwner_Denied(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").
		Return(true, domain.RoleMember, nil)

	uc := NewCreateCustomFieldUseCase(cfRepo, memberRepo, publisher)
	def, err := uc.Execute(context.Background(), "board-123", "user-456", "Sprint", domain.FieldTypeText, nil, 0, false)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotOwner, err)
	assert.Nil(t, def)

	memberRepo.AssertExpectations(t)
}

func TestCreateCustomField_EmptyName(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	cfRepo.On("CountDefinitionsByBoardID", mock.Anything, "board-123").Return(5, nil)

	uc := NewCreateCustomFieldUseCase(cfRepo, memberRepo, publisher)
	def, err := uc.Execute(context.Background(), "board-123", "user-123", "", domain.FieldTypeText, nil, 0, false)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyFieldName, err)
	assert.Nil(t, def)

	memberRepo.AssertExpectations(t)
}

func TestCreateCustomField_InvalidType(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	cfRepo.On("CountDefinitionsByBoardID", mock.Anything, "board-123").Return(5, nil)

	uc := NewCreateCustomFieldUseCase(cfRepo, memberRepo, publisher)
	def, err := uc.Execute(context.Background(), "board-123", "user-123", "Field", domain.FieldType("invalid"), nil, 0, false)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidFieldType, err)
	assert.Nil(t, def)

	memberRepo.AssertExpectations(t)
}

func TestCreateCustomField_MaxReached(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	cfRepo.On("CountDefinitionsByBoardID", mock.Anything, "board-123").Return(30, nil)

	uc := NewCreateCustomFieldUseCase(cfRepo, memberRepo, publisher)
	def, err := uc.Execute(context.Background(), "board-123", "user-123", "Sprint", domain.FieldTypeText, nil, 0, false)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrMaxCustomFieldsReached, err)
	assert.Nil(t, def)

	memberRepo.AssertExpectations(t)
	cfRepo.AssertExpectations(t)
}

func TestSetCustomFieldValue_Success(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cfRepo.On("GetDefinitionByID", mock.Anything, "field-123").
		Return(&domain.CustomFieldDefinition{
			ID:        "field-123",
			BoardID:   "board-123",
			Name:      "Sprint",
			FieldType: domain.FieldTypeText,
		}, nil)
	cfRepo.On("SetValue", mock.Anything, mock.AnythingOfType("*domain.CustomFieldValue")).Return(nil)
	publisher.On("PublishCustomFieldValueSet", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewSetCustomFieldValueUseCase(cfRepo, memberRepo, publisher)
	text := "Sprint 42"
	value, err := uc.Execute(context.Background(), "card-123", "board-123", "field-123", "user-123", &text, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, value)
	assert.NotNil(t, value.ValueText)
	assert.Equal(t, "Sprint 42", *value.ValueText)

	cfRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestSetCustomFieldValue_Number(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cfRepo.On("GetDefinitionByID", mock.Anything, "field-123").
		Return(&domain.CustomFieldDefinition{
			ID:        "field-123",
			BoardID:   "board-123",
			Name:      "Story Points",
			FieldType: domain.FieldTypeNumber,
		}, nil)
	cfRepo.On("SetValue", mock.Anything, mock.AnythingOfType("*domain.CustomFieldValue")).Return(nil)
	publisher.On("PublishCustomFieldValueSet", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewSetCustomFieldValueUseCase(cfRepo, memberRepo, publisher)
	num := 8.0
	value, err := uc.Execute(context.Background(), "card-123", "board-123", "field-123", "user-123", nil, &num, nil)

	assert.NoError(t, err)
	assert.NotNil(t, value)
	assert.NotNil(t, value.ValueNumber)
	assert.Equal(t, 8.0, *value.ValueNumber)

	cfRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestSetCustomFieldValue_Date(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cfRepo.On("GetDefinitionByID", mock.Anything, "field-123").
		Return(&domain.CustomFieldDefinition{
			ID:        "field-123",
			BoardID:   "board-123",
			Name:      "Start Date",
			FieldType: domain.FieldTypeDate,
		}, nil)
	cfRepo.On("SetValue", mock.Anything, mock.AnythingOfType("*domain.CustomFieldValue")).Return(nil)
	publisher.On("PublishCustomFieldValueSet", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewSetCustomFieldValueUseCase(cfRepo, memberRepo, publisher)
	dt := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	value, err := uc.Execute(context.Background(), "card-123", "board-123", "field-123", "user-123", nil, nil, &dt)

	assert.NoError(t, err)
	assert.NotNil(t, value)
	assert.NotNil(t, value.ValueDate)

	cfRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestSetCustomFieldValue_Dropdown(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cfRepo.On("GetDefinitionByID", mock.Anything, "field-123").
		Return(&domain.CustomFieldDefinition{
			ID:        "field-123",
			BoardID:   "board-123",
			Name:      "Size",
			FieldType: domain.FieldTypeDropdown,
			Options:   []string{"Small", "Medium", "Large"},
		}, nil)
	cfRepo.On("SetValue", mock.Anything, mock.AnythingOfType("*domain.CustomFieldValue")).Return(nil)
	publisher.On("PublishCustomFieldValueSet", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewSetCustomFieldValueUseCase(cfRepo, memberRepo, publisher)
	text := "Medium"
	value, err := uc.Execute(context.Background(), "card-123", "board-123", "field-123", "user-123", &text, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, value)
	assert.NotNil(t, value.ValueText)
	assert.Equal(t, "Medium", *value.ValueText)

	cfRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestSetCustomFieldValue_Dropdown_InvalidOption(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	cfRepo.On("GetDefinitionByID", mock.Anything, "field-123").
		Return(&domain.CustomFieldDefinition{
			ID:        "field-123",
			BoardID:   "board-123",
			Name:      "Size",
			FieldType: domain.FieldTypeDropdown,
			Options:   []string{"Small", "Medium", "Large"},
		}, nil)

	uc := NewSetCustomFieldValueUseCase(cfRepo, memberRepo, publisher)
	text := "ExtraLarge"
	value, err := uc.Execute(context.Background(), "card-123", "board-123", "field-123", "user-123", &text, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidFieldValue, err)
	assert.Nil(t, value)

	cfRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestSetCustomFieldValue_NonMember_Denied(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewSetCustomFieldValueUseCase(cfRepo, memberRepo, publisher)
	text := "value"
	value, err := uc.Execute(context.Background(), "card-123", "board-123", "field-123", "user-999", &text, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, value)

	memberRepo.AssertExpectations(t)
}

func TestDeleteCustomField_Owner(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	cfRepo.On("DeleteDefinition", mock.Anything, "field-123").Return(nil)
	publisher.On("PublishCustomFieldDeleted", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewDeleteCustomFieldUseCase(cfRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "field-123", "board-123", "user-123")

	assert.NoError(t, err)

	cfRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestDeleteCustomField_Member_Denied(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").
		Return(true, domain.RoleMember, nil)

	uc := NewDeleteCustomFieldUseCase(cfRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "field-123", "board-123", "user-456")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotOwner, err)

	memberRepo.AssertExpectations(t)
}

func TestGetCardCustomFields_Success(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)

	text := "Sprint 42"
	cfRepo.On("GetCardValues", mock.Anything, "card-123", "board-123").
		Return([]*domain.CustomFieldValue{
			{ID: "val-1", CardID: "card-123", BoardID: "board-123", FieldID: "field-1", ValueText: &text},
		}, nil)

	uc := NewGetCardCustomFieldsUseCase(cfRepo, memberRepo)
	values, err := uc.Execute(context.Background(), "card-123", "board-123", "user-123")

	assert.NoError(t, err)
	assert.Len(t, values, 1)
	assert.Equal(t, "Sprint 42", *values[0].ValueText)

	cfRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetCardCustomFields_NonMember_Denied(t *testing.T) {
	cfRepo := new(MockCustomFieldRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewGetCardCustomFieldsUseCase(cfRepo, memberRepo)
	values, err := uc.Execute(context.Background(), "card-123", "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, values)

	memberRepo.AssertExpectations(t)
}
