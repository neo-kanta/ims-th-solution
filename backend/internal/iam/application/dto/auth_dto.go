package dto

import "time"

// LoginResult is the application-level DTO returned after successful authentication.
type LoginResult struct {
	Token       string      `json:"token"`
	ExpiresAt   time.Time   `json:"expires_at"`
	User        UserProfile `json:"user"`
	Permissions UserPermissions `json:"permissions"`
}

// UserProfile represents the current user's identity information.
type UserProfile struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
	IsOnLeave   bool     `json:"is_on_leave"`
}

// UserPermissions holds the user's effective permissions.
type UserPermissions struct {
	Functions []string `json:"functions"`
	Contracts []string `json:"contracts"`
}
