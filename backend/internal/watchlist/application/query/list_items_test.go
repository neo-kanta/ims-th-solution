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

type captureItemRepo struct {
	capturedFilter repository.WatchlistItemFilter
	items          []*entity.WatchlistItem
}

func (f *captureItemRepo) Create(_ context.Context, _ pgx.Tx, _ *entity.WatchlistItem) error {
	return nil
}
func (f *captureItemRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.WatchlistItem, error) {
	return nil, nil
}
func (f *captureItemRepo) Update(_ context.Context, _ pgx.Tx, _ *entity.WatchlistItem) error {
	return nil
}
func (f *captureItemRepo) SoftDelete(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (f *captureItemRepo) List(_ context.Context, filter repository.WatchlistItemFilter) ([]*entity.WatchlistItem, int, error) {
	f.capturedFilter = filter
	return f.items, len(f.items), nil
}

type listItemsPortfolioPort struct{}

func (f *listItemsPortfolioPort) GetPortfolioScope(_ context.Context, _ uuid.UUID) (*watchlistdomain.PortfolioScopeInfo, error) {
	return nil, nil
}

type listItemsIAMChecker struct {
	fundIDs     []string
	hasWildcard bool
}

func (f *listItemsIAMChecker) HasFunctionPermission(_ uuid.UUID, _ string) (bool, error) {
	return true, nil
}
func (f *listItemsIAMChecker) HasDataPermission(_ uuid.UUID, _ string) (bool, error) {
	return true, nil
}
func (f *listItemsIAMChecker) GetAccessibleContracts(_ uuid.UUID) ([]string, error) {
	if f.hasWildcard {
		return []string{"*"}, nil
	}
	return f.fundIDs, nil
}

// ── Tests ────────────────────────────────────────────────────────────────────

// Test 10 (P2-1): Unscoped list sets UnionPersonalOwnerID and FundIDs on filter.
func TestListItems_Unscoped_SetsUnionFields(t *testing.T) {
	actorID := uuid.New()
	fundID := uuid.New()

	repo := &captureItemRepo{}
	iam := &listItemsIAMChecker{fundIDs: []string{fundID.String()}}
	h := NewListItemsHandler(repo, &listItemsPortfolioPort{}, iam)

	_, err := h.Handle(context.Background(), ListItemsInput{
		ActorID: actorID,
		// ScopeType nil → unscoped
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := repo.capturedFilter
	if f.UnionPersonalOwnerID == nil {
		t.Error("UnionPersonalOwnerID should be set for unscoped query")
	} else if *f.UnionPersonalOwnerID != actorID {
		t.Errorf("UnionPersonalOwnerID = %s, want %s", *f.UnionPersonalOwnerID, actorID)
	}
	if len(f.FundIDs) != 1 || f.FundIDs[0] != fundID {
		t.Errorf("FundIDs = %v, want [%s]", f.FundIDs, fundID)
	}
	if f.UnionPortfolioAll {
		t.Error("UnionPortfolioAll should be false when actor has specific fund access")
	}
}

// Test 10b: Wildcard access sets UnionPortfolioAll.
func TestListItems_Unscoped_WildcardSetsPortfolioAll(t *testing.T) {
	actorID := uuid.New()

	repo := &captureItemRepo{}
	iam := &listItemsIAMChecker{hasWildcard: true}
	h := NewListItemsHandler(repo, &listItemsPortfolioPort{}, iam)

	if _, err := h.Handle(context.Background(), ListItemsInput{ActorID: actorID}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := repo.capturedFilter
	if f.UnionPersonalOwnerID == nil {
		t.Error("UnionPersonalOwnerID should be set")
	}
	if !f.UnionPortfolioAll {
		t.Error("UnionPortfolioAll should be true for wildcard access")
	}
}

// Explicit PERSONAL scope only sets OwnerUserID (not UnionPersonalOwnerID).
func TestListItems_PersonalScope_SetsOwnerUserID(t *testing.T) {
	actorID := uuid.New()
	repo := &captureItemRepo{}
	iam := &listItemsIAMChecker{}
	h := NewListItemsHandler(repo, &listItemsPortfolioPort{}, iam)

	personal := entity.ScopePersonal
	if _, err := h.Handle(context.Background(), ListItemsInput{ActorID: actorID, ScopeType: &personal}); err != nil {
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

func TestListItems_PortfolioIDWithoutPortfolioScope_ReturnsInvalidQuery(t *testing.T) {
	actorID := uuid.New()
	portfolioID := uuid.New()
	repo := &captureItemRepo{}
	iam := &listItemsIAMChecker{}
	h := NewListItemsHandler(repo, &listItemsPortfolioPort{}, iam)

	_, err := h.Handle(context.Background(), ListItemsInput{ActorID: actorID, PortfolioID: &portfolioID})
	if !errors.Is(err, watchlistdomain.ErrInvalidQuery) {
		t.Fatalf("err = %v, want ErrInvalidQuery", err)
	}
	if repo.capturedFilter.Limit != 0 {
		t.Errorf("repository should not be called for invalid portfolio query; filter=%+v", repo.capturedFilter)
	}
}

func TestListItems_PortfolioIDWithPersonalScope_ReturnsInvalidQuery(t *testing.T) {
	actorID := uuid.New()
	portfolioID := uuid.New()
	personal := entity.ScopePersonal
	repo := &captureItemRepo{}
	iam := &listItemsIAMChecker{}
	h := NewListItemsHandler(repo, &listItemsPortfolioPort{}, iam)

	_, err := h.Handle(context.Background(), ListItemsInput{ActorID: actorID, ScopeType: &personal, PortfolioID: &portfolioID})
	if !errors.Is(err, watchlistdomain.ErrInvalidQuery) {
		t.Fatalf("err = %v, want ErrInvalidQuery", err)
	}
}
