// Package mock is a PerformanceRepository backed by DUMMY data in seed.json.
// seed.json is embedded into the binary, so no file needs to exist at runtime.
// Regenerate it with: go run ./cmd/seed
package mock

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/domain"
	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/service"
)

//go:embed seed.json
var seedJSON []byte

// seedRecord is the JSON shape written by cmd/seed.
type seedRecord struct {
	ID         string            `json:"id"`
	Programme  string            `json:"programme"`
	Product    string            `json:"product"`
	PhaseDates map[string]string `json:"phaseDates"`
}

// Repository serves records from memory.
type Repository struct {
	records []domain.FunnelRecord
}

// New parses the embedded seed data once.
func New() (*Repository, error) {
	var raw []seedRecord
	if err := json.Unmarshal(seedJSON, &raw); err != nil {
		return nil, fmt.Errorf("parse seed.json: %w", err)
	}

	records := make([]domain.FunnelRecord, 0, len(raw))
	for _, r := range raw {
		dates := make(map[domain.Phase]time.Time, len(r.PhaseDates))
		for phase, s := range r.PhaseDates {
			t, err := time.Parse(domain.DateLayout, s)
			if err != nil {
				return nil, fmt.Errorf("seed record %s, phase %s: %w", r.ID, phase, err)
			}
			dates[domain.Phase(phase)] = t
		}
		records = append(records, domain.FunnelRecord{
			ID:         r.ID,
			Programme:  domain.Programme(r.Programme),
			Product:    r.Product,
			PhaseDates: dates,
		})
	}
	return &Repository{records: records}, nil
}

// ListFunnelRecords returns records of the programme with at least one
// phase date inside the range.
func (r *Repository) ListFunnelRecords(_ context.Context, f service.Filter) ([]domain.FunnelRecord, error) {
	var out []domain.FunnelRecord
	for _, rec := range r.records {
		if rec.Programme != f.Programme {
			continue
		}
		for _, d := range rec.PhaseDates {
			if f.Range.Contains(d) {
				out = append(out, rec)
				break
			}
		}
	}
	return out, nil
}

// Source is shown in the UI badge.
func (r *Repository) Source() string { return "mock" }
