package postgres

import "database/sql"

// Repositories содержит все PostgreSQL репозитории Board Service
type Repositories struct {
	Board      *BoardRepository
	Column     *ColumnRepository
	Card       *CardRepository
	Membership *MembershipRepository
}

// NewRepositories создает все репозитории с единой DB connection
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Board:      NewBoardRepository(db),
		Column:     NewColumnRepository(db),
		Card:       NewCardRepository(db),
		Membership: NewMembershipRepository(db),
	}
}
