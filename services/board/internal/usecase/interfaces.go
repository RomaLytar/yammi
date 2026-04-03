package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

// BoardRepository определяет интерфейс для работы с досками
type BoardRepository interface {
	// Create создает новую доску и сохраняет её в БД
	Create(ctx context.Context, board *domain.Board) error

	// GetByID возвращает доску по ID
	GetByID(ctx context.Context, boardID string) (*domain.Board, error)

	// ListByUserID возвращает список досок пользователя (с фильтрацией, поиском, сортировкой)
	ListByUserID(ctx context.Context, userID string, limit int, cursor string, ownerOnly bool, search string, sortBy string) ([]*domain.Board, string, error)

	// Update обновляет доску (с проверкой оптимистичной блокировки)
	Update(ctx context.Context, board *domain.Board) error

	// Delete удаляет доску по ID
	Delete(ctx context.Context, boardID string) error

	// BatchDelete удаляет несколько досок в одной транзакции
	BatchDelete(ctx context.Context, boardIDs []string) error

	// TouchUpdatedAt обновляет updated_at доски (при изменении карточек/колонок)
	TouchUpdatedAt(ctx context.Context, boardID string) error
}

// ColumnRepository определяет интерфейс для работы с колонками
type ColumnRepository interface {
	// Create создает новую колонку
	Create(ctx context.Context, column *domain.Column) error

	// BatchCreate создает несколько колонок в одном запросе
	BatchCreate(ctx context.Context, columns []*domain.Column) error

	// GetByID возвращает колонку по ID
	GetByID(ctx context.Context, columnID string) (*domain.Column, error)

	// ListByBoardID возвращает все колонки доски в порядке сортировки
	ListByBoardID(ctx context.Context, boardID string) ([]*domain.Column, error)

	// Update обновляет колонку
	Update(ctx context.Context, column *domain.Column) error

	// Delete удаляет колонку по ID (сначала удаляет карточки колонки)
	Delete(ctx context.Context, columnID, boardID string) error
}

// CardRepository определяет интерфейс для работы с карточками
type CardRepository interface {
	// Create создает новую карточку
	Create(ctx context.Context, card *domain.Card) error

	// GetByID возвращает карточку по ID (фильтруется по boardID для защиты от IDOR)
	GetByID(ctx context.Context, cardID, boardID string) (*domain.Card, error)

	// GetLastInColumn возвращает последнюю карточку в колонке (для генерации lexorank)
	GetLastInColumn(ctx context.Context, columnID string) (*domain.Card, error)

	// ListByColumnID возвращает все карточки колонки в порядке позиции
	ListByColumnID(ctx context.Context, columnID string) ([]*domain.Card, error)

	// Update обновляет карточку
	Update(ctx context.Context, card *domain.Card) error

	// Delete удаляет карточку по ID (boardID для partition pruning)
	Delete(ctx context.Context, cardID, boardID string) error

	// CountByBoard возвращает количество карточек по колонкам доски
	CountByBoard(ctx context.Context, boardID string) (map[string]int, error)

	// BatchDelete удаляет несколько карточек по ID в рамках одной доски
	BatchDelete(ctx context.Context, boardID string, cardIDs []string) error

	// UnassignByUser снимает assignee со всех карточек удалённого участника
	UnassignByUser(ctx context.Context, boardID, userID string) (int, error)

	// SearchByBoardID ищет карточки по доске с опциональными фильтрами
	SearchByBoardID(ctx context.Context, boardID string, search string, assigneeID string, priority string, taskType string) ([]*domain.Card, error)

	// ListByReleaseID возвращает карточки релиза
	ListByReleaseID(ctx context.Context, boardID, releaseID string) ([]*domain.Card, error)

	// ListBacklog возвращает карточки без релиза (бэклог)
	ListBacklog(ctx context.Context, boardID string) ([]*domain.Card, error)

	// MoveToBacklog снимает release_id со всех карточек релиза
	MoveToBacklog(ctx context.Context, boardID, releaseID string) (int, error)

	// MoveToBacklogExceptColumn снимает release_id с карточек релиза, кроме указанной колонки
	MoveToBacklogExceptColumn(ctx context.Context, boardID, releaseID, exceptColumnID string) (int, error)

	// SetReleaseID устанавливает или снимает release_id у карточки
	SetReleaseID(ctx context.Context, cardID, boardID string, releaseID *string) error
}

// MembershipRepository определяет интерфейс для работы с членством в досках
type MembershipRepository interface {
	// AddMember добавляет пользователя в доску с указанной ролью и возвращает созданного участника
	AddMember(ctx context.Context, boardID, userID string, role domain.Role) (*domain.Member, error)

	// RemoveMember удаляет пользователя из доски
	RemoveMember(ctx context.Context, boardID, userID string) error

	// IsMember проверяет, является ли пользователь членом доски и возвращает его роль
	IsMember(ctx context.Context, boardID, userID string) (bool, domain.Role, error)

	// ListMembers возвращает список членов доски с пагинацией
	ListMembers(ctx context.Context, boardID string, limit, offset int) ([]*domain.Member, error)
}

// ActivityRepository определяет интерфейс для работы с журналом активности
type ActivityRepository interface {
	// Create создает запись активности
	Create(ctx context.Context, activity *domain.Activity) error

	// ListByCardID возвращает записи активности карточки с пагинацией
	ListByCardID(ctx context.Context, cardID, boardID string, limit int, cursor string) ([]*domain.Activity, string, error)
}

// AttachmentRepository определяет интерфейс для работы с вложениями
type AttachmentRepository interface {
	// Create создает запись о вложении
	Create(ctx context.Context, attachment *domain.Attachment) error

	// GetByID возвращает вложение по ID (фильтруется по boardID для защиты от IDOR)
	GetByID(ctx context.Context, attachmentID, boardID string) (*domain.Attachment, error)

	// ListByCardID возвращает все вложения карточки
	ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Attachment, error)

	// Delete удаляет вложение по ID
	Delete(ctx context.Context, attachmentID, boardID string) error

	// CountByCardID возвращает количество вложений карточки
	CountByCardID(ctx context.Context, cardID, boardID string) (int, error)
}

// LabelRepository определяет интерфейс для работы с метками
type LabelRepository interface {
	// Create создает новую метку
	Create(ctx context.Context, label *domain.Label) error

	// BatchCreate создает несколько меток в одном запросе
	BatchCreate(ctx context.Context, labels []*domain.Label) error

	// GetByID возвращает метку по ID (фильтруется по boardID для защиты от IDOR)
	GetByID(ctx context.Context, labelID, boardID string) (*domain.Label, error)

	// ListByBoardID возвращает все метки доски
	ListByBoardID(ctx context.Context, boardID string) ([]*domain.Label, error)

	// Update обновляет метку
	Update(ctx context.Context, label *domain.Label) error

	// Delete удаляет метку по ID (фильтруется по boardID для защиты от IDOR)
	Delete(ctx context.Context, labelID, boardID string) error

	// AddToCard назначает метку на карточку
	AddToCard(ctx context.Context, cardID, boardID, labelID string) error

	// RemoveFromCard снимает метку с карточки
	RemoveFromCard(ctx context.Context, cardID, boardID, labelID string) error

	// ListByCardID возвращает все метки карточки
	ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Label, error)

	// CountByBoardID возвращает количество меток доски
	CountByBoardID(ctx context.Context, boardID string) (int, error)

	// CreateWithLimit создает метку с проверкой лимита в одном запросе
	CreateWithLimit(ctx context.Context, label *domain.Label, maxCount int) error
}

// ChecklistRepository определяет интерфейс для работы с чек-листами
type ChecklistRepository interface {
	CreateChecklist(ctx context.Context, checklist *domain.Checklist) error
	GetChecklistByID(ctx context.Context, checklistID, boardID string) (*domain.Checklist, error)
	ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Checklist, error)
	UpdateChecklist(ctx context.Context, checklist *domain.Checklist) error
	DeleteChecklist(ctx context.Context, checklistID, boardID string) error
	CreateItem(ctx context.Context, item *domain.ChecklistItem) error
	GetItemByID(ctx context.Context, itemID, boardID string) (*domain.ChecklistItem, error)
	ListItemsByChecklistID(ctx context.Context, checklistID, boardID string) ([]domain.ChecklistItem, error)
	UpdateItem(ctx context.Context, item *domain.ChecklistItem) error
	DeleteItem(ctx context.Context, itemID, boardID string) error
	ToggleItem(ctx context.Context, itemID, boardID string, isChecked bool) error
	ToggleItemAtomic(ctx context.Context, itemID, boardID string) (bool, error)
}

// CardLinkRepository определяет интерфейс для работы со связями карточек
type CardLinkRepository interface {
	// Create создает новую связь между карточками
	Create(ctx context.Context, link *domain.CardLink) error

	// CreateVerified создает связь, проверяя существование parent card в одном запросе
	CreateVerified(ctx context.Context, link *domain.CardLink) error

	// Delete удаляет связь по ID (boardID для partition pruning)
	Delete(ctx context.Context, linkID, boardID string) error

	// GetByID возвращает связь по ID (boardID для partition pruning)
	GetByID(ctx context.Context, linkID, boardID string) (*domain.CardLink, error)

	// ListChildren возвращает все дочерние связи карточки (boardID для partition pruning)
	ListChildren(ctx context.Context, parentID, boardID string) ([]*domain.CardLink, error)

	// ListParents возвращает все родительские связи карточки (без boardID — child может быть на любой доске)
	ListParents(ctx context.Context, childID string) ([]*domain.CardLink, error)

	// Exists проверяет существование связи между двумя карточками
	Exists(ctx context.Context, parentID, childID, boardID string) (bool, error)
}

// CustomFieldRepository определяет интерфейс для работы с кастомными полями
type CustomFieldRepository interface {
	// CreateDefinition создает новое определение кастомного поля
	CreateDefinition(ctx context.Context, def *domain.CustomFieldDefinition) error

	// GetDefinitionByID возвращает определение по ID (фильтруется по boardID для защиты от IDOR)
	GetDefinitionByID(ctx context.Context, defID, boardID string) (*domain.CustomFieldDefinition, error)

	// ListDefinitionsByBoardID возвращает все определения кастомных полей доски
	ListDefinitionsByBoardID(ctx context.Context, boardID string) ([]*domain.CustomFieldDefinition, error)

	// UpdateDefinition обновляет определение кастомного поля
	UpdateDefinition(ctx context.Context, def *domain.CustomFieldDefinition) error

	// DeleteDefinition удаляет определение кастомного поля (CASCADE удалит значения, фильтруется по boardID)
	DeleteDefinition(ctx context.Context, defID, boardID string) error

	// CountDefinitionsByBoardID возвращает количество определений кастомных полей доски
	CountDefinitionsByBoardID(ctx context.Context, boardID string) (int, error)

	// SetValue создает или обновляет значение кастомного поля для карточки
	SetValue(ctx context.Context, value *domain.CustomFieldValue) error

	// GetCardValues возвращает все значения кастомных полей карточки
	GetCardValues(ctx context.Context, cardID, boardID string) ([]*domain.CustomFieldValue, error)

	// DeleteValue удаляет значение кастомного поля карточки
	DeleteValue(ctx context.Context, cardID, boardID, fieldID string) error
}

// AutomationRuleRepository определяет интерфейс для работы с правилами автоматизации
type AutomationRuleRepository interface {
	// Create создает новое правило автоматизации
	Create(ctx context.Context, rule *domain.AutomationRule) error

	// GetByID возвращает правило по ID (фильтруется по boardID для защиты от IDOR)
	GetByID(ctx context.Context, ruleID, boardID string) (*domain.AutomationRule, error)

	// ListByBoardID возвращает все правила доски
	ListByBoardID(ctx context.Context, boardID string) ([]*domain.AutomationRule, error)

	// ListEnabledByBoardAndTrigger возвращает активные правила по доске и типу триггера
	ListEnabledByBoardAndTrigger(ctx context.Context, boardID string, triggerType domain.TriggerType) ([]*domain.AutomationRule, error)

	// Update обновляет правило
	Update(ctx context.Context, rule *domain.AutomationRule) error

	// Delete удаляет правило по ID (фильтруется по boardID для защиты от IDOR)
	Delete(ctx context.Context, ruleID, boardID string) error

	// CountByBoardID возвращает количество правил доски
	CountByBoardID(ctx context.Context, boardID string) (int, error)

	// CreateExecution создает запись о выполнении правила
	CreateExecution(ctx context.Context, exec *domain.AutomationExecution) error

	// ListExecutionsByRuleID возвращает историю выполнений правила
	ListExecutionsByRuleID(ctx context.Context, ruleID, boardID string, limit int) ([]*domain.AutomationExecution, error)
}

// BoardSettingsRepository определяет интерфейс для работы с настройками доски
type BoardSettingsRepository interface {
	// GetByBoardID возвращает настройки доски (дефолтные, если не найдены)
	GetByBoardID(ctx context.Context, boardID string) (*domain.BoardSettings, error)

	// Upsert создает или обновляет настройки доски
	Upsert(ctx context.Context, settings *domain.BoardSettings) error
}

// UserLabelRepository определяет интерфейс для работы с пользовательскими метками
type UserLabelRepository interface {
	// Create создает новую пользовательскую метку
	Create(ctx context.Context, label *domain.UserLabel) error

	// GetByID возвращает метку по ID
	GetByID(ctx context.Context, labelID string) (*domain.UserLabel, error)

	// ListByUserID возвращает все метки пользователя
	ListByUserID(ctx context.Context, userID string) ([]*domain.UserLabel, error)

	// Update обновляет метку
	Update(ctx context.Context, label *domain.UserLabel) error

	// Delete удаляет метку по ID
	Delete(ctx context.Context, labelID string) error

	// CountByUserID возвращает количество меток пользователя
	CountByUserID(ctx context.Context, userID string) (int, error)

	// CreateWithLimit создает метку с проверкой лимита в одном запросе
	CreateWithLimit(ctx context.Context, label *domain.UserLabel, maxCount int) error
}

// BoardTemplateRepository определяет интерфейс для работы с шаблонами досок
type BoardTemplateRepository interface {
	Create(ctx context.Context, t *domain.BoardTemplate) error
	GetByID(ctx context.Context, id string) (*domain.BoardTemplate, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.BoardTemplate, error)
	Update(ctx context.Context, t *domain.BoardTemplate) error
	Delete(ctx context.Context, id string) error
}

// ReleaseRepository определяет интерфейс для работы с релизами
type ReleaseRepository interface {
	// Create создает новый релиз
	Create(ctx context.Context, release *domain.Release) error

	// GetByID возвращает релиз по ID (фильтруется по boardID для защиты от IDOR)
	GetByID(ctx context.Context, releaseID, boardID string) (*domain.Release, error)

	// ListByBoardID возвращает все релизы доски
	ListByBoardID(ctx context.Context, boardID string) ([]*domain.Release, error)

	// GetActiveByBoardID возвращает активный релиз доски (или ErrReleaseNotFound)
	GetActiveByBoardID(ctx context.Context, boardID string) (*domain.Release, error)

	// Update обновляет релиз
	Update(ctx context.Context, release *domain.Release) error

	// Delete удаляет релиз по ID (фильтруется по boardID для защиты от IDOR)
	Delete(ctx context.Context, releaseID, boardID string) error

	// CountByBoardID возвращает количество релизов доски
	CountByBoardID(ctx context.Context, boardID string) (int, error)
}

// FileStorage определяет интерфейс для работы с файловым хранилищем
type FileStorage interface {
	// GenerateUploadURL генерирует pre-signed URL для загрузки файла
	GenerateUploadURL(ctx context.Context, key, contentType string, size int64) (string, error)

	// GenerateDownloadURL генерирует pre-signed URL для скачивания файла
	GenerateDownloadURL(ctx context.Context, key string) (string, error)

	// Delete удаляет файл из хранилища
	Delete(ctx context.Context, key string) error

	// Exists проверяет существование файла в хранилище
	Exists(ctx context.Context, key string) (bool, error)

	// Stat возвращает размер и content-type загруженного файла
	Stat(ctx context.Context, key string) (size int64, contentType string, err error)
}
