package query

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
)

// ── Fakes ────────────────────────────────────────────────────────────────────

type captureAlertRepo struct {
	capturedFilter repository.AlertEventFilter
	alerts         []*entity.AlertEvent
}

func (f *captureAlertRepo) Insert(_ context.Context, _ pgx.Tx, _ *entity.AlertEvent) error {
	return nil
}
func (f *captureAlertRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.AlertEvent, error) {
	return nil, nil
}
func (f *captureAlertRepo) UpdateNotificationStatus(_ context.Context, _ uuid.UUID, _ entity.NotificationStatus, _ *string) error {
	return nil
}
func (f *captureAlertRepo) Acknowledge(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ uuid.UUID, _ *string) error {
	return nil
}
func (f *captureAlertRepo) List(_ context.Context, filter repository.AlertEventFilter) ([]*entity.AlertEvent, int, error) {
	f.capturedFilter = filter
	return f.alerts, len(f.alerts), nil
}

type listAlertsPortfolioPort struct{}

func (f *listAlertsPortfolioPort) GetPortfolioScope(_ context.Context, _ uuid.UUID) (*watchlistdomain.PortfolioScopeInfo, error) {
	return nil, nil
}

type listAlertsIAMChecker struct {
	fundIDs     []string
	hasWildcard bool
}

func (f *listAlertsIAMChecker) HasFunctionPermission(_ uuid.UUID, _ string) (bool, error) {
	return true, nil
}
func (f *listAlertsIAMChecker) HasDataPermission(_ uuid.UUID, _ string) (bool, error) {
	return true, nil
}
func (f *listAlertsIAMChecker) GetAccessibleContracts(_ uuid.UUID) ([]string, error) {
	if f.hasWildcard {
		return []string{"*"}, nil
	}
	return f.fundIDs, nil
}

// ── Tests ────────────────────────────────────────────────────────────────────

// Test 11 (P2-2): Unscoped alerts query sets UnionPersonalOwnerID and FundIDs.
func TestListAlerts_Unscoped_SetsUnionFields(t *testing.T) {
	actorID := uuid.New()
	fundID := uuid.New()

	repo := &captureAlertRepo{}
	iam := &listAlertsIAMChecker{fundIDs: []string{fundID.String()}}
	h := NewListAlertsHandler(repo, &listAlertsPortfolioPort{}, iam)

	_, err := h.Handle(context.Background(), ListAlertsInput{
		ActorID: actorID,
		// ScopeType nil → unscoped
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := repo.capturedFilter
	if f.UnionPersonalOwnerID == nil {
		t.Error("UnionPersonalOwnerID should be set for unscoped alert query")
	} else if *f.UnionPersonalOwnerID != actorID {
		t.Errorf("UnionPersonalOwnerID = %s, want %s", *f.UnionPersonalOwnerID, actorID)
	}
	if len(f.FundIDs) != 1 || f.FundIDs[0] != fundID {
		t.Errorf("FundIDs = %v, want [%s]", f.FundIDs, fundID)
	}
	if f.UnionPortfolioAll {
		t.Error("UnionPortfolioAll should be false for specific fund access")
	}
}

// Wildcard access sets UnionPortfolioAll.
func TestListAlerts_Unscoped_WildcardSetsPortfolioAll(t *testing.T) {
	actorID := uuid.New()
	repo := &captureAlertRepo{}
	iam := &listAlertsIAMChecker{hasWildcard: true}
	h := NewListAlertsHandler(repo, &listAlertsPortfolioPort{}, iam)

	if _, err := h.Handle(context.Background(), ListAlertsInput{ActorID: actorID}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := repo.capturedFilter
	if !f.UnionPortfolioAll {
		t.Error("UnionPortfolioAll should be true for wildcard access")
	}
	if f.UnionPersonalOwnerID == nil {
		t.Error("UnionPersonalOwnerID should be set")
	}
}

// Explicit PERSONAL scope sets OwnerUserID, not the union field.
func TestListAlerts_PersonalScope_SetsOwnerUserID(t *testing.T) {
	actorID := uuid.New()
	repo := &captureAlertRepo{}
	iam := &listAlertsIAMChecker{}
	h := NewListAlertsHandler(repo, &listAlertsPortfolioPort{}, iam)

	personal := entity.ScopePersonal
	if _, err := h.Handle(context.Background(), ListAlertsInput{ActorID: actorID, ScopeType: &personal}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := repo.capturedFilter
	if f.UnionPersonalOwnerID != nil {
		t.Error("UnionPersonalOwnerID should be nil for explicit PERSONAL scope")
	}
	if f.OwnerUserID == nil || *f.OwnerUserID != actorID {
		t.Errorf("OwnerUserID = %v, want %s", f.OwnerUserID, actorID)
	}
}

func TestListAlerts_PortfolioIDWithoutPortfolioScope_ReturnsInvalidQuery(t *testing.T) {
	actorID := uuid.New()
	portfolioID := uuid.New()
	repo := &captureAlertRepo{}
	iam := &listAlertsIAMChecker{}
	h := NewListAlertsHandler(repo, &listAlertsPortfolioPort{}, iam)

	_, err := h.Handle(context.Background(), ListAlertsInput{ActorID: actorID, PortfolioID: &portfolioID})
	if !errors.Is(err, watchlistdomain.ErrInvalidQuery) {
		t.Fatalf("err = %v, want ErrInvalidQuery", err)
	}
	if repo.capturedFilter.Limit != 0 {
		t.Errorf("repository should not be called for invalid portfolio query; filter=%+v", repo.capturedFilter)
	}
}

func TestListAlerts_PortfolioIDWithPersonalScope_ReturnsInvalidQuery(t *testing.T) {
	actorID := uuid.New()
	portfolioID := uuid.New()
	personal := entity.ScopePersonal
	repo := &captureAlertRepo{}
	iam := &listAlertsIAMChecker{}
	h := NewListAlertsHandler(repo, &listAlertsPortfolioPort{}, iam)

	_, err := h.Handle(context.Background(), ListAlertsInput{ActorID: actorID, ScopeType: &personal, PortfolioID: &portfolioID})
	if !errors.Is(err, watchlistdomain.ErrInvalidQuery) {
		t.Fatalf("err = %v, want ErrInvalidQuery", err)
	}
}
