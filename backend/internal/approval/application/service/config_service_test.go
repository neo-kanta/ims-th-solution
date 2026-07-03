package service

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

func singleUserStage(n int, isFinal bool) StageInput {
	uid := uuid.New()
	return StageInput{
		StageNumber:           n,
		StageName:             "Stage " + strings.Repeat("I", n),
		ApproverMode:          vo.ApproverModeSingleUser,
		ApproverUserID:        &uid,
		RequiredApprovalCount: 1,
		IsFinalStage:          isFinal,
	}
}

func TestValidateStages_Empty_ReturnsError(t *testing.T) {
	t.Parallel()
	_, err := validateStages(nil)
	if err == nil {
		t.Fatal("expected error for empty stages, got nil")
	}
}

func TestValidateStages_TooMany_ReturnsError(t *testing.T) {
	t.Parallel()
	stages := []StageInput{
		singleUserStage(1, false),
		singleUserStage(2, false),
		singleUserStage(3, false),
		singleUserStage(4, true), // 4 > max of 3
	}
	_, err := validateStages(stages)
	if err == nil {
		t.Fatal("expected error for > 3 stages, got nil")
	}
}

func TestValidateStages_SingleStage_NoFinalMarked_ReturnsError(t *testing.T) {
	t.Parallel()
	stages := []StageInput{singleUserStage(1, false)} // IsFinalStage intentionally false
	_, err := validateStages(stages)
	if err == nil {
		t.Fatal("expected error when no stage is marked as is_final_stage, got nil")
	}
}

func TestValidateStages_SingleStage_ExplicitFinal_Passes(t *testing.T) {
	t.Parallel()
	stages := []StageInput{singleUserStage(1, true)}
	out, err := validateStages(stages)
	if err != nil {
		t.Fatalf("unexpected error for single stage explicitly marked final: %v", err)
	}
	if len(out) != 1 || !out[0].IsFinalStage {
		t.Fatal("single stage marked final must pass validation")
	}
}

func TestValidateStages_MultipleFinals_ReturnsError(t *testing.T) {
	t.Parallel()
	stages := []StageInput{
		singleUserStage(1, true), // marked final
		singleUserStage(2, true), // also marked final — invalid
	}
	_, err := validateStages(stages)
	if err == nil {
		t.Fatal("expected error when multiple stages marked as final, got nil")
	}
	if !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("error message should mention 'exactly one', got: %v", err)
	}
}

func TestValidateStages_FinalNotOnHighestStage_ReturnsError(t *testing.T) {
	t.Parallel()
	stages := []StageInput{
		singleUserStage(1, true), // final on stage 1, but stage 2 exists — invalid
		singleUserStage(2, false),
	}
	_, err := validateStages(stages)
	if err == nil {
		t.Fatal("expected error when final is not on the highest stage, got nil")
	}
	if !strings.Contains(err.Error(), "highest stage_number") {
		t.Fatalf("error message should mention 'highest stage_number', got: %v", err)
	}
}

func TestValidateStages_ValidTwoStage_FinalOnHighest(t *testing.T) {
	t.Parallel()
	stages := []StageInput{
		singleUserStage(1, false),
		singleUserStage(2, true), // correct: final on the highest stage
	}
	out, err := validateStages(stages)
	if err != nil {
		t.Fatalf("unexpected error for valid two-stage config: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(out))
	}
	for _, s := range out {
		if s.StageNumber == 2 && !s.IsFinalStage {
			t.Fatal("stage 2 must be marked as final")
		}
		if s.StageNumber == 1 && s.IsFinalStage {
			t.Fatal("stage 1 must not be marked as final")
		}
	}
}

func TestValidateStages_ThreeStage_NoFinalMarked_ReturnsError(t *testing.T) {
	t.Parallel()
	stages := []StageInput{
		singleUserStage(1, false),
		singleUserStage(2, false),
		singleUserStage(3, false), // none marked final — must now be rejected
	}
	_, err := validateStages(stages)
	if err == nil {
		t.Fatal("expected error when no stage is marked as is_final_stage, got nil")
	}
}

func TestValidateStages_DuplicateStageNumber_ReturnsError(t *testing.T) {
	t.Parallel()
	stages := []StageInput{
		singleUserStage(1, false),
		singleUserStage(1, true), // duplicate stage_number 1
	}
	_, err := validateStages(stages)
	if err == nil {
		t.Fatal("expected error for duplicate stage_number, got nil")
	}
}

func TestValidateStages_SingleUserModeRequiresUserID(t *testing.T) {
	t.Parallel()
	stages := []StageInput{
		{
			StageNumber:  1,
			StageName:    "Review",
			ApproverMode: vo.ApproverModeSingleUser,
			// ApproverUserID intentionally nil
			IsFinalStage: true,
		},
	}
	_, err := validateStages(stages)
	if err == nil {
		t.Fatal("expected error when SINGLE_USER mode has no approver_user_id")
	}
}
