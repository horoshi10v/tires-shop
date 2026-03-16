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
	ID          uuid.UUID `json:"id"`
	TelegramID  int64     `json:"telegram_id"`
	Username    string    `json:"username"`
	FirstName   string    `json:"first_name"`
	PhoneNumber string    `json:"phone_number"`
	Role        UserRole  `json:"role"`
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

// UserFilter defines criteria for searching users.
type UserFilter struct {
	Page     int
	PageSize int
	Search   string // Searches by Username, FirstName, LastName, PhoneNumber
	Role     UserRole
}

// CreateWorkerDTO is for creating/promoting a user.
type CreateWorkerDTO struct {
	TelegramID  int64    `json:"telegram_id"` // Optional if searching by other fields
	Username    string   `json:"username"`
	FirstName   string   `json:"first_name"`
	PhoneNumber string   `json:"phone_number"`
	Role        UserRole `json:"role" binding:"required,oneof=STAFF ADMIN"`
}

// UpdateUserRoleDTO is specifically for changing a user's role.
type UpdateUserRoleDTO struct {
	Role UserRole `json:"role" binding:"required,oneof=BUYER STAFF ADMIN"`
}

type UserRepository interface {
	UpsertUser(ctx context.Context, user *User) error // Create if not exists, update if exists
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, filter UserFilter) ([]User, int64, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role UserRole) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AuthService interface {
	LoginTelegram(ctx context.Context, initData string) (*AuthResponseDTO, error)
}

type UserService interface {
	ListUsers(ctx context.Context, filter UserFilter) ([]User, int64, error)
	AddWorker(ctx context.Context, dto CreateWorkerDTO) (uuid.UUID, error)
	UpdateUserRole(ctx context.Context, id uuid.UUID, role UserRole) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
