package entity

import (
	"time"

	"github.com/google/uuid"
)

// User is the aggregate root for identity and access management.
// It represents a system user account with authentication and status information.
type User struct {
	ID           uuid.UUID
	Username     string
	DisplayName  string
	Email        string
	PasswordHash string
	IsActive     bool
	IsOnLeave    bool
	Groups       []string // Group names the user belongs to
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    *uuid.UUID
	UpdatedBy    *uuid.UUID
}

// IsLoginAllowed checks if the user is permitted to log in.
// Users who are inactive or on approved leave cannot log in.
func (u *User) IsLoginAllowed() (bool, string) {
	if !u.IsActive {
		return false, "account is deactivated"
	}
	if u.IsOnLeave {
		return false, "account is currently on leave — normal login is restricted"
	}
	return true, ""
}
