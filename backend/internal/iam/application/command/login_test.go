package command

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application"
	appservice "github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

type loginTestUserRepo struct {
	user *entity.User
}

func (r *loginTestUserRepo) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	if r.user != nil && r.user.Username == username {
		copyUser := *r.user
		return &copyUser, nil
	}
	return nil, nil
}
func (r *loginTestUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) { return nil, nil }
func (r *loginTestUserRepo) Create(ctx context.Context, user *entity.User) error               { return nil }
func (r *loginTestUserRepo) Update(ctx context.Context, user *entity.User) error               { return nil }
func (r *loginTestUserRepo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	return nil
}
func (r *loginTestUserRepo) List(ctx context.Context, filter domain.UserFilter) ([]entity.User, int, error) {
	return nil, 0, nil
}

type loginTestSessionRepo struct{}

func (r *loginTestSessionRepo) Create(ctx context.Context, session *entity.Session) error { return nil }
func (r *loginTestSessionRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error) {
	return nil, nil
}
func (r *loginTestSessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Session, error) {
	return nil, nil
}
func (r *loginTestSessionRepo) RevokeByID(ctx context.Context, id uuid.UUID) error                         { return nil }
func (r *loginTestSessionRepo) RevokeByIDWithReason(ctx context.Context, id uuid.UUID, reason string) error { return nil }
func (r *loginTestSessionRepo) RevokeByFamily(ctx context.Context, family uuid.UUID) error                  { return nil }
func (r *loginTestSessionRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error                { return nil }
func (r *loginTestSessionRepo) RevokeAllForUserWithReason(ctx context.Context, userID uuid.UUID, reason string) error {
	return nil
}
func (r *loginTestSessionRepo) DeleteExpired(ctx context.Context) (int64, error)                   { return 0, nil }
func (r *loginTestSessionRepo) ListActiveForUser(ctx context.Context, userID uuid.UUID) ([]entity.Session, error) {
	return nil, nil
}
func (r *loginTestSessionRepo) CountActiveForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}
func (r *loginTestSessionRepo) RevokeOldestForUser(ctx context.Context, userID uuid.UUID, reason string) error {
	return nil
}
func (r *loginTestSessionRepo) UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error { return nil }

type loginTestMFARepo struct {
	enrollment *entity.MFAEnrollment
}

func (r *loginTestMFARepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.MFAEnrollment, error) {
	return r.enrollment, nil
}
func (r *loginTestMFARepo) CreateEnrollment(ctx context.Context, enrollment *entity.MFAEnrollment) error {
	return nil
}
func (r *loginTestMFARepo) EnableEnrollment(ctx context.Context, userID uuid.UUID) error  { return nil }
func (r *loginTestMFARepo) DisableEnrollment(ctx context.Context, userID uuid.UUID) error { return nil }
func (r *loginTestMFARepo) DeleteEnrollment(ctx context.Context, userID uuid.UUID) error  { return nil }
func (r *loginTestMFARepo) StoreRecoveryCodes(ctx context.Context, codes []entity.MFARecoveryCode) error {
	return nil
}
func (r *loginTestMFARepo) FindUnusedRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]entity.MFARecoveryCode, error) {
	return nil, nil
}
func (r *loginTestMFARepo) UseRecoveryCode(ctx context.Context, codeID uuid.UUID) error     { return nil }
func (r *loginTestMFARepo) DeleteRecoveryCodes(ctx context.Context, userID uuid.UUID) error { return nil }

type loginTestPerms struct {
	functions []string
	contracts []string
}

func (p *loginTestPerms) GetUserFunctionPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return p.functions, nil
}
func (p *loginTestPerms) GetUserDataPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return p.contracts, nil
}

type loginTestAuditRepo struct{}

func (r *loginTestAuditRepo) Record(ctx context.Context, event *entity.AuditEvent) error { return nil }
func (r *loginTestAuditRepo) List(ctx context.Context, filter domain.AuditFilter) ([]entity.AuditEvent, int, error) {
	return nil, 0, nil
}

func TestLoginCommand_ReturnsMFAChallengeWhenRequired(t *testing.T) {
	hash, err := valueobject.HashPassword("StrongPass123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	user := &entity.User{ID: uuid.New(), Username: "alice", PasswordHash: hash, IsActive: true}
	cmd := NewLoginCommand(
		&loginTestUserRepo{user: user},
		&loginTestSessionRepo{},
		&loginTestMFARepo{enrollment: &entity.MFAEnrollment{UserID: user.ID, IsVerified: true, IsEnabled: true}},
		&loginTestPerms{},
		application.NewTokenService("test-secret", clk),
		nil,
		appservice.NewMFAChallengeService("test-secret", clk),
		appservice.NewAuditService(&loginTestAuditRepo{}),
		clk,
		SessionPolicy{},
		false,
	)

	result, err := cmd.Execute(context.Background(), LoginInput{
		Username: "alice",
		Password: "StrongPass123",
	})
	if err != nil {
		t.Fatalf("execute login: %v", err)
	}
	if !result.MFARequired || result.MFAToken == "" {
		t.Fatal("expected MFA challenge token")
	}
}

func TestLoginCommand_BlocksPrivilegedUserWithoutMFAWhenRequired(t *testing.T) {
	hash, err := valueobject.HashPassword("StrongPass123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	user := &entity.User{ID: uuid.New(), Username: "admin", PasswordHash: hash, IsActive: true}
	cmd := NewLoginCommand(
		&loginTestUserRepo{user: user},
		&loginTestSessionRepo{},
		&loginTestMFARepo{},
		&loginTestPerms{functions: []string{"IAM_USER_UPDATE"}},
		application.NewTokenService("test-secret", clk),
		nil,
		appservice.NewMFAChallengeService("test-secret", clk),
		appservice.NewAuditService(&loginTestAuditRepo{}),
		clk,
		SessionPolicy{},
		true,
	)

	result, err := cmd.Execute(context.Background(), LoginInput{
		Username: "admin",
		Password: "StrongPass123",
	})
	if err == nil || result != nil {
		t.Fatal("expected privileged login without MFA enrollment to fail")
	}
}
