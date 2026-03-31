package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
	"github.com/neo-kanta/ims-th-solution/backend/platform/validation"
)

// AdminHandler handles IAM administrative endpoints.
type AdminHandler struct {
	adminCmd *command.AdminUserCommand
}

// NewAdminHandler creates an AdminHandler.
func NewAdminHandler(adminCmd *command.AdminUserCommand) *AdminHandler {
	return &AdminHandler{
		adminCmd: adminCmd,
	}
}

// CreateUser handles POST /admin/users
// @Summary Create User
// @Description Administrative creation of a new IAM user
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.CreateUserRequest true "New User Payload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users [post]
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	adminID, err := getAdminID(r)
	if err != nil {
		httputil.Unauthorized(w, err.Error())
		return
	}

	var req request.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	if fieldErrors := validation.ValidateStruct(&req); fieldErrors != nil {
		httputil.JSON(w, http.StatusBadRequest, map[string]interface{}{"error": "validation failed", "details": fieldErrors})
		return
	}

	newID, err := h.adminCmd.CreateUser(r.Context(), command.CreateUserInput{
		AdminID:     adminID,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Password:    req.Password,
		IPAddress:   command.ExtractIPAddress(r),
		UserAgent:   r.UserAgent(),
	})
	if err != nil {
		handleAuthError(w, err)
		return
	}

	httputil.JSON(w, http.StatusCreated, map[string]string{"id": newID.String()})
}

// DisableUser handles POST /admin/users/{id}/disable
// @Summary Disable User
// @Description Disable a specific user account preventing future logins
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "User UUID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users/{id}/disable [post]
func (h *AdminHandler) DisableUser(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, "disable")
}

// EnableUser handles POST /admin/users/{id}/enable
// @Summary Enable User
// @Description Re-enable a disabled user account
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "User UUID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users/{id}/enable [post]
func (h *AdminHandler) EnableUser(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, "enable")
}

// LockUser handles POST /admin/users/{id}/lock
// @Summary Lock User
// @Description Immediately lock a user out of the system
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "User UUID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users/{id}/lock [post]
func (h *AdminHandler) LockUser(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, "lock")
}

// UnlockUser handles POST /admin/users/{id}/unlock
// @Summary Unlock User
// @Description Unlock an account that was either manually or automatically locked
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "User UUID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users/{id}/unlock [post]
func (h *AdminHandler) UnlockUser(w http.ResponseWriter, r *http.Request) {
	h.setStatus(w, r, "unlock")
}

func (h *AdminHandler) setStatus(w http.ResponseWriter, r *http.Request, action string) {
	adminID, err := getAdminID(r)
	if err != nil {
		httputil.Unauthorized(w, err.Error())
		return
	}

	targetIDStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httputil.BadRequest(w, "invalid user id")
		return
	}

	if err := h.adminCmd.SetUserStatus(r.Context(), command.SetUserStatusInput{
		AdminID:   adminID,
		TargetID:  targetID,
		Action:    action,
		IPAddress: command.ExtractIPAddress(r),
		UserAgent: r.UserAgent(),
	}); err != nil {
		handleAuthError(w, err)
		return
	}

	httputil.NoContent(w)
}

// ResetPassword handles POST /admin/users/{id}/reset-password
// @Summary Admin Password Reset
// @Description Forcefully reset a user's password and flag it for mandatory change on next login
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User UUID" format(uuid)
// @Param request body request.AdminResetPasswordRequest true "New Password Payload"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users/{id}/reset-password [post]
func (h *AdminHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	adminID, err := getAdminID(r)
	if err != nil {
		httputil.Unauthorized(w, err.Error())
		return
	}

	targetIDStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httputil.BadRequest(w, "invalid user id")
		return
	}

	var req request.AdminResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	if fieldErrors := validation.ValidateStruct(&req); fieldErrors != nil {
		httputil.JSON(w, http.StatusBadRequest, map[string]interface{}{"error": "validation failed", "details": fieldErrors})
		return
	}

	if err := h.adminCmd.ResetPassword(r.Context(), command.ResetPasswordInput{
		AdminID:     adminID,
		TargetID:    targetID,
		NewPassword: req.NewPassword,
		IPAddress:   command.ExtractIPAddress(r),
		UserAgent:   r.UserAgent(),
	}); err != nil {
		handleAuthError(w, err)
		return
	}

	httputil.NoContent(w)
}

func getAdminID(r *http.Request) (uuid.UUID, error) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		return uuid.Nil, http.ErrNoCookie
	}
	return uuid.Parse(claims.Subject)
}
