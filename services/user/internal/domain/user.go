package domain

import "time"

type User struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
	Bio       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUserFromEvent(id, email, name string) *User {
	now := time.Now()
	return &User{
		ID:        id,
		Email:     email,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (u *User) Update(name, avatarURL, bio string) error {
	if name == "" {
		return ErrEmptyName
	}
	u.Name = name
	u.AvatarURL = avatarURL
	u.Bio = bio
	u.UpdatedAt = time.Now()
	return nil
}
