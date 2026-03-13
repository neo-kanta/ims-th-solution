package response

import "time"

// LoginResponse is the HTTP response body for a successful login.
type LoginResponse struct {
	Token       string          `json:"token"`
	ExpiresAt   time.Time       `json:"expires_at"`
	User        UserResponse    `json:"user"`
	Permissions PermissionsResp `json:"permissions"`
}

// UserResponse is the HTTP response for user profile data.
type UserResponse struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
	IsOnLeave   bool     `json:"is_on_leave"`
}

// PermissionsResp is the HTTP response for user permissions.
type PermissionsResp struct {
	Functions []string `json:"functions"`
	Contracts []string `json:"contracts"`
}
