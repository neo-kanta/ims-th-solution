package request

// LoginRequest is the HTTP request body for POST /auth/login.
type LoginRequest struct {
	Username     string `json:"username" validate:"required,min=3,max=100"`
	Password     string `json:"password" validate:"required,min=8"`
	TOTPCode     string `json:"totp_code,omitempty"`     // MFA TOTP code (if MFA enrolled)
	RecoveryCode string `json:"recovery_code,omitempty"` // MFA recovery code (alternative)
	MFAToken     string `json:"mfa_token,omitempty"`     // MFA challenge token returned after password step
}

// RefreshRequest is the HTTP request body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest is the HTTP request body for POST /auth/logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
