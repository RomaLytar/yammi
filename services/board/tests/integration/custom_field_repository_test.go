package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestCustomFieldRepository_CreateDefinition(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create custom field definition
	def, err := domain.NewCustomFieldDefinition("", board.ID, "Sprint", domain.FieldTypeText, nil, 0, false)
	if err != nil {
		t.Fatalf("Failed to create domain definition: %v", err)
	}

	err = cfRepo.CreateDefinition(ctx, def)
	if err != nil {
		t.Fatalf("Failed to save definition: %v", err)
	}

	// Verify it exists
	loaded, err := cfRepo.GetDefinitionByID(ctx, def.ID)
	if err != nil {
		t.Fatalf("Failed to load definition: %v", err)
	}

	if loaded.Name != "Sprint" {
		t.Errorf("Expected name Sprint, got %s", loaded.Name)
	}

	if loaded.FieldType != domain.FieldTypeText {
		t.Errorf("Expected field type text, got %s", loaded.FieldType)
	}

	if loaded.BoardID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.BoardID)
	}
}

func TestCustomFieldRepository_CreateDefinition_Duplicate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	def1, _ := domain.NewCustomFieldDefinition("", board.ID, "Sprint", domain.FieldTypeText, nil, 0, false)
	cfRepo.CreateDefinition(ctx, def1)

	def2, _ := domain.NewCustomFieldDefinition("", board.ID, "Sprint", domain.FieldTypeNumber, nil, 1, false)
	err := cfRepo.CreateDefinition(ctx, def2)

	if err != domain.ErrCustomFieldExists {
		t.Errorf("Expected ErrCustomFieldExists, got %v", err)
	}
}

func TestCustomFieldRepository_CreateDefinition_Dropdown(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	options := []string{"Small", "Medium", "Large"}
	def, _ := domain.NewCustomFieldDefinition("", board.ID, "T-Shirt Size", domain.FieldTypeDropdown, options, 0, false)
	err := cfRepo.CreateDefinition(ctx, def)
	if err != nil {
		t.Fatalf("Failed to save dropdown definition: %v", err)
	}

	loaded, _ := cfRepo.GetDefinitionByID(ctx, def.ID)
	if len(loaded.Options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(loaded.Options))
	}
	if loaded.Options[1] != "Medium" {
		t.Errorf("Expected option[1] Medium, got %s", loaded.Options[1])
	}
}

func TestCustomFieldRepository_ListDefinitionsByBoardID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	names := []string{"Sprint", "Story Points", "Start Date"}
	for i, name := range names {
		ft := domain.FieldTypeText
		if i == 1 {
			ft = domain.FieldTypeNumber
		} else if i == 2 {
			ft = domain.FieldTypeDate
		}
		def, _ := domain.NewCustomFieldDefinition("", board.ID, name, ft, nil, i, false)
		cfRepo.CreateDefinition(ctx, def)
	}

	defs, err := cfRepo.ListDefinitionsByBoardID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to list definitions: %v", err)
	}

	if len(defs) != 3 {
		t.Errorf("Expected 3 definitions, got %d", len(defs))
	}
}

func TestCustomFieldRepository_UpdateDefinition(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	def, _ := domain.NewCustomFieldDefinition("", board.ID, "Sprint", domain.FieldTypeText, nil, 0, false)
	cfRepo.CreateDefinition(ctx, def)

	def.Update("Sprint Number", nil, true)
	err := cfRepo.UpdateDefinition(ctx, def)
	if err != nil {
		t.Fatalf("Failed to update definition: %v", err)
	}

	loaded, _ := cfRepo.GetDefinitionByID(ctx, def.ID)
	if loaded.Name != "Sprint Number" {
		t.Errorf("Expected name Sprint Number, got %s", loaded.Name)
	}
	if !loaded.Required {
		t.Error("Expected required to be true")
	}
}

func TestCustomFieldRepository_DeleteDefinition(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	def, _ := domain.NewCustomFieldDefinition("", board.ID, "Sprint", domain.FieldTypeText, nil, 0, false)
	cfRepo.CreateDefinition(ctx, def)

	err := cfRepo.DeleteDefinition(ctx, def.ID)
	if err != nil {
		t.Fatalf("Failed to delete definition: %v", err)
	}

	_, err = cfRepo.GetDefinitionByID(ctx, def.ID)
	if err != domain.ErrCustomFieldNotFound {
		t.Errorf("Expected ErrCustomFieldNotFound after delete, got %v", err)
	}
}

func TestCustomFieldRepository_CountDefinitionsByBoardID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	count, _ := cfRepo.CountDefinitionsByBoardID(ctx, board.ID)
	if count != 0 {
		t.Errorf("Expected 0 definitions, got %d", count)
	}

	for i := 0; i < 3; i++ {
		def, _ := domain.NewCustomFieldDefinition("", board.ID, uuid.NewString(), domain.FieldTypeText, nil, i, false)
		cfRepo.CreateDefinition(ctx, def)
	}

	count, _ = cfRepo.CountDefinitionsByBoardID(ctx, board.ID)
	if count != 3 {
		t.Errorf("Expected 3 definitions, got %d", count)
	}
}

func TestCustomFieldRepository_SetAndGetValue(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	// Create board, column, card
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	// Create field definition
	def, _ := domain.NewCustomFieldDefinition("", board.ID, "Sprint", domain.FieldTypeText, nil, 0, false)
	cfRepo.CreateDefinition(ctx, def)

	// Set value
	value := domain.NewCustomFieldValue("", card.ID, board.ID, def.ID)
	value.SetText("Sprint 42")
	err := cfRepo.SetValue(ctx, value)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Get values
	values, err := cfRepo.GetCardValues(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to get values: %v", err)
	}

	if len(values) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(values))
	}

	if values[0].ValueText == nil || *values[0].ValueText != "Sprint 42" {
		t.Errorf("Expected value text Sprint 42, got %v", values[0].ValueText)
	}
}

func TestCustomFieldRepository_SetValue_Upsert(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	def, _ := domain.NewCustomFieldDefinition("", board.ID, "Sprint", domain.FieldTypeText, nil, 0, false)
	cfRepo.CreateDefinition(ctx, def)

	// Set value first time
	v1 := domain.NewCustomFieldValue("", card.ID, board.ID, def.ID)
	v1.SetText("Sprint 1")
	cfRepo.SetValue(ctx, v1)

	// Set value second time (upsert)
	v2 := domain.NewCustomFieldValue("", card.ID, board.ID, def.ID)
	v2.SetText("Sprint 2")
	err := cfRepo.SetValue(ctx, v2)
	if err != nil {
		t.Fatalf("Failed to upsert value: %v", err)
	}

	// Should still have 1 value (upserted)
	values, _ := cfRepo.GetCardValues(ctx, card.ID, board.ID)
	if len(values) != 1 {
		t.Errorf("Expected 1 value after upsert, got %d", len(values))
	}

	if values[0].ValueText == nil || *values[0].ValueText != "Sprint 2" {
		t.Errorf("Expected upserted value Sprint 2, got %v", values[0].ValueText)
	}
}

func TestCustomFieldRepository_DeleteValue(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	def, _ := domain.NewCustomFieldDefinition("", board.ID, "Sprint", domain.FieldTypeText, nil, 0, false)
	cfRepo.CreateDefinition(ctx, def)

	value := domain.NewCustomFieldValue("", card.ID, board.ID, def.ID)
	value.SetText("Sprint 42")
	cfRepo.SetValue(ctx, value)

	err := cfRepo.DeleteValue(ctx, card.ID, board.ID, def.ID)
	if err != nil {
		t.Fatalf("Failed to delete value: %v", err)
	}

	values, _ := cfRepo.GetCardValues(ctx, card.ID, board.ID)
	if len(values) != 0 {
		t.Errorf("Expected 0 values after delete, got %d", len(values))
	}
}

func TestCustomFieldRepository_CascadeDelete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cfRepo := postgres.NewCustomFieldRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	def, _ := domain.NewCustomFieldDefinition("", board.ID, "Sprint", domain.FieldTypeText, nil, 0, false)
	cfRepo.CreateDefinition(ctx, def)

	value := domain.NewCustomFieldValue("", card.ID, board.ID, def.ID)
	value.SetText("Sprint 42")
	cfRepo.SetValue(ctx, value)

	// Delete definition — should cascade to values
	cfRepo.DeleteDefinition(ctx, def.ID)

	values, _ := cfRepo.GetCardValues(ctx, card.ID, board.ID)
	if len(values) != 0 {
		t.Errorf("Expected 0 values after cascade delete, got %d", len(values))
	}
}
