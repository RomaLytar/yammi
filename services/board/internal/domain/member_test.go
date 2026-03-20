package domain

import (
	"testing"
	"time"
)

func TestNewMember(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		role    Role
		wantErr error
	}{
		{
			name:    "valid owner",
			userID:  "user-123",
			role:    RoleOwner,
			wantErr: nil,
		},
		{
			name:    "valid member",
			userID:  "user-456",
			role:    RoleMember,
			wantErr: nil,
		},
		{
			name:    "empty user ID with owner role",
			userID:  "",
			role:    RoleOwner,
			wantErr: ErrEmptyOwnerID,
		},
		{
			name:    "empty user ID with member role",
			userID:  "",
			role:    RoleMember,
			wantErr: ErrEmptyOwnerID,
		},
		{
			name:    "invalid role",
			userID:  "user-123",
			role:    Role("admin"),
			wantErr: ErrInvalidRole,
		},
		{
			name:    "empty role string",
			userID:  "user-123",
			role:    Role(""),
			wantErr: ErrInvalidRole,
		},
		{
			name:    "both empty user ID and invalid role",
			userID:  "",
			role:    Role("invalid"),
			wantErr: ErrEmptyOwnerID, // userID проверяется первым
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member, err := NewMember(tt.userID, tt.role)

			if err != tt.wantErr {
				t.Errorf("NewMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				if member != nil {
					t.Errorf("NewMember() returned member when error expected")
				}
				return
			}

			// Проверяем корректность созданного участника
			if member == nil {
				t.Fatal("NewMember() returned nil member")
			}

			if member.UserID != tt.userID {
				t.Errorf("NewMember() UserID = %v, want %v", member.UserID, tt.userID)
			}

			if member.Role != tt.role {
				t.Errorf("NewMember() Role = %v, want %v", member.Role, tt.role)
			}

			if member.JoinedAt.IsZero() {
				t.Error("NewMember() JoinedAt is zero")
			}

			// JoinedAt должна быть близка к текущему времени
			if time.Since(member.JoinedAt) > time.Second {
				t.Error("NewMember() JoinedAt is too far in the past")
			}
		})
	}
}

func TestMember_IsOwner(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{
			name: "owner role",
			role: RoleOwner,
			want: true,
		},
		{
			name: "member role",
			role: RoleMember,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member, err := NewMember("user-123", tt.role)
			if err != nil {
				t.Fatalf("Failed to create test member: %v", err)
			}

			got := member.IsOwner()
			if got != tt.want {
				t.Errorf("Member.IsOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMember_CanModifyBoard(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{
			name: "owner can modify board",
			role: RoleOwner,
			want: true,
		},
		{
			name: "member cannot modify board",
			role: RoleMember,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member, err := NewMember("user-123", tt.role)
			if err != nil {
				t.Fatalf("Failed to create test member: %v", err)
			}

			got := member.CanModifyBoard()
			if got != tt.want {
				t.Errorf("Member.CanModifyBoard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMember_CanModifyCards(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{
			name: "owner can modify cards",
			role: RoleOwner,
			want: true,
		},
		{
			name: "member can modify cards",
			role: RoleMember,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member, err := NewMember("user-123", tt.role)
			if err != nil {
				t.Fatalf("Failed to create test member: %v", err)
			}

			got := member.CanModifyCards()
			if got != tt.want {
				t.Errorf("Member.CanModifyCards() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{
			name: "owner role is valid",
			role: RoleOwner,
			want: true,
		},
		{
			name: "member role is valid",
			role: RoleMember,
			want: true,
		},
		{
			name: "empty role is invalid",
			role: Role(""),
			want: false,
		},
		{
			name: "admin role is invalid",
			role: Role("admin"),
			want: false,
		},
		{
			name: "guest role is invalid",
			role: Role("guest"),
			want: false,
		},
		{
			name: "uppercase OWNER is invalid",
			role: Role("OWNER"),
			want: false,
		},
		{
			name: "uppercase MEMBER is invalid",
			role: Role("MEMBER"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.role.IsValid()
			if got != tt.want {
				t.Errorf("Role.IsValid() = %v, want %v for role %q", got, tt.want, tt.role)
			}
		})
	}
}

func TestRole_String(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want string
	}{
		{
			name: "owner role string",
			role: RoleOwner,
			want: "owner",
		},
		{
			name: "member role string",
			role: RoleMember,
			want: "member",
		},
		{
			name: "custom role string",
			role: Role("custom"),
			want: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.role.String()
			if got != tt.want {
				t.Errorf("Role.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMember_PermissionsMatrix(t *testing.T) {
	// Тест матрицы прав для различных ролей
	tests := []struct {
		role             Role
		canModifyBoard   bool
		canModifyCards   bool
		isOwner          bool
	}{
		{
			role:             RoleOwner,
			canModifyBoard:   true,
			canModifyCards:   true,
			isOwner:          true,
		},
		{
			role:             RoleMember,
			canModifyBoard:   false,
			canModifyCards:   true,
			isOwner:          false,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			member, err := NewMember("user-123", tt.role)
			if err != nil {
				t.Fatalf("Failed to create member with role %v: %v", tt.role, err)
			}

			if got := member.CanModifyBoard(); got != tt.canModifyBoard {
				t.Errorf("Member.CanModifyBoard() = %v, want %v for role %v", got, tt.canModifyBoard, tt.role)
			}

			if got := member.CanModifyCards(); got != tt.canModifyCards {
				t.Errorf("Member.CanModifyCards() = %v, want %v for role %v", got, tt.canModifyCards, tt.role)
			}

			if got := member.IsOwner(); got != tt.isOwner {
				t.Errorf("Member.IsOwner() = %v, want %v for role %v", got, tt.isOwner, tt.role)
			}
		})
	}
}

func TestMember_RoleConstants(t *testing.T) {
	// Проверяем, что константы ролей имеют правильные значения
	if RoleOwner != "owner" {
		t.Errorf("RoleOwner = %v, want 'owner'", RoleOwner)
	}

	if RoleMember != "member" {
		t.Errorf("RoleMember = %v, want 'member'", RoleMember)
	}

	// Проверяем, что роли валидны
	if !RoleOwner.IsValid() {
		t.Error("RoleOwner is not valid")
	}

	if !RoleMember.IsValid() {
		t.Error("RoleMember is not valid")
	}
}
