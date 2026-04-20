package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
	"github.com/neo-kanta/ims-th-solution/backend/platform/validation"
)

// MFAHandler handles MFA-related HTTP endpoints.
type MFAHandler struct {
	mfaCmd           *command.MFAEnrollCommand
	allowDevTOTPCode bool
}

// NewMFAHandler creates a new MFAHandler.
func NewMFAHandler(mfaCmd *command.MFAEnrollCommand, allowDevTOTPCode bool) *MFAHandler {
	return &MFAHandler{
		mfaCmd:           mfaCmd,
		allowDevTOTPCode: allowDevTOTPCode,
	}
}

// Enroll handles POST /auth/mfa/enroll.
// @Summary Start MFA Enrollment
// @Description Generate TOTP secret and recovery codes for MFA setup
// @Tags MFA
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.MFAEnrollResponse
// @Failure 401 {object} map[string]interface{}
// @Router /auth/mfa/enroll [post]
func (h *MFAHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	result, err := h.mfaCmd.Enroll(r.Context(), command.EnrollInput{
		UserID:    userID,
		IPAddress: command.ExtractIPAddress(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		httputil.InternalError(w, "failed to start MFA enrollment")
		return
	}

	httputil.OK(w, response.MFAEnrollResponse{
		ProvisioningURI: result.ProvisioningURI,
		RecoveryCodes:   result.RecoveryCodes,
	})
}

// Verify handles POST /auth/mfa/verify.
// @Summary Verify and Enable MFA
// @Description Verify TOTP code to complete MFA enrollment and activate
// @Tags MFA
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.MFAVerifyRequest true "TOTP Code"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/mfa/verify [post]
func (h *MFAHandler) Verify(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req request.MFAVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	if fieldErrors := validation.ValidateStruct(&req); fieldErrors != nil {
		httputil.JSON(w, http.StatusBadRequest, map[string]interface{}{"error": "validation failed", "details": fieldErrors})
		return
	}

	if err := h.mfaCmd.Verify(r.Context(), command.VerifyInput{
		UserID:    userID,
		TOTPCode:  req.TOTPCode,
		IPAddress: command.ExtractIPAddress(r),
		UserAgent: r.UserAgent(),
	}); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	httputil.NoContent(w)
}

// Disable handles POST /auth/mfa/disable.
// @Summary Disable MFA
// @Description Disable MFA for the current user (requires valid TOTP code)
// @Tags MFA
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.MFADisableRequest true "TOTP Code"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/mfa/disable [post]
func (h *MFAHandler) Disable(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req request.MFADisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	if fieldErrors := validation.ValidateStruct(&req); fieldErrors != nil {
		httputil.JSON(w, http.StatusBadRequest, map[string]interface{}{"error": "validation failed", "details": fieldErrors})
		return
	}

	if err := h.mfaCmd.Disable(r.Context(), command.DisableInput{
		UserID:    userID,
		TOTPCode:  req.TOTPCode,
		IPAddress: command.ExtractIPAddress(r),
		UserAgent: r.UserAgent(),
	}); err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	httputil.NoContent(w)
}

// Status handles GET /auth/mfa/status.
// @Summary Get MFA Status
// @Description Get the current MFA enrollment status for the authenticated user
// @Tags MFA
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.MFAStatusResponse
// @Failure 401 {object} map[string]interface{}
// @Router /auth/mfa/status [get]
func (h *MFAHandler) Status(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	status, err := h.mfaCmd.GetStatus(r.Context(), userID)
	if err != nil {
		httputil.InternalError(w, "failed to get MFA status")
		return
	}

	httputil.OK(w, response.MFAStatusResponse{
		Enrolled:          status.Enrolled,
		Enabled:           status.Enabled,
		RecoveryCodesLeft: status.RecoveryCodesLeft,
	})
}

// DevTOTPCode handles GET /auth/mfa/dev/totp-code.
// @Summary Get Current TOTP Code (Dev Only)
// @Description Development/test-only helper to retrieve the current TOTP code for the authenticated user
// @Tags MFA
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.MFADevTOTPCodeResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /auth/mfa/dev/totp-code [get]
func (h *MFAHandler) DevTOTPCode(w http.ResponseWriter, r *http.Request) {
	if !h.allowDevTOTPCode {
		httputil.NotFound(w, "resource not found")
		return
	}

	userID, err := getUserID(r)
	if err != nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	result, err := h.mfaCmd.GetDevelopmentTOTPCode(r.Context(), userID)
	if err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	httputil.OK(w, response.MFADevTOTPCodeResponse{
		TOTPCode: result.TOTPCode,
	})
}

func getUserID(r *http.Request) (uuid.UUID, error) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		return uuid.Nil, http.ErrNoCookie
	}
	return uuid.Parse(claims.Subject)
}
