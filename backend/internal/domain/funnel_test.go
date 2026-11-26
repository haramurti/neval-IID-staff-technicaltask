package domain

import (
	"errors"
	"testing"
	"time"
)

func day(s string) time.Time {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

func mustRange(t *testing.T, start, end string) DateRange {
	t.Helper()
	r, err := NewDateRange(start, end)
	if err != nil {
		t.Fatalf("NewDateRange(%q, %q): %v", start, end, err)
	}
	return r
}

// ptr makes a *float64 for expected values.
func ptr(v float64) *float64 { return &v }

// sameRate compares two optional rates (nil == nil, else values close enough).
func sameRate(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	d := *a - *b
	return d < 1e-9 && d > -1e-9
}

func TestConversion(t *testing.T) {
	tests := []struct {
		name     string
		num, den int
		want     *float64
	}{
		{"normal", 40, 100, ptr(0.4)},
		{"zero numerator", 0, 10, ptr(0)},
		{"zero denominator gives nil, not NaN", 0, 0, nil},
		{"nonzero over zero gives nil, not Inf", 5, 0, nil},
		{"above 100 percent is allowed", 3, 2, ptr(1.5)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Conversion(tt.num, tt.den); !sameRate(got, tt.want) {
				t.Errorf("Conversion(%d, %d) = %v, want %v", tt.num, tt.den, got, tt.want)
			}
		})
	}
}

func TestNewDateRange(t *testing.T) {
	tests := []struct {
		name       string
		start, end string
		wantErr    bool
	}{
		{"valid range", "2026-01-01", "2026-03-31", false},
		{"same day", "2026-05-05", "2026-05-05", false},
		{"start after end", "2026-04-01", "2026-03-01", true},
		{"bad start format", "01-01-2026", "2026-03-01", true},
		{"bad end format", "2026-01-01", "2026/03/01", true},
		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDateRange(tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrInvalidDateRange) {
				t.Errorf("error should wrap ErrInvalidDateRange, got %v", err)
			}
		})
	}
}

func TestDateRangeContainsIsInclusive(t *testing.T) {
	r := mustRange(t, "2026-02-01", "2026-02-28")
	cases := map[string]bool{
		"2026-01-31": false,
		"2026-02-01": true, // first day included
		"2026-02-15": true,
		"2026-02-28": true, // last day included
		"2026-03-01": false,
	}
	for d, want := range cases {
		if got := r.Contains(day(d)); got != want {
			t.Errorf("Contains(%s) = %v, want %v", d, got, want)
		}
	}
}

func TestBuildFunnel(t *testing.T) {
	// Small hand-made data set.
	records := []FunnelRecord{
		{ID: "1", Programme: OGX, PhaseDates: map[Phase]time.Time{
			SU: day("2026-01-05"), APL: day("2026-01-10"), ACC: day("2026-01-20"),
			APD: day("2026-02-01"), RE: day("2026-03-01"), FIN: day("2026-04-15"), CO: day("2026-04-20"),
		}},
		{ID: "2", Programme: OGX, PhaseDates: map[Phase]time.Time{
			SU: day("2026-01-06"), APL: day("2026-01-15"),
		}},
		{ID: "3", Programme: OGX, PhaseDates: map[Phase]time.Time{
			SU: day("2026-02-10"),
		}},
		{ID: "4", Programme: ICX, PhaseDates: map[Phase]time.Time{ // other programme: ignored for OGX
			SU: day("2026-01-05"), APL: day("2026-01-06"),
		}},
	}

	tests := []struct {
		name       string
		programme  Programme
		start, end string
		wantCounts []int // in Phases order: SU APL ACC APD RE FIN CO
		wantSuApl  *float64
		wantSuCo   *float64
	}{
		{
			name: "whole period", programme: OGX, start: "2026-01-01", end: "2026-12-31",
			wantCounts: []int{3, 2, 1, 1, 1, 1, 1},
			wantSuApl:  ptr(2.0 / 3.0), wantSuCo: ptr(1.0 / 3.0),
		},
		{
			name: "january only", programme: OGX, start: "2026-01-01", end: "2026-01-31",
			wantCounts: []int{2, 2, 1, 0, 0, 0, 0},
			wantSuApl:  ptr(1), wantSuCo: ptr(0),
		},
		{
			name: "only later phases in range, SU is 0 so rates are nil", programme: OGX,
			start: "2026-03-01", end: "2026-04-30",
			wantCounts: []int{0, 0, 0, 0, 1, 1, 1},
			wantSuApl:  nil, wantSuCo: nil,
		},
		{
			name: "no data in period", programme: OGX, start: "2025-01-01", end: "2025-12-31",
			wantCounts: []int{0, 0, 0, 0, 0, 0, 0},
			wantSuApl:  nil, wantSuCo: nil,
		},
		{
			name: "other programme", programme: ICX, start: "2026-01-01", end: "2026-12-31",
			wantCounts: []int{1, 1, 0, 0, 0, 0, 0},
			wantSuApl:  ptr(1), wantSuCo: ptr(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := BuildFunnel(tt.programme, mustRange(t, tt.start, tt.end), records)

			if len(f.Phases) != len(Phases) {
				t.Fatalf("got %d phases, want %d", len(f.Phases), len(Phases))
			}
			for i, pc := range f.Phases {
				if pc.Phase != Phases[i] {
					t.Errorf("phase %d = %s, want %s (order matters)", i, pc.Phase, Phases[i])
				}
				if pc.Count != tt.wantCounts[i] {
					t.Errorf("%s count = %d, want %d", pc.Phase, pc.Count, tt.wantCounts[i])
				}
			}
			if f.Phases[0].ConversionFromPrevious != nil {
				t.Error("first phase must have nil conversion")
			}
			if !sameRate(f.SuToApl, tt.wantSuApl) {
				t.Errorf("SuToApl = %v, want %v", f.SuToApl, tt.wantSuApl)
			}
			if !sameRate(f.SuToCo, tt.wantSuCo) {
				t.Errorf("SuToCo = %v, want %v", f.SuToCo, tt.wantSuCo)
			}
		})
	}
}

func TestBuildFunnelStepConversion(t *testing.T) {
	// 4 SU, 2 APL, 0 ACC, 0 APD -> APL/SU = 0.5, ACC/APL = 0, APD/ACC = nil
	var records []FunnelRecord
	for i := 0; i < 4; i++ {
		dates := map[Phase]time.Time{SU: day("2026-01-01")}
		if i < 2 {
			dates[APL] = day("2026-01-02")
		}
		records = append(records, FunnelRecord{Programme: OGX, PhaseDates: dates})
	}

	f := BuildFunnel(OGX, mustRange(t, "2026-01-01", "2026-01-31"), records)

	want := map[Phase]*float64{APL: ptr(0.5), ACC: ptr(0), APD: nil}
	for _, pc := range f.Phases {
		w, ok := want[pc.Phase]
		if !ok {
			continue
		}
		if !sameRate(pc.ConversionFromPrevious, w) {
			t.Errorf("%s conversion = %v, want %v", pc.Phase, pc.ConversionFromPrevious, w)
		}
	}
}
