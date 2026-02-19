package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin UserRole = "ADMIN"
	RoleStaff UserRole = "STAFF"
	RoleBuyer UserRole = "BUYER"
)

// User represents an authenticated entity in the system.
type User struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	Role       UserRole  `json:"role"`
}

// AuthRequestDTO is what the React frontend sends us.
type AuthRequestDTO struct {
	InitData string `json:"init_data" binding:"required"` // The raw string from window.Telegram.WebApp.initData
}

// AuthResponseDTO is what we give back (The JWT Token).
type AuthResponseDTO struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UserRepository interface {
	UpsertUser(ctx context.Context, user *User) error // Create if not exists, update if exists
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type AuthService interface {
	LoginTelegram(ctx context.Context, initData string) (*AuthResponseDTO, error)
}
