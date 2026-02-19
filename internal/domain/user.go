package domain

import "github.com/google/uuid"

// UserRole defines the access level of a user.
type UserRole string

const (
	RoleAdmin UserRole = "ADMIN"
	RoleStaff UserRole = "STAFF"
	RoleBuyer UserRole = "BUYER"
)

// User represents an authenticated entity in the system.
type User struct {
	ID    uuid.UUID
	Name  string
	Email string
	Role  UserRole
}
