package domain

import "time"

// Role представляет роль участника доски
type Role string

const (
	RoleOwner  Role = "owner"  // Владелец — полные права
	RoleMember Role = "member" // Участник — CRUD карточек, чтение доски
)

// IsValid проверяет валидность роли
func (r Role) IsValid() bool {
	return r == RoleOwner || r == RoleMember
}

// String реализует fmt.Stringer
func (r Role) String() string {
	return string(r)
}

// Member — value object, представляющий участника доски.
// НЕ хранится в Board aggregate (для производительности).
// Живет в отдельной таблице board_members.
type Member struct {
	UserID   string
	Role     Role
	JoinedAt time.Time
}

// NewMember создает нового участника с валидацией
func NewMember(userID string, role Role) (*Member, error) {
	if userID == "" {
		return nil, ErrEmptyOwnerID
	}

	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	return &Member{
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}, nil
}

// IsOwner проверяет, является ли участник владельцем
func (m *Member) IsOwner() bool {
	return m.Role == RoleOwner
}

// CanModifyBoard проверяет, может ли участник изменять доску (owner only)
func (m *Member) CanModifyBoard() bool {
	return m.Role == RoleOwner
}

// CanModifyCards проверяет, может ли участник изменять карточки (owner + member)
func (m *Member) CanModifyCards() bool {
	return m.Role == RoleOwner || m.Role == RoleMember
}
