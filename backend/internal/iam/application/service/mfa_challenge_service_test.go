package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

func TestMFAChallengeService_IssueAndValidate(t *testing.T) {
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	svc := service.NewMFAChallengeService("test-secret", clk)
	userID := uuid.New()

	token, _, err := svc.Issue(userID, "alice")
	if err != nil {
		t.Fatalf("issue challenge: %v", err)
	}
	if err := svc.Validate(token, userID, "alice"); err != nil {
		t.Fatalf("validate challenge: %v", err)
	}
}

func TestMFAChallengeService_RejectsUserMismatch(t *testing.T) {
	clk := clock.FixedClock{FixedTime: time.Now().UTC()}
	svc := service.NewMFAChallengeService("test-secret", clk)
	userID := uuid.New()

	token, _, err := svc.Issue(userID, "alice")
	if err != nil {
		t.Fatalf("issue challenge: %v", err)
	}
	if err := svc.Validate(token, uuid.New(), "alice"); err == nil {
		t.Fatal("expected user mismatch to fail")
	}
}
