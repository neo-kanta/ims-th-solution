package dto

import "time"

// LoginResult is the application-level DTO returned after successful authentication.
type LoginResult struct {
	AccessToken           string          `json:"access_token,omitempty"`
	AccessTokenExpiresAt  time.Time       `json:"access_token_expires_at,omitempty"`
	RefreshToken          string          `json:"refresh_token,omitempty"`
	RefreshTokenExpiresAt time.Time       `json:"refresh_token_expires_at,omitempty"`
	ForcePasswordChange   bool            `json:"force_password_change"`
	User                  UserProfile     `json:"user,omitempty"`
	Permissions           UserPermissions `json:"permissions,omitempty"`
	// MFA challenge fields
	MFARequired bool   `json:"mfa_required,omitempty"`
	MFAToken    string `json:"mfa_token,omitempty"` // opaque token for MFA step
}

// UserProfile represents the current user's identity information.
type UserProfile struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
}

// UserPermissions holds the user's effective permissions.
type UserPermissions struct {
	Functions []string `json:"functions"`
	Contracts []string `json:"contracts"`
}

// RefreshResult is the DTO returned after successful token refresh.
type RefreshResult struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

// SessionInfo represents a single active session for listing purposes.
type SessionInfo struct {
	ID             string    `json:"id"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	LastActivityAt time.Time `json:"last_activity_at"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsCurrent      bool      `json:"is_current,omitempty"`
}
