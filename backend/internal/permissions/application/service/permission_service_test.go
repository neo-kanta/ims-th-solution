package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/permissions/domain"
)

type roleRepo struct {
	domain.Repository
	roles []domain.Role
}

func (r roleRepo) GetUserApprovedRoles(context.Context, uuid.UUID) ([]domain.Role, error) {
	return r.roles, nil
}

func TestValidateRoleAssignmentPriorityAndScope(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	tests := []struct {
		name    string
		actor   domain.Role
		target  domain.Role
		wantErr bool
	}{
		{
			name: "higher priority within scope can assign",
			actor: domain.Role{
				RoleCode:                 "HEAD_INVESTMENT",
				PriorityRank:             90,
				AssignmentScope:          "INVESTMENT",
				CanRequestRoleAssignment: true,
			},
			target: domain.Role{RoleCode: "PORTFOLIO_MANAGER", PriorityRank: 75, AssignmentScope: "INVESTMENT"},
		},
		{
			name: "same priority cannot assign",
			actor: domain.Role{
				RoleCode:                 "COMPLIANCE_OFFICER",
				PriorityRank:             85,
				AssignmentScope:          "COMPLIANCE",
				CanRequestRoleAssignment: true,
			},
			target:  domain.Role{RoleCode: "RISK_MANAGER", PriorityRank: 85, AssignmentScope: "RISK"},
			wantErr: true,
		},
		{
			name: "lower priority cannot assign higher priority",
			actor: domain.Role{
				RoleCode:                 "PORTFOLIO_MANAGER",
				PriorityRank:             75,
				AssignmentScope:          "INVESTMENT",
				CanRequestRoleAssignment: true,
			},
			target:  domain.Role{RoleCode: "HEAD_INVESTMENT", PriorityRank: 90, AssignmentScope: "INVESTMENT"},
			wantErr: true,
		},
		{
			name: "out of scope is denied",
			actor: domain.Role{
				RoleCode:                 "TRADING_SUPERVISOR",
				PriorityRank:             80,
				AssignmentScope:          "TRADING",
				CanRequestRoleAssignment: true,
			},
			target:  domain.Role{RoleCode: "PORTFOLIO_MANAGER", PriorityRank: 75, AssignmentScope: "INVESTMENT"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{repo: roleRepo{roles: []domain.Role{tt.actor}}}
			err := svc.validateRoleAssignment(context.Background(), actorID, &tt.target)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestHighRiskPermissionDetection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code  string
		after map[string]any
		want  bool
	}{
		{code: "permission.audit.export", want: true},
		{code: "permission.change_request.approve", want: true},
		{code: "permission.notification.edit", after: map[string]any{"can_configure": true}, want: true},
		{code: "permission.users.view", after: map[string]any{"can_view": true}, want: false},
	}

	for _, tc := range cases {
		if got := isHighRiskPermission(tc.code, tc.after); got != tc.want {
			t.Fatalf("isHighRiskPermission(%q)=%v, want %v", tc.code, got, tc.want)
		}
	}
}

func TestBlockingChecks(t *testing.T) {
	t.Parallel()

	checks := []domain.Check{
		{Status: domain.CheckStatusPassed, Severity: domain.CheckSeverityInfo},
		{Status: domain.CheckStatusWarning, Severity: domain.CheckSeverityWarning},
	}
	if hasBlocking(checks) {
		t.Fatalf("non-failed checks should not block")
	}

	checks = append(checks, domain.Check{Status: domain.CheckStatusFailed, Severity: domain.CheckSeverityBlocker})
	if !hasBlocking(checks) {
		t.Fatalf("failed blocker check should block")
	}
}
