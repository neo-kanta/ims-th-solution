package command

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	auditentity "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

// MFAEnrollCommand handles MFA enrollment, verification, and disabling.
type MFAEnrollCommand struct {
	mfaRepo  domain.MFARepository
	userRepo domain.UserRepository
	auditSvc auditdomain.Recorder
	totpSvc  *application.TOTPService
}

// NewMFAEnrollCommand creates a new MFAEnrollCommand.
func NewMFAEnrollCommand(
	mfaRepo domain.MFARepository,
	userRepo domain.UserRepository,
	auditSvc auditdomain.Recorder,
	totpSvc *application.TOTPService,
) *MFAEnrollCommand {
	if auditSvc == nil {
		auditSvc = auditdomain.NopRecorder{}
	}
	return &MFAEnrollCommand{
		mfaRepo:  mfaRepo,
		userRepo: userRepo,
		auditSvc: auditSvc,
		totpSvc:  totpSvc,
	}
}

// EnrollInput is the input for starting MFA enrollment.
type EnrollInput struct {
	UserID    uuid.UUID
	IPAddress string
	UserAgent string
}

// EnrollResult is the output of starting MFA enrollment.
type EnrollResult struct {
	Secret          string   `json:"secret"` // base32 TOTP secret (only shown once)
	ProvisioningURI string   `json:"provisioning_uri"`
	RecoveryCodes   []string `json:"recovery_codes"` // plaintext recovery codes (only shown once)
}

// Enroll starts MFA enrollment: generates TOTP secret and recovery codes.
// The user must then verify with a valid TOTP code to activate.
func (c *MFAEnrollCommand) Enroll(ctx context.Context, input EnrollInput) (*EnrollResult, error) {
	user, err := c.userRepo.FindByID(ctx, input.UserID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate TOTP secret
	encryptedSecret, provisioningURI, err := c.totpSvc.GenerateSecret(user.Username)
	if err != nil {
		return nil, fmt.Errorf("generating TOTP secret: %w", err)
	}

	// Create enrollment (upserts if re-enrolling)
	enrollment := &entity.MFAEnrollment{
		ID:              uuid.New(),
		UserID:          input.UserID,
		MFAType:         "totp",
		SecretEncrypted: encryptedSecret,
	}
	if err := c.mfaRepo.CreateEnrollment(ctx, enrollment); err != nil {
		return nil, fmt.Errorf("creating MFA enrollment: %w", err)
	}

	// Generate recovery codes
	plaintextCodes, hashedCodes, err := c.totpSvc.GenerateRecoveryCodes(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("generating recovery codes: %w", err)
	}
	if err := c.mfaRepo.StoreRecoveryCodes(ctx, hashedCodes); err != nil {
		return nil, fmt.Errorf("storing recovery codes: %w", err)
	}

	// Audit
	c.recordAudit(ctx, &input.UserID, auditentity.AuditMFAEnrolled, "user", input.UserID.String(), input.IPAddress, input.UserAgent, nil)

	return &EnrollResult{
		ProvisioningURI: provisioningURI,
		RecoveryCodes:   plaintextCodes,
	}, nil
}

// VerifyInput is the input for verifying and activating MFA.
type VerifyInput struct {
	UserID    uuid.UUID
	TOTPCode  string
	IPAddress string
	UserAgent string
}

// Verify validates a TOTP code and activates MFA for the user.
func (c *MFAEnrollCommand) Verify(ctx context.Context, input VerifyInput) error {
	enrollment, err := c.mfaRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("finding MFA enrollment: %w", err)
	}
	if enrollment == nil {
		return fmt.Errorf("MFA enrollment not found; please enroll first")
	}

	valid, err := c.totpSvc.ValidateCode(enrollment.SecretEncrypted, input.TOTPCode)
	if err != nil {
		return fmt.Errorf("validating TOTP code: %w", err)
	}
	if !valid {
		c.recordAudit(ctx, &input.UserID, auditentity.AuditMFAChallengeFail, "user", input.UserID.String(), input.IPAddress, input.UserAgent,
			map[string]interface{}{"reason": "invalid code during enrollment verification"})
		return fmt.Errorf("invalid TOTP code")
	}

	if err := c.mfaRepo.EnableEnrollment(ctx, input.UserID); err != nil {
		return fmt.Errorf("enabling MFA: %w", err)
	}

	c.recordAudit(ctx, &input.UserID, auditentity.AuditMFAEnabled, "user", input.UserID.String(), input.IPAddress, input.UserAgent, nil)
	return nil
}

// DisableInput is the input for disabling MFA.
type DisableInput struct {
	UserID    uuid.UUID
	TOTPCode  string // require current code to disable
	IPAddress string
	UserAgent string
}

// Disable deactivates MFA for a user (requires valid TOTP or admin action).
func (c *MFAEnrollCommand) Disable(ctx context.Context, input DisableInput) error {
	enrollment, err := c.mfaRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("finding MFA enrollment: %w", err)
	}
	if enrollment == nil || !enrollment.IsActive() {
		return fmt.Errorf("MFA is not active")
	}

	// Verify TOTP code before disabling
	valid, err := c.totpSvc.ValidateCode(enrollment.SecretEncrypted, input.TOTPCode)
	if err != nil {
		return fmt.Errorf("validating TOTP code: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid TOTP code")
	}

	if err := c.mfaRepo.DisableEnrollment(ctx, input.UserID); err != nil {
		return fmt.Errorf("disabling MFA: %w", err)
	}
	if err := c.mfaRepo.DeleteRecoveryCodes(ctx, input.UserID); err != nil {
		slog.Error("failed to delete recovery codes", "error", err)
	}

	c.recordAudit(ctx, &input.UserID, auditentity.AuditMFADisabled, "user", input.UserID.String(), input.IPAddress, input.UserAgent, nil)
	return nil
}

// AdminDisable allows an admin to disable MFA without the user's TOTP code.
func (c *MFAEnrollCommand) AdminDisable(ctx context.Context, adminID uuid.UUID, targetUserID uuid.UUID, ipAddress, userAgent string) error {
	if err := c.mfaRepo.DisableEnrollment(ctx, targetUserID); err != nil {
		return fmt.Errorf("disabling MFA: %w", err)
	}
	if err := c.mfaRepo.DeleteRecoveryCodes(ctx, targetUserID); err != nil {
		slog.Error("failed to delete recovery codes", "error", err)
	}

	c.recordAudit(ctx, &adminID, auditentity.AuditMFADisabled, "user", targetUserID.String(), ipAddress, userAgent,
		map[string]interface{}{"admin_action": true})
	return nil
}

// MFAStatusResult returns the current MFA status for a user.
type MFAStatusResult struct {
	Enrolled          bool `json:"enrolled"`
	Enabled           bool `json:"enabled"`
	RecoveryCodesLeft int  `json:"recovery_codes_left"`
}

// DevTOTPCodeResult returns the current TOTP code for development/test workflows.
type DevTOTPCodeResult struct {
	TOTPCode string `json:"totp_code"`
}

// GetStatus returns the current MFA status for a user.
func (c *MFAEnrollCommand) GetStatus(ctx context.Context, userID uuid.UUID) (*MFAStatusResult, error) {
	enrollment, err := c.mfaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding MFA enrollment: %w", err)
	}

	result := &MFAStatusResult{}
	if enrollment != nil {
		result.Enrolled = true
		result.Enabled = enrollment.IsActive()
	}

	codes, err := c.mfaRepo.FindUnusedRecoveryCodes(ctx, userID)
	if err == nil {
		result.RecoveryCodesLeft = len(codes)
	}

	return result, nil
}

// GetDevelopmentTOTPCode returns the current TOTP code for the caller's enrolled secret.
// This must only be exposed behind development/test-only transport wiring.
func (c *MFAEnrollCommand) GetDevelopmentTOTPCode(ctx context.Context, userID uuid.UUID) (*DevTOTPCodeResult, error) {
	enrollment, err := c.mfaRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding MFA enrollment: %w", err)
	}
	if enrollment == nil {
		return nil, fmt.Errorf("MFA enrollment not found; please enroll first")
	}

	code, err := c.totpSvc.GenerateCurrentCode(enrollment.SecretEncrypted)
	if err != nil {
		return nil, fmt.Errorf("generating development TOTP code: %w", err)
	}

	return &DevTOTPCodeResult{TOTPCode: code}, nil
}

func (c *MFAEnrollCommand) recordAudit(ctx context.Context, actorID *uuid.UUID, eventType, targetType, targetID, ipAddress, userAgent string, metadata map[string]interface{}) {
	c.auditSvc.Record(ctx, actorID, eventType, targetType, targetID, ipAddress, userAgent, metadata)
}
