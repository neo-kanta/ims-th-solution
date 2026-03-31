package request

// MFAVerifyRequest is the HTTP request for verifying and enabling MFA.
type MFAVerifyRequest struct {
	TOTPCode string `json:"totp_code" validate:"required,len=6"`
}

// MFADisableRequest is the HTTP request for disabling MFA.
type MFADisableRequest struct {
	TOTPCode string `json:"totp_code" validate:"required,len=6"`
}
