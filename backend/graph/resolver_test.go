package graph

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/graph/model"
	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/domain"
)

type fakeFunnels struct {
	funnel domain.Funnel
	err    error
}

func (f fakeFunnels) GetFunnel(context.Context, domain.Programme, string, string) (domain.Funnel, error) {
	return f.funnel, f.err
}

func TestFunnelResolver(t *testing.T) {
	r, _ := domain.NewDateRange("2026-01-01", "2026-01-31")
	ok := domain.BuildFunnel(domain.OGX, r, nil)
	ok.DataSource = "mock"

	tests := []struct {
		name        string
		fake        fakeFunnels
		wantErr     string // substring; "" = no error
		mustNotLeak string
	}{
		{name: "maps domain to model", fake: fakeFunnels{funnel: ok}},
		{
			name:    "input error is shown to client",
			fake:    fakeFunnels{err: domain.ErrInvalidDateRange},
			wantErr: "invalid date range",
		},
		{
			name:        "internal error is hidden from client",
			fake:        fakeFunnels{err: errors.New("expa: 401 token=secret123")},
			wantErr:     "could not load funnel data",
			mustNotLeak: "secret123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := (&Resolver{Funnels: tt.fake}).Query()
			got, err := q.Funnel(context.Background(), model.ProgrammeOgx, "2026-01-01", "2026-01-31")

			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				if got.StartDate != "2026-01-01" || got.EndDate != "2026-01-31" {
					t.Errorf("dates = %s..%s", got.StartDate, got.EndDate)
				}
				if len(got.Phases) != len(domain.Phases) || got.Phases[0].Phase != model.PhaseSu {
					t.Errorf("phases not mapped in order: %+v", got.Phases)
				}
				if got.DataSource != "mock" {
					t.Errorf("DataSource = %q", got.DataSource)
				}
				return
			}

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.wantErr)
			}
			if tt.mustNotLeak != "" && strings.Contains(err.Error(), tt.mustNotLeak) {
				t.Fatalf("error leaks %q: %v", tt.mustNotLeak, err)
			}
		})
	}
}
