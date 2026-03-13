package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/your-org/ims-th-solution/backend/internal/iam/application"
	"github.com/your-org/ims-th-solution/backend/internal/iam/application/command"
	"github.com/your-org/ims-th-solution/backend/internal/iam/application/query"
	"github.com/your-org/ims-th-solution/backend/internal/iam/domain"
	"github.com/your-org/ims-th-solution/backend/internal/iam/transport/dto/request"
	"github.com/your-org/ims-th-solution/backend/internal/iam/transport/dto/response"
	apperrors "github.com/your-org/ims-th-solution/backend/platform/errors"
	"github.com/your-org/ims-th-solution/backend/platform/httputil"
	"github.com/your-org/ims-th-solution/backend/platform/middleware"
	"github.com/your-org/ims-th-solution/backend/platform/validation"
)

// AuthHandler handles HTTP requests for authentication endpoints.
// Handlers are thin: parse request → call use case → format response.
type AuthHandler struct {
	loginCmd *command.LoginCommand
	getMeQry *query.GetMeQuery
	tokenSvc *application.TokenService
	userRepo domain.UserRepository
}

// NewAuthHandler creates an AuthHandler with its dependencies.
func NewAuthHandler(
	loginCmd *command.LoginCommand,
	getMeQry *query.GetMeQuery,
	tokenSvc *application.TokenService,
	userRepo domain.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		loginCmd: loginCmd,
		getMeQry: getMeQry,
		tokenSvc: tokenSvc,
		userRepo: userRepo,
	}
}

// Login handles POST /auth/login — authenticates user and returns JWT token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}

	// Validate request fields
	if fieldErrors := validation.ValidateStruct(&req); fieldErrors != nil {
		httputil.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "validation failed",
			"details": fieldErrors,
		})
		return
	}

	// Execute login use case
	result, err := h.loginCmd.Execute(r.Context(), command.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		handleAuthError(w, err)
		return
	}

	// Map application DTO → transport response DTO
	httputil.OK(w, response.LoginResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
		User: response.UserResponse{
			ID:          result.User.ID,
			Username:    result.User.Username,
			DisplayName: result.User.DisplayName,
			Email:       result.User.Email,
			Groups:      result.User.Groups,
			IsOnLeave:   result.User.IsOnLeave,
		},
		Permissions: response.PermissionsResp{
			Functions: result.Permissions.Functions,
			Contracts: result.Permissions.Contracts,
		},
	})
}

// GetMe handles GET /auth/me — returns the current authenticated user's profile.
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	profile, err := h.getMeQry.Execute(r.Context(), claims.UserID)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	httputil.OK(w, response.UserResponse{
		ID:          profile.ID,
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
		Email:       profile.Email,
		Groups:      profile.Groups,
		IsOnLeave:   profile.IsOnLeave,
	})
}

// RefreshToken handles POST /auth/refresh — refreshes the JWT token.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	// Look up user to ensure they're still active
	user, err := h.userRepo.FindByUsername(r.Context(), claims.Username)
	if err != nil || user == nil {
		httputil.Unauthorized(w, "user not found")
		return
	}

	if allowed, reason := user.IsLoginAllowed(); !allowed {
		httputil.Forbidden(w, reason)
		return
	}

	token, expiresAt, err := h.tokenSvc.RefreshToken(user)
	if err != nil {
		httputil.InternalError(w, "failed to refresh token")
		return
	}

	httputil.OK(w, map[string]interface{}{
		"token":      token,
		"expires_at": expiresAt,
	})
}

// handleAuthError maps application errors to HTTP responses.
func handleAuthError(w http.ResponseWriter, err error) {
	var bizErr *apperrors.BusinessError
	if errors.As(err, &bizErr) {
		switch bizErr.Code {
		case apperrors.CodeUnauthorized:
			httputil.Unauthorized(w, bizErr.Message)
		case apperrors.CodeForbidden:
			httputil.Forbidden(w, bizErr.Message)
		default:
			httputil.BadRequest(w, bizErr.Message)
		}
		return
	}

	var notFoundErr *apperrors.NotFoundError
	if errors.As(err, &notFoundErr) {
		httputil.NotFound(w, notFoundErr.Error())
		return
	}

	httputil.InternalError(w, "an internal error occurred")
}
