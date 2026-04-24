package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/dto/response"
	apperrors "github.com/neo-kanta/ims-th-solution/backend/platform/errors"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
	"github.com/neo-kanta/ims-th-solution/backend/platform/validation"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	loginCmd            *command.LoginCommand
	refreshCmd          *command.RefreshTokenCommand
	logoutCmd           *command.LogoutCommand
	logoutAllCmd        *command.LogoutAllCommand
	changePasswordCmd   *command.ChangePasswordCommand
	getMeQry            *query.GetMeQuery
	loginLimiter        middleware.RateLimiter
	loginPolicy         middleware.RateLimitPolicy
	loginUserLimiter    middleware.RateLimiter // per IP+username rate limiter (nil = disabled)
	loginUserPolicy     middleware.RateLimitPolicy
	refreshTokenLimiter middleware.RateLimiter // per-token rate limiter (nil = disabled)
	refreshTokenPolicy  middleware.RateLimitPolicy
}

// NewAuthHandler creates an AuthHandler with its dependencies.
func NewAuthHandler(
	loginCmd *command.LoginCommand,
	refreshCmd *command.RefreshTokenCommand,
	logoutCmd *command.LogoutCommand,
	logoutAllCmd *command.LogoutAllCommand,
	changePasswordCmd *command.ChangePasswordCommand,
	getMeQry *query.GetMeQuery,
	loginLimiter middleware.RateLimiter,
	loginPolicy middleware.RateLimitPolicy,
	loginUserLimiter middleware.RateLimiter,
	loginUserPolicy middleware.RateLimitPolicy,
	refreshTokenLimiter middleware.RateLimiter,
	refreshTokenPolicy middleware.RateLimitPolicy,
) *AuthHandler {
	return &AuthHandler{
		loginCmd:            loginCmd,
		refreshCmd:          refreshCmd,
		logoutCmd:           logoutCmd,
		logoutAllCmd:        logoutAllCmd,
		changePasswordCmd:   changePasswordCmd,
		getMeQry:            getMeQry,
		loginLimiter:        loginLimiter,
		loginPolicy:         loginPolicy,
		loginUserLimiter:    loginUserLimiter,
		loginUserPolicy:     loginUserPolicy,
		refreshTokenLimiter: refreshTokenLimiter,
		refreshTokenPolicy:  refreshTokenPolicy,
	}
}

// Login handles POST /auth/login.
// @Summary User Login
// @Description Authenticate user and return access and refresh tokens. Returns mfa_required if MFA is active.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "Login Credentials"
// @Success 200 {object} response.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}

	if fieldErrors := validation.ValidateStruct(&req); fieldErrors != nil {
		httputil.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "validation failed",
			"details": fieldErrors,
		})
		return
	}

	// Per IP+username rate limiting — prevents distributed brute force against a single account.
	// This is checked here (not middleware) because the username is in the request body.
	clientIP := middleware.GetClientIP(r)
	if h.loginUserLimiter != nil {
		key := fmt.Sprintf("rl:login_user:%s:%s", clientIP, strings.ToLower(req.Username))
		allowed, _, retryAfter, err := h.loginUserLimiter.Allow(r.Context(), key)
		if err != nil {
			httputil.InternalError(w, "rate limit check failed")
			return
		}
		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			httputil.JSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"error":       "too many login attempts for this account",
				"retry_after": retryAfter,
			})
			return
		}
	}

	result, err := h.loginCmd.Execute(r.Context(), command.LoginInput{
		Username:     req.Username,
		Password:     req.Password,
		TOTPCode:     req.TOTPCode,
		RecoveryCode: req.RecoveryCode,
		MFAToken:     req.MFAToken,
		IPAddress:    command.ExtractIPAddress(r),
		UserAgent:    r.UserAgent(),
	})
	if err != nil {
		handleAuthError(w, err)
		return
	}

	// MFA challenge response
	if result.MFARequired {
		httputil.OK(w, response.LoginResponse{
			MFARequired: true,
			MFAToken:    result.MFAToken,
		})
		return
	}

	h.resetLoginRateLimits(r.Context(), clientIP, req.Username)

	httputil.OK(w, response.LoginResponse{
		AccessToken:           result.AccessToken,
		AccessTokenExpiresAt:  result.AccessTokenExpiresAt,
		RefreshToken:          result.RefreshToken,
		RefreshTokenExpiresAt: result.RefreshTokenExpiresAt,
		ForcePasswordChange:   result.ForcePasswordChange,
		MFAEnrollmentRequired: result.MFAEnrollmentRequired,
		RestrictedSession:     result.RestrictedSession,
		User: response.UserResponse{
			ID:          result.User.ID,
			Username:    result.User.Username,
			DisplayName: result.User.DisplayName,
			Email:       result.User.Email,
			Groups:      result.User.Groups,
		},
		Permissions: response.PermissionsResp{
			Functions: result.Permissions.Functions,
			Contracts: result.Permissions.Contracts,
		},
	})
}

func (h *AuthHandler) resetLoginRateLimits(ctx context.Context, clientIP string, username string) {
	if h.loginLimiter != nil {
		key := fmt.Sprintf("rl:%s:%s", h.loginPolicy.Name, clientIP)
		_ = h.loginLimiter.Reset(ctx, key)
	}
	if h.loginUserLimiter != nil {
		key := fmt.Sprintf("rl:%s:%s:%s", h.loginUserPolicy.Name, clientIP, strings.ToLower(username))
		_ = h.loginUserLimiter.Reset(ctx, key)
	}
}

// RefreshToken handles POST /auth/refresh.
// @Summary Refresh Token
// @Description Issue a new access token using a refresh token (enforces idle and absolute timeouts)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.RefreshRequest true "Refresh Token"
// @Success 200 {object} response.RefreshResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req request.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}

	if fieldErrors := validation.ValidateStruct(&req); fieldErrors != nil {
		httputil.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "validation failed",
			"details": fieldErrors,
		})
		return
	}

	// Per-token rate limiting — prevents token stuffing on stolen refresh tokens
	if h.refreshTokenLimiter != nil {
		tokenHash := application.HashRefreshToken(req.RefreshToken)
		key := fmt.Sprintf("rl:refresh_token:%s", tokenHash)
		allowed, _, retryAfter, err := h.refreshTokenLimiter.Allow(r.Context(), key)
		if err != nil {
			httputil.InternalError(w, "rate limit check failed")
			return
		}
		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			httputil.JSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"error":       "too many refresh attempts for this token",
				"retry_after": retryAfter,
			})
			return
		}
	}

	result, err := h.refreshCmd.Execute(r.Context(), command.RefreshInput{
		RefreshToken: req.RefreshToken,
		IPAddress:    command.ExtractIPAddress(r),
		UserAgent:    r.UserAgent(),
	})
	if err != nil {
		handleAuthError(w, err)
		return
	}

	httputil.OK(w, response.RefreshResponse{
		AccessToken:           result.AccessToken,
		AccessTokenExpiresAt:  result.AccessTokenExpiresAt,
		RefreshToken:          result.RefreshToken,
		RefreshTokenExpiresAt: result.RefreshTokenExpiresAt,
	})
}

// Logout handles POST /auth/logout.
// @Summary User Logout
// @Description Revoke a single refresh token session
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.LogoutRequest true "Refresh Token"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req request.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.BadRequest(w, "invalid user ID in token")
		return
	}

	if err := h.logoutCmd.Execute(r.Context(), command.LogoutInput{
		RefreshToken: req.RefreshToken,
		UserID:       userID,
		IPAddress:    command.ExtractIPAddress(r),
		UserAgent:    r.UserAgent(),
	}); err != nil {
		httputil.InternalError(w, "failed to logout")
		return
	}

	httputil.NoContent(w)
}

// LogoutAll handles POST /auth/logout-all.
// @Summary Logout All Sessions
// @Description Revoke all refresh token sessions globally for the user
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 204 "No Content"
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.BadRequest(w, "invalid user ID in token")
		return
	}

	if err := h.logoutAllCmd.Execute(r.Context(), command.LogoutAllInput{
		UserID:    userID,
		IPAddress: command.ExtractIPAddress(r),
		UserAgent: r.UserAgent(),
	}); err != nil {
		httputil.InternalError(w, "failed to logout all sessions")
		return
	}

	httputil.NoContent(w)
}

// ChangePassword handles POST /auth/change-password.
// @Summary Change Password
// @Description Change active user's password and revoke active sessions
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.ChangePasswordRequest true "Change Password Payload"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.BadRequest(w, "invalid user ID in token")
		return
	}

	var req request.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	if fieldErrors := validation.ValidateStruct(&req); fieldErrors != nil {
		httputil.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "validation failed",
			"details": fieldErrors,
		})
		return
	}

	if err := h.changePasswordCmd.Execute(r.Context(), command.ChangePasswordInput{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
		IPAddress:   command.ExtractIPAddress(r),
		UserAgent:   r.UserAgent(),
	}); err != nil {
		handleAuthError(w, err)
		return
	}

	httputil.NoContent(w)
}

// GetMe handles GET /auth/me.
// @Summary Get Current User
// @Description Retrieve the profile and permissions of the currently authenticated user
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.MeResponse
// @Failure 401 {object} map[string]interface{}
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	result, err := h.getMeQry.Execute(r.Context(), claims.Subject)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	httputil.OK(w, response.MeResponse{
		User: response.UserResponse{
			ID:          result.User.ID,
			Username:    result.User.Username,
			DisplayName: result.User.DisplayName,
			Email:       result.User.Email,
			Groups:      result.User.Groups,
		},
		Permissions: response.PermissionsResp{
			Functions: result.Permissions.Functions,
			Contracts: result.Permissions.Contracts,
		},
	})
}

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
