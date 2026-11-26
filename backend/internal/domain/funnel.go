// Package domain holds the core business concepts of the dashboard:
// programmes, funnel phases, and how counts and conversions are computed.
//
// It is plain Go: no GraphQL, no HTTP, no EXPA. Everything else depends
// on this package, never the other way around.
package domain

import (
	"errors"
	"fmt"
	"time"
)

// Programme is the exchange direction we report on.
type Programme string

const (
	OGX Programme = "OGX" // outgoing: our members going abroad
	ICX Programme = "ICX" // incoming: people coming to our LC
)

// Valid reports whether p is a known programme.
func (p Programme) Valid() bool { return p == OGX || p == ICX }

// Phase is one step of the EXPA funnel.
type Phase string

const (
	SU  Phase = "SU"  // Sign-up
	APL Phase = "APL" // Applied
	ACC Phase = "ACC" // Accepted
	APD Phase = "APD" // Approved
	RE  Phase = "RE"  // Realized
	FIN Phase = "FIN" // Finished
	CO  Phase = "CO"  // Completed
)

// Phases lists every phase in funnel order. To add a phase, add it here
// (and to the GraphQL schema enum).
var Phases = []Phase{SU, APL, ACC, APD, RE, FIN, CO}

// DateLayout is the only date format we accept: YYYY-MM-DD.
const DateLayout = "2006-01-02"

// ErrInvalidDateRange is returned for bad or reversed dates.
var ErrInvalidDateRange = errors.New("invalid date range")

// DateRange is an inclusive range of whole days.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// NewDateRange parses two YYYY-MM-DD strings and checks start <= end.
func NewDateRange(start, end string) (DateRange, error) {
	s, err := time.Parse(DateLayout, start)
	if err != nil {
		return DateRange{}, fmt.Errorf("%w: startDate %q must be YYYY-MM-DD", ErrInvalidDateRange, start)
	}
	e, err := time.Parse(DateLayout, end)
	if err != nil {
		return DateRange{}, fmt.Errorf("%w: endDate %q must be YYYY-MM-DD", ErrInvalidDateRange, end)
	}
	if s.After(e) {
		return DateRange{}, fmt.Errorf("%w: startDate must be before or equal to endDate", ErrInvalidDateRange)
	}
	return DateRange{Start: s, End: e}, nil
}

// Contains reports whether t falls on a day inside the range (both ends included).
func (r DateRange) Contains(t time.Time) bool {
	return !t.Before(r.Start) && !t.After(r.End)
}

// FunnelRecord is one person/application and the date they reached each phase.
// A phase missing from PhaseDates means the person never reached it.
type FunnelRecord struct {
	ID         string
	Programme  Programme
	Product    string // e.g. "oGV", "oGTa", "iGTe"
	PhaseDates map[Phase]time.Time
}

// PhaseCount is the number of records that reached a phase in the period.
type PhaseCount struct {
	Phase Phase
	Count int
	// ConversionFromPrevious = Count / previous phase Count.
	// nil for the first phase, or when the previous count is 0.
	ConversionFromPrevious *float64
}

// Funnel is the computed result for one programme and one period.
type Funnel struct {
	Programme  Programme
	Range      DateRange
	Phases     []PhaseCount
	SuToApl    *float64
	AplToApd   *float64
	ApdToRe    *float64
	SuToCo     *float64
	DataSource string // "mock" or "expa"
}

// Conversion returns num/den, or nil when den is 0 (so we never return NaN or Inf).
func Conversion(num, den int) *float64 {
	if den == 0 {
		return nil
	}
	v := float64(num) / float64(den)
	return &v
}

// BuildFunnel counts, for each phase, how many records reached that phase
// on a date inside r, then computes the conversion rates.
//
// Counting rule: a record counts for a phase if *that phase's date* is in
// the range. This matches how EXPA reports "how many APDs happened in March".
// Note: because each phase is counted by its own date, a step conversion can
// be above 100% for short periods (e.g. APDs this month from last month's APLs).
func BuildFunnel(p Programme, r DateRange, records []FunnelRecord) Funnel {
	counts := make(map[Phase]int, len(Phases))
	for _, rec := range records {
		if rec.Programme != p {
			continue
		}
		for phase, date := range rec.PhaseDates {
			if r.Contains(date) {
				counts[phase]++
			}
		}
	}

	phases := make([]PhaseCount, len(Phases))
	for i, phase := range Phases {
		pc := PhaseCount{Phase: phase, Count: counts[phase]}
		if i > 0 {
			pc.ConversionFromPrevious = Conversion(counts[phase], counts[Phases[i-1]])
		}
		phases[i] = pc
	}

	return Funnel{
		Programme: p,
		Range:     r,
		Phases:    phases,
		SuToApl:   Conversion(counts[APL], counts[SU]),
		AplToApd:  Conversion(counts[APD], counts[APL]),
		ApdToRe:   Conversion(counts[RE], counts[APD]),
		SuToCo:    Conversion(counts[CO], counts[SU]),
	}
}
