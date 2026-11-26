package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/domain"
)

// fakeRepo is a tiny in-test PerformanceRepository.
type fakeRepo struct {
	records []domain.FunnelRecord
	err     error
	called  bool
}

func (f *fakeRepo) ListFunnelRecords(_ context.Context, _ Filter) ([]domain.FunnelRecord, error) {
	f.called = true
	return f.records, f.err
}

func (f *fakeRepo) Source() string { return "fake" }

func TestGetFunnel(t *testing.T) {
	oneRecord := []domain.FunnelRecord{{
		Programme:  domain.OGX,
		PhaseDates: map[domain.Phase]time.Time{domain.SU: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
	}}
	repoErr := errors.New("boom")

	tests := []struct {
		name         string
		programme    domain.Programme
		start, end   string
		repo         *fakeRepo
		wantErr      bool
		wantRepoCall bool
		wantSU       int
	}{
		{"ok", domain.OGX, "2026-01-01", "2026-12-31", &fakeRepo{records: oneRecord}, false, true, 1},
		{"invalid programme", "XYZ", "2026-01-01", "2026-12-31", &fakeRepo{}, true, false, 0},
		{"reversed dates are rejected before hitting the repo", domain.OGX, "2026-12-31", "2026-01-01", &fakeRepo{}, true, false, 0},
		{"bad date format", domain.OGX, "2026/01/01", "2026-12-31", &fakeRepo{}, true, false, 0},
		{"repository error is returned", domain.OGX, "2026-01-01", "2026-12-31", &fakeRepo{err: repoErr}, true, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewFunnelService(tt.repo)
			f, err := svc.GetFunnel(context.Background(), tt.programme, tt.start, tt.end)

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.repo.called != tt.wantRepoCall {
				t.Errorf("repo called = %v, want %v", tt.repo.called, tt.wantRepoCall)
			}
			if err != nil {
				return
			}
			if f.DataSource != "fake" {
				t.Errorf("DataSource = %q, want %q", f.DataSource, "fake")
			}
			if f.Phases[0].Count != tt.wantSU {
				t.Errorf("SU = %d, want %d", f.Phases[0].Count, tt.wantSU)
			}
		})
	}
}
