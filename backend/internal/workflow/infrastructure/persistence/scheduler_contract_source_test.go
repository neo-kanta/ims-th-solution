package persistence

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFilterActiveSchedulerContractIDs_ReturnsBrandNewActiveContract(t *testing.T) {
	contractID := uuid.New()
	records := []schedulerContractRecord{{
		ContractID:    contractID,
		IsActive:      true,
		EffectiveFrom: dateUTC(2026, 1, 1),
	}}

	got := filterActiveSchedulerContractIDs(records, dateUTC(2026, 4, 24))
	if len(got) != 1 || got[0] != contractID {
		t.Fatalf("expected brand-new active contract %s, got %v", contractID, got)
	}
}

func TestFilterActiveSchedulerContractIDs_ExcludesInactiveAndStaleContracts(t *testing.T) {
	activeID := uuid.New()
	inactiveID := uuid.New()
	staleID := uuid.New()
	futureID := uuid.New()
	staleTo := dateUTC(2026, 1, 31)
	records := []schedulerContractRecord{
		{ContractID: activeID, IsActive: true, EffectiveFrom: dateUTC(2026, 1, 1)},
		{ContractID: inactiveID, IsActive: false, EffectiveFrom: dateUTC(2026, 1, 1)},
		{ContractID: staleID, IsActive: true, EffectiveFrom: dateUTC(2026, 1, 1), EffectiveTo: &staleTo},
		{ContractID: futureID, IsActive: true, EffectiveFrom: dateUTC(2026, 5, 1)},
	}

	got := filterActiveSchedulerContractIDs(records, dateUTC(2026, 4, 24))
	if len(got) != 1 || got[0] != activeID {
		t.Fatalf("expected only active contract %s, got %v", activeID, got)
	}
}

func dateUTC(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
