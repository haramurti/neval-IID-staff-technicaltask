package export

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/domain"
)

func TestWriteCSV(t *testing.T) {
	r, _ := domain.NewDateRange("2026-01-01", "2026-01-31")
	records := []domain.FunnelRecord{
		{Programme: domain.OGX, PhaseDates: map[domain.Phase]time.Time{
			domain.SU: r.Start, domain.APL: r.Start,
		}},
		{Programme: domain.OGX, PhaseDates: map[domain.Phase]time.Time{domain.SU: r.Start}},
	}
	f := domain.BuildFunnel(domain.OGX, r, records)
	f.DataSource = "mock"

	var buf bytes.Buffer
	if err := WriteCSV(&buf, f, time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	// The output must be valid CSV.
	cr := csv.NewReader(strings.NewReader(buf.String()))
	cr.FieldsPerRecord = -1 // rows have different lengths
	rows, err := cr.ReadAll()
	if err != nil {
		t.Fatalf("not valid CSV: %v", err)
	}

	tests := []struct {
		name string
		want []string
	}{
		{"lc name", []string{"LC", LCName}},
		{"programme label", []string{"Programme", "oGX"}},
		{"data source", []string{"Data source", "mock"}},
		{"SU row, first phase has empty conversion", []string{"SU", "Sign-up", "2", ""}},
		{"APL row with percent", []string{"APL", "Applied", "1", "50.0"}},
		{"ACC row, 0 of 1 is 0 percent", []string{"ACC", "Accepted", "0", "0.0"}},
		{"APD row, 0 of 0 is empty, not NaN", []string{"APD", "Approved", "0", ""}},
		{"key conversion", []string{"SU -> APL", "50.0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !containsRow(rows, tt.want) {
				t.Errorf("row %q not found in:\n%s", tt.want, buf.String())
			}
		})
	}

	if strings.Contains(buf.String(), "NaN") || strings.Contains(buf.String(), "Inf") {
		t.Error("CSV must never contain NaN or Inf")
	}
}

func TestFilename(t *testing.T) {
	r, _ := domain.NewDateRange("2026-01-01", "2026-11-26")
	got := Filename(domain.Funnel{Programme: domain.ICX, Range: r})
	want := "lc-funnel_iCX_2026-01-01_2026-11-26.csv"
	if got != want {
		t.Errorf("Filename = %q, want %q", got, want)
	}
}

func containsRow(rows [][]string, want []string) bool {
	for _, r := range rows {
		if len(r) != len(want) {
			continue
		}
		same := true
		for i := range r {
			if r[i] != want[i] {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	return false
}
