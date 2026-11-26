package mock

import (
	"context"
	"testing"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/domain"
	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/service"
)

func TestSeedLoadsAndFilters(t *testing.T) {
	repo, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if len(repo.records) == 0 {
		t.Fatal("seed.json is empty; run: go run ./cmd/seed")
	}

	tests := []struct {
		name       string
		programme  domain.Programme
		start, end string
		wantAny    bool
	}{
		{"oGX in 2026 has data", domain.OGX, "2026-01-01", "2026-11-26", true},
		{"iCX in 2026 has data", domain.ICX, "2026-01-01", "2026-11-26", true},
		{"before the data window is empty", domain.OGX, "2025-01-01", "2025-12-31", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := domain.NewDateRange(tt.start, tt.end)
			got, err := repo.ListFunnelRecords(context.Background(), service.Filter{Programme: tt.programme, Range: r})
			if err != nil {
				t.Fatal(err)
			}
			if (len(got) > 0) != tt.wantAny {
				t.Fatalf("got %d records, wantAny %v", len(got), tt.wantAny)
			}
			for _, rec := range got {
				if rec.Programme != tt.programme {
					t.Fatalf("record %s has programme %s, want %s", rec.ID, rec.Programme, tt.programme)
				}
			}
		})
	}
}

func TestSeedFunnelShapeIsRealistic(t *testing.T) {
	repo, err := New()
	if err != nil {
		t.Fatal(err)
	}
	svc := service.NewFunnelService(repo)
	f, err := svc.GetFunnel(context.Background(), domain.OGX, "2026-01-01", "2026-11-26")
	if err != nil {
		t.Fatal(err)
	}
	// Over the full window each phase should be <= the one before it.
	for i := 1; i < len(f.Phases); i++ {
		if f.Phases[i].Count > f.Phases[i-1].Count {
			t.Errorf("%s (%d) > %s (%d)", f.Phases[i].Phase, f.Phases[i].Count, f.Phases[i-1].Phase, f.Phases[i-1].Count)
		}
	}
	if f.DataSource != "mock" {
		t.Errorf("DataSource = %q, want mock", f.DataSource)
	}
}
