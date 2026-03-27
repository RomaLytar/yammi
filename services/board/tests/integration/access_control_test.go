package integration

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// ---------- helper: create board with owner + optional member ----------

type testBoard struct {
	board      *domain.Board
	ownerID    string
	memberID   string
	outsiderID string
	column     *domain.Column
	card       *domain.Card
}

// setupBoard creates a board with owner (auto-member), a member, and an outsider.
// Optionally creates a column and card for tests that need them.
func setupBoard(t *testing.T, db *sql.DB, withCard bool) testBoard {
	t.Helper()
	ctx := context.Background()

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)

	ownerID := uuid.NewString()
	memberID := uuid.NewString()
	outsiderID := uuid.NewString()

	board, err := domain.NewBoard("Board-"+uuid.NewString()[:8], "test", ownerID)
	if err != nil {
		t.Fatalf("NewBoard: %v", err)
	}
	if err := boardRepo.Create(ctx, board); err != nil {
		t.Fatalf("Create board: %v", err)
	}
	if err := memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	tb := testBoard{
		board:      board,
		ownerID:    ownerID,
		memberID:   memberID,
		outsiderID: outsiderID,
	}

	if withCard {
		col, err := domain.NewColumn(board.ID, "Col", 0)
		if err != nil {
			t.Fatalf("NewColumn: %v", err)
		}
		if err := columnRepo.Create(ctx, col); err != nil {
			t.Fatalf("Create column: %v", err)
		}
		card, err := domain.NewCard(col.ID, "Card", "desc", "n", nil, memberID, nil, "", "")
		if err != nil {
			t.Fatalf("NewCard: %v", err)
		}
		if err := cardRepo.Create(ctx, card); err != nil {
			t.Fatalf("Create card: %v", err)
		}
		tb.column = col
		tb.card = card
	}

	return tb
}

// ==================== LABELS ====================

func TestACL_Label_MemberCanCreate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewCreateLabelUseCase(postgres.NewLabelRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	label, err := uc.Execute(ctx, tb.board.ID, tb.memberID, "Bug", "#FF0000")
	if err != nil {
		t.Fatalf("Member should create label: %v", err)
	}
	if label.Name != "Bug" {
		t.Errorf("Expected name 'Bug', got %s", label.Name)
	}
}

func TestACL_Label_NonMemberCannotCreate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewCreateLabelUseCase(postgres.NewLabelRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	_, err := uc.Execute(ctx, tb.board.ID, tb.outsiderID, "Bug", "#FF0000")
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

func TestACL_Label_OwnerCanDelete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	labelRepo := postgres.NewLabelRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	// Create label first
	label, _ := usecase.NewCreateLabelUseCase(labelRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "ToDelete", "#000000")

	// Owner deletes
	err := usecase.NewDeleteLabelUseCase(labelRepo, memberRepo, pub).Execute(ctx, label.ID, tb.board.ID, tb.ownerID)
	if err != nil {
		t.Fatalf("Owner should delete label: %v", err)
	}
}

func TestACL_Label_MemberCannotDelete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	labelRepo := postgres.NewLabelRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	label, _ := usecase.NewCreateLabelUseCase(labelRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "Protected", "#111111")

	err := usecase.NewDeleteLabelUseCase(labelRepo, memberRepo, pub).Execute(ctx, label.ID, tb.board.ID, tb.memberID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for member delete, got %v", err)
	}
}

func TestACL_Label_NonMemberCannotDelete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	labelRepo := postgres.NewLabelRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	label, _ := usecase.NewCreateLabelUseCase(labelRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "Secret", "#222222")

	err := usecase.NewDeleteLabelUseCase(labelRepo, memberRepo, pub).Execute(ctx, label.ID, tb.board.ID, tb.outsiderID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for outsider delete, got %v", err)
	}
}

func TestACL_Label_MemberCanAttachToCard(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	labelRepo := postgres.NewLabelRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	label, _ := usecase.NewCreateLabelUseCase(labelRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.memberID, "Attach", "#333333")

	err := usecase.NewAddLabelToCardUseCase(labelRepo, nil, memberRepo, pub).Execute(ctx, tb.card.ID, tb.board.ID, label.ID, tb.memberID)
	if err != nil {
		t.Fatalf("Member should attach label to card: %v", err)
	}

	// Verify
	labels, err := usecase.NewGetCardLabelsUseCase(labelRepo, memberRepo).Execute(ctx, tb.card.ID, tb.board.ID, tb.memberID)
	if err != nil {
		t.Fatalf("Failed to get card labels: %v", err)
	}
	if len(labels) != 1 || labels[0].ID != label.ID {
		t.Errorf("Expected 1 label attached, got %d", len(labels))
	}
}

func TestACL_Label_NonMemberCannotAttach(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	labelRepo := postgres.NewLabelRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	label, _ := usecase.NewCreateLabelUseCase(labelRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "NoAccess", "#444444")

	err := usecase.NewAddLabelToCardUseCase(labelRepo, nil, memberRepo, pub).Execute(ctx, tb.card.ID, tb.board.ID, label.ID, tb.outsiderID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for outsider attach, got %v", err)
	}
}

func TestACL_Label_NonMemberCannotList(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewListLabelsUseCase(postgres.NewLabelRepository(db), postgres.NewMembershipRepository(db))
	_, err := uc.Execute(ctx, tb.board.ID, tb.outsiderID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for outsider list, got %v", err)
	}
}

func TestACL_Label_MemberCanDetach(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	labelRepo := postgres.NewLabelRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	label, _ := usecase.NewCreateLabelUseCase(labelRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.memberID, "Detach", "#555555")
	_ = usecase.NewAddLabelToCardUseCase(labelRepo, nil, memberRepo, pub).Execute(ctx, tb.card.ID, tb.board.ID, label.ID, tb.memberID)

	err := usecase.NewRemoveLabelFromCardUseCase(labelRepo, memberRepo, pub).Execute(ctx, tb.card.ID, tb.board.ID, label.ID, tb.memberID)
	if err != nil {
		t.Fatalf("Member should detach label: %v", err)
	}
}

// ==================== CHECKLISTS ====================

func TestACL_Checklist_MemberCanCreate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	uc := usecase.NewCreateChecklistUseCase(postgres.NewChecklistRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	cl, err := uc.Execute(ctx, tb.card.ID, tb.board.ID, tb.memberID, "Checklist 1", 0)
	if err != nil {
		t.Fatalf("Member should create checklist: %v", err)
	}
	if cl.Title != "Checklist 1" {
		t.Errorf("Expected title 'Checklist 1', got %s", cl.Title)
	}
}

func TestACL_Checklist_NonMemberCannotCreate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	uc := usecase.NewCreateChecklistUseCase(postgres.NewChecklistRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	_, err := uc.Execute(ctx, tb.card.ID, tb.board.ID, tb.outsiderID, "Hacked", 0)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

func TestACL_ChecklistItem_MemberCanCreateAndToggle(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	checklistRepo := postgres.NewChecklistRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	cl, _ := usecase.NewCreateChecklistUseCase(checklistRepo, memberRepo, pub).Execute(ctx, tb.card.ID, tb.board.ID, tb.memberID, "CL", 0)
	item, err := usecase.NewCreateChecklistItemUseCase(checklistRepo, memberRepo).Execute(ctx, cl.ID, tb.board.ID, tb.memberID, "Task A", 0)
	if err != nil {
		t.Fatalf("Member should create item: %v", err)
	}

	isChecked, err := usecase.NewToggleChecklistItemUseCase(checklistRepo, memberRepo, pub).Execute(ctx, item.ID, tb.board.ID, tb.memberID)
	if err != nil {
		t.Fatalf("Member should toggle item: %v", err)
	}
	if !isChecked {
		t.Error("Expected item to be checked after toggle")
	}
}

func TestACL_ChecklistItem_NonMemberCannotToggle(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	checklistRepo := postgres.NewChecklistRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	cl, _ := usecase.NewCreateChecklistUseCase(checklistRepo, memberRepo, pub).Execute(ctx, tb.card.ID, tb.board.ID, tb.memberID, "CL2", 0)
	item, _ := usecase.NewCreateChecklistItemUseCase(checklistRepo, memberRepo).Execute(ctx, cl.ID, tb.board.ID, tb.memberID, "Task B", 0)

	_, err := usecase.NewToggleChecklistItemUseCase(checklistRepo, memberRepo, pub).Execute(ctx, item.ID, tb.board.ID, tb.outsiderID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

// ==================== CARD LINKS ====================

func TestACL_CardLink_MemberCanLink(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	// Create a second card in the same board
	cardRepo := postgres.NewCardRepository(db)
	child, _ := domain.NewCard(tb.column.ID, "Child", "desc", "o", nil, tb.memberID, nil, "", "")
	_ = cardRepo.Create(ctx, child)

	uc := usecase.NewLinkCardsUseCase(
		postgres.NewCardLinkRepository(db), cardRepo,
		postgres.NewMembershipRepository(db), &mockPublisher{},
	)
	link, err := uc.Execute(ctx, tb.card.ID, child.ID, tb.board.ID, tb.memberID)
	if err != nil {
		t.Fatalf("Member should link cards: %v", err)
	}
	if link.ParentID != tb.card.ID || link.ChildID != child.ID {
		t.Error("Link parent/child mismatch")
	}
}

func TestACL_CardLink_NonMemberCannotLink(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	cardRepo := postgres.NewCardRepository(db)
	child, _ := domain.NewCard(tb.column.ID, "Child2", "desc", "p", nil, tb.memberID, nil, "", "")
	_ = cardRepo.Create(ctx, child)

	uc := usecase.NewLinkCardsUseCase(
		postgres.NewCardLinkRepository(db), cardRepo,
		postgres.NewMembershipRepository(db), &mockPublisher{},
	)
	_, err := uc.Execute(ctx, tb.card.ID, child.ID, tb.board.ID, tb.outsiderID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

func TestACL_CardLink_MemberCanViewChildren(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	cardRepo := postgres.NewCardRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	child, _ := domain.NewCard(tb.column.ID, "ViewChild", "desc", "q", nil, tb.memberID, nil, "", "")
	_ = cardRepo.Create(ctx, child)
	_, _ = usecase.NewLinkCardsUseCase(cardLinkRepo, cardRepo, memberRepo, pub).Execute(ctx, tb.card.ID, child.ID, tb.board.ID, tb.memberID)

	links, err := usecase.NewGetCardChildrenUseCase(cardLinkRepo, memberRepo).Execute(ctx, tb.card.ID, tb.board.ID, tb.memberID)
	if err != nil {
		t.Fatalf("Member should view children: %v", err)
	}
	if len(links) != 1 {
		t.Errorf("Expected 1 child, got %d", len(links))
	}
}

func TestACL_CardLink_NonMemberCannotViewChildren(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	uc := usecase.NewGetCardChildrenUseCase(postgres.NewCardLinkRepository(db), postgres.NewMembershipRepository(db))
	_, err := uc.Execute(ctx, tb.card.ID, tb.board.ID, tb.outsiderID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

// ==================== CUSTOM FIELDS ====================

func TestACL_CustomField_OwnerCanCreateDefinition(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewCreateCustomFieldUseCase(postgres.NewCustomFieldRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	def, err := uc.Execute(ctx, tb.board.ID, tb.ownerID, "Priority", domain.FieldTypeText, nil, 0, false)
	if err != nil {
		t.Fatalf("Owner should create custom field: %v", err)
	}
	if def.Name != "Priority" {
		t.Errorf("Expected name 'Priority', got %s", def.Name)
	}
}

func TestACL_CustomField_MemberCannotCreateDefinition(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewCreateCustomFieldUseCase(postgres.NewCustomFieldRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	_, err := uc.Execute(ctx, tb.board.ID, tb.memberID, "Blocked", domain.FieldTypeText, nil, 0, false)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for member, got %v", err)
	}
}

func TestACL_CustomField_NonMemberCannotCreateDefinition(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewCreateCustomFieldUseCase(postgres.NewCustomFieldRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	_, err := uc.Execute(ctx, tb.board.ID, tb.outsiderID, "Hacked", domain.FieldTypeText, nil, 0, false)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for outsider, got %v", err)
	}
}

func TestACL_CustomField_OwnerCanDelete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	cfRepo := postgres.NewCustomFieldRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	def, _ := usecase.NewCreateCustomFieldUseCase(cfRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "Del", domain.FieldTypeText, nil, 0, false)

	err := usecase.NewDeleteCustomFieldUseCase(cfRepo, memberRepo, pub).Execute(ctx, def.ID, tb.board.ID, tb.ownerID)
	if err != nil {
		t.Fatalf("Owner should delete custom field: %v", err)
	}
}

func TestACL_CustomField_MemberCannotDelete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	cfRepo := postgres.NewCustomFieldRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	def, _ := usecase.NewCreateCustomFieldUseCase(cfRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "NoDel", domain.FieldTypeText, nil, 0, false)

	err := usecase.NewDeleteCustomFieldUseCase(cfRepo, memberRepo, pub).Execute(ctx, def.ID, tb.board.ID, tb.memberID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for member delete, got %v", err)
	}
}

func TestACL_CustomField_MemberCanSetValue(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	cfRepo := postgres.NewCustomFieldRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	def, _ := usecase.NewCreateCustomFieldUseCase(cfRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "Status", domain.FieldTypeText, nil, 0, false)

	txt := "Active"
	val, err := usecase.NewSetCustomFieldValueUseCase(cfRepo, memberRepo, pub).Execute(ctx, tb.card.ID, tb.board.ID, def.ID, tb.memberID, &txt, nil, nil)
	if err != nil {
		t.Fatalf("Member should set value: %v", err)
	}
	if val.ValueText == nil || *val.ValueText != "Active" {
		t.Error("Expected value 'Active'")
	}
}

func TestACL_CustomField_NonMemberCannotSetValue(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, true)
	ctx := context.Background()

	cfRepo := postgres.NewCustomFieldRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	def, _ := usecase.NewCreateCustomFieldUseCase(cfRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "Locked", domain.FieldTypeText, nil, 0, false)

	txt := "Hacked"
	_, err := usecase.NewSetCustomFieldValueUseCase(cfRepo, memberRepo, pub).Execute(ctx, tb.card.ID, tb.board.ID, def.ID, tb.outsiderID, &txt, nil, nil)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

func TestACL_CustomField_MemberCanList(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	cfRepo := postgres.NewCustomFieldRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)

	defs, err := usecase.NewListCustomFieldsUseCase(cfRepo, memberRepo).Execute(ctx, tb.board.ID, tb.memberID)
	if err != nil {
		t.Fatalf("Member should list custom fields: %v", err)
	}
	_ = defs // empty is ok
}

func TestACL_CustomField_NonMemberCannotList(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewListCustomFieldsUseCase(postgres.NewCustomFieldRepository(db), postgres.NewMembershipRepository(db))
	_, err := uc.Execute(ctx, tb.board.ID, tb.outsiderID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

// ==================== AUTOMATION RULES ====================

func TestACL_Automation_OwnerCanCreate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewCreateAutomationRuleUseCase(postgres.NewAutomationRuleRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	rule, err := uc.Execute(ctx, tb.board.ID, tb.ownerID, "Auto-assign",
		domain.TriggerCardCreated, map[string]string{},
		domain.ActionSetPriority, map[string]string{"priority": "high"})
	if err != nil {
		t.Fatalf("Owner should create automation: %v", err)
	}
	if rule.Name != "Auto-assign" {
		t.Errorf("Expected name 'Auto-assign', got %s", rule.Name)
	}
}

func TestACL_Automation_MemberCannotCreate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewCreateAutomationRuleUseCase(postgres.NewAutomationRuleRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	_, err := uc.Execute(ctx, tb.board.ID, tb.memberID, "Blocked",
		domain.TriggerCardCreated, map[string]string{},
		domain.ActionSetPriority, map[string]string{"priority": "low"})
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for member, got %v", err)
	}
}

func TestACL_Automation_NonMemberCannotCreate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	uc := usecase.NewCreateAutomationRuleUseCase(postgres.NewAutomationRuleRepository(db), postgres.NewMembershipRepository(db), &mockPublisher{})
	_, err := uc.Execute(ctx, tb.board.ID, tb.outsiderID, "Hacked",
		domain.TriggerCardCreated, map[string]string{},
		domain.ActionSetPriority, map[string]string{"priority": "low"})
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

func TestACL_Automation_OwnerCanDelete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	ruleRepo := postgres.NewAutomationRuleRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	rule, _ := usecase.NewCreateAutomationRuleUseCase(ruleRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "Del",
		domain.TriggerCardCreated, map[string]string{}, domain.ActionSetPriority, map[string]string{"priority": "high"})

	err := usecase.NewDeleteAutomationRuleUseCase(ruleRepo, memberRepo, pub).Execute(ctx, rule.ID, tb.board.ID, tb.ownerID)
	if err != nil {
		t.Fatalf("Owner should delete automation: %v", err)
	}
}

func TestACL_Automation_MemberCannotDelete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	ruleRepo := postgres.NewAutomationRuleRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	pub := &mockPublisher{}

	rule, _ := usecase.NewCreateAutomationRuleUseCase(ruleRepo, memberRepo, pub).Execute(ctx, tb.board.ID, tb.ownerID, "NoDel",
		domain.TriggerCardCreated, map[string]string{}, domain.ActionSetPriority, map[string]string{"priority": "high"})

	err := usecase.NewDeleteAutomationRuleUseCase(ruleRepo, memberRepo, pub).Execute(ctx, rule.ID, tb.board.ID, tb.memberID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for member delete, got %v", err)
	}
}

func TestACL_Automation_MemberCanList(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	rules, err := usecase.NewListAutomationRulesUseCase(postgres.NewAutomationRuleRepository(db), postgres.NewMembershipRepository(db)).Execute(ctx, tb.board.ID, tb.memberID)
	if err != nil {
		t.Fatalf("Member should list automations: %v", err)
	}
	_ = rules
}

func TestACL_Automation_NonMemberCannotList(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	tb := setupBoard(t, db, false)
	ctx := context.Background()

	_, err := usecase.NewListAutomationRulesUseCase(postgres.NewAutomationRuleRepository(db), postgres.NewMembershipRepository(db)).Execute(ctx, tb.board.ID, tb.outsiderID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

// ==================== CROSS-BOARD ACCESS ====================

func TestACL_CrossBoard_MemberCannotAccessOtherBoard(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	ctx := context.Background()

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)

	// Board A with userA as owner
	userA := uuid.NewString()
	boardA, _ := domain.NewBoard("Board-A", "desc", userA)
	_ = boardRepo.Create(ctx, boardA)

	// Board B with userB as owner
	userB := uuid.NewString()
	boardB, _ := domain.NewBoard("Board-B", "desc", userB)
	_ = boardRepo.Create(ctx, boardB)

	// UserA tries to read Board B
	uc := usecase.NewGetBoardUseCase(boardRepo, memberRepo)
	_, err := uc.Execute(ctx, boardB.ID, userA)
	if err != domain.ErrAccessDenied {
		t.Errorf("UserA should not access BoardB, got %v", err)
	}

	// UserB tries to read Board A
	_, err = uc.Execute(ctx, boardA.ID, userB)
	if err != domain.ErrAccessDenied {
		t.Errorf("UserB should not access BoardA, got %v", err)
	}
}

func TestACL_CrossBoard_CannotCreateCardInOtherBoard(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	ctx := context.Background()

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	activityRepo := postgres.NewActivityRepository(db)
	pub := &mockPublisher{}

	// Board A
	userA := uuid.NewString()
	boardA, _ := domain.NewBoard("BoardA", "desc", userA)
	_ = boardRepo.Create(ctx, boardA)
	colA, _ := domain.NewColumn(boardA.ID, "ColA", 0)
	_ = columnRepo.Create(ctx, colA)

	// Board B
	userB := uuid.NewString()
	boardB, _ := domain.NewBoard("BoardB", "desc", userB)
	_ = boardRepo.Create(ctx, boardB)

	// UserB tries to create card in Board A
	uc := usecase.NewCreateCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, pub, nil)
	_, err := uc.Execute(ctx, colA.ID, boardA.ID, userB, "Hacked Card", "desc", "", nil, nil, "", "")
	if err != domain.ErrAccessDenied {
		t.Errorf("UserB should not create card in BoardA, got %v", err)
	}
}

func TestACL_CrossBoard_CannotCreateLabelInOtherBoard(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	ctx := context.Background()

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)

	userA := uuid.NewString()
	boardA, _ := domain.NewBoard("BA", "desc", userA)
	_ = boardRepo.Create(ctx, boardA)

	userB := uuid.NewString()
	boardB, _ := domain.NewBoard("BB", "desc", userB)
	_ = boardRepo.Create(ctx, boardB)

	// UserB tries to create label in Board A
	uc := usecase.NewCreateLabelUseCase(postgres.NewLabelRepository(db), memberRepo, &mockPublisher{})
	_, err := uc.Execute(ctx, boardA.ID, userB, "Cross", "#FF0000")
	if err != domain.ErrAccessDenied {
		t.Errorf("UserB should not create label in BoardA, got %v", err)
	}
}

func TestACL_CrossBoard_CannotCreateAutomationInOtherBoard(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)
	ctx := context.Background()

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)

	userA := uuid.NewString()
	boardA, _ := domain.NewBoard("BA2", "desc", userA)
	_ = boardRepo.Create(ctx, boardA)

	userB := uuid.NewString()
	boardB, _ := domain.NewBoard("BB2", "desc", userB)
	_ = boardRepo.Create(ctx, boardB)

	uc := usecase.NewCreateAutomationRuleUseCase(postgres.NewAutomationRuleRepository(db), memberRepo, &mockPublisher{})
	_, err := uc.Execute(ctx, boardA.ID, userB, "CrossAuto",
		domain.TriggerCardCreated, map[string]string{}, domain.ActionSetPriority, map[string]string{"priority": "high"})
	if err != domain.ErrAccessDenied {
		t.Errorf("UserB should not create automation in BoardA, got %v", err)
	}
}
