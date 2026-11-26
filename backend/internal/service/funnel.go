// Package service contains the use cases of the app. It sits between the
// input adapters (GraphQL, CSV export) and the data adapters (mock, EXPA).
package service

import (
	"context"
	"fmt"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/domain"
)

// Filter is what we ask the repository for.
type Filter struct {
	Programme domain.Programme
	Range     domain.DateRange
}

// PerformanceRepository is the "port" for loading funnel data.
// It is defined here, by the code that USES it (Go style: accept interfaces).
// Implementations: repository/mock (seed.json) and repository/expa (EXPA API).
type PerformanceRepository interface {
	// ListFunnelRecords returns records of the programme that reached at
	// least one phase inside the date range.
	ListFunnelRecords(ctx context.Context, f Filter) ([]domain.FunnelRecord, error)
	// Source names the data source shown in the UI: "mock" or "expa".
	Source() string
}

// FunnelService computes funnel numbers.
type FunnelService struct {
	repo PerformanceRepository
}

// NewFunnelService wires the service to a repository.
func NewFunnelService(repo PerformanceRepository) *FunnelService {
	return &FunnelService{repo: repo}
}

// GetFunnel validates input, loads records and builds the funnel.
// Dates are YYYY-MM-DD strings, as they come from the API.
func (s *FunnelService) GetFunnel(ctx context.Context, p domain.Programme, start, end string) (domain.Funnel, error) {
	if !p.Valid() {
		return domain.Funnel{}, fmt.Errorf("unknown programme %q", p)
	}
	r, err := domain.NewDateRange(start, end)
	if err != nil {
		return domain.Funnel{}, err
	}

	records, err := s.repo.ListFunnelRecords(ctx, Filter{Programme: p, Range: r})
	if err != nil {
		return domain.Funnel{}, fmt.Errorf("load funnel records: %w", err)
	}

	f := domain.BuildFunnel(p, r, records)
	f.DataSource = s.repo.Source()
	return f, nil
}
