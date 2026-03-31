package response

import "time"

// LoginResponse is the HTTP response body for a successful login.
type LoginResponse struct {
	AccessToken           string          `json:"access_token,omitempty"`
	AccessTokenExpiresAt  time.Time       `json:"access_token_expires_at,omitempty"`
	RefreshToken          string          `json:"refresh_token,omitempty"`
	RefreshTokenExpiresAt time.Time       `json:"refresh_token_expires_at,omitempty"`
	ForcePasswordChange   bool            `json:"force_password_change"`
	User                  UserResponse    `json:"user,omitempty"`
	Permissions           PermissionsResp `json:"permissions,omitempty"`
	// MFA challenge response
	MFARequired bool   `json:"mfa_required,omitempty"`
	MFAToken    string `json:"mfa_token,omitempty"`
}

// UserResponse is the HTTP response for user profile data.
type UserResponse struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Groups      []string `json:"groups"`
}

// PermissionsResp is the HTTP response for user permissions.
type PermissionsResp struct {
	Functions []string `json:"functions"`
	Contracts []string `json:"contracts"`
}

// RefreshResponse is the HTTP response for a successful token refresh.
type RefreshResponse struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

// MeResponse is the HTTP response for GET /auth/me.
type MeResponse struct {
	User        UserResponse    `json:"user"`
	Permissions PermissionsResp `json:"permissions"`
}

// MFAEnrollResponse is the HTTP response for MFA enrollment.
type MFAEnrollResponse struct {
	ProvisioningURI string   `json:"provisioning_uri"`
	RecoveryCodes   []string `json:"recovery_codes"`
}

// MFAStatusResponse is the HTTP response for MFA status check.
type MFAStatusResponse struct {
	Enrolled          bool `json:"enrolled"`
	Enabled           bool `json:"enabled"`
	RecoveryCodesLeft int  `json:"recovery_codes_left"`
}

// SessionResponse is the HTTP response for a single session.
type SessionResponse struct {
	ID             string    `json:"id"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	LastActivityAt time.Time `json:"last_activity_at"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// AuditEventResponse is the HTTP response for a single audit event.
type AuditEventResponse struct {
	ID         string                 `json:"id"`
	ActorID    *string                `json:"actor_id"`
	EventType  string                 `json:"event_type"`
	TargetType string                 `json:"target_type"`
	TargetID   string                 `json:"target_id"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// AuditListResponse is the paginated audit event list response.
type AuditListResponse struct {
	Events []AuditEventResponse `json:"events"`
	Total  int                  `json:"total"`
	Offset int                  `json:"offset"`
	Limit  int                  `json:"limit"`
}
