// Package export turns a computed funnel into downloadable files.
package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/domain"
)

// LCName is printed at the top of every report.
const LCName = "AIESEC in Universitas Brawijaya"

var phaseNames = map[domain.Phase]string{
	domain.SU:  "Sign-up",
	domain.APL: "Applied",
	domain.ACC: "Accepted",
	domain.APD: "Approved",
	domain.RE:  "Realized",
	domain.FIN: "Finished",
	domain.CO:  "Completed",
}

// percent formats a ratio as "40.0", or "" when nil (no data to divide by).
// Empty cells are friendlier than "–" for spreadsheet formulas.
func percent(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v*100, 'f', 1, 64)
}

// WriteCSV writes the report in three small blocks:
// info (what this file is), phases (the funnel), key conversions.
func WriteCSV(w io.Writer, f domain.Funnel, generatedAt time.Time) error {
	cw := csv.NewWriter(w)

	rows := [][]string{
		{"LC", LCName},
		{"Programme", programmeLabel(f.Programme)},
		{"Start date", f.Range.Start.Format(domain.DateLayout)},
		{"End date", f.Range.End.Format(domain.DateLayout)},
		{"Data source", f.DataSource},
		{"Generated at", generatedAt.Format(time.RFC3339)},
		{},
		{"Phase", "Phase name", "Count", "Conversion from previous (%)"},
	}
	for _, p := range f.Phases {
		rows = append(rows, []string{
			string(p.Phase), phaseNames[p.Phase], strconv.Itoa(p.Count), percent(p.ConversionFromPrevious),
		})
	}
	rows = append(rows,
		[]string{},
		[]string{"Key conversion", "Rate (%)"},
		[]string{"SU -> APL", percent(f.SuToApl)},
		[]string{"APL -> APD", percent(f.AplToApd)},
		[]string{"APD -> RE", percent(f.ApdToRe)},
		[]string{"SU -> CO", percent(f.SuToCo)},
	)

	// Rows have different lengths on purpose (small report, not a table dump).
	cw.UseCRLF = true // best for Excel
	for _, r := range rows {
		if err := cw.Write(r); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// Filename builds e.g. "lc-funnel_oGX_2026-01-01_2026-11-26.csv".
func Filename(f domain.Funnel) string {
	return fmt.Sprintf("lc-funnel_%s_%s_%s.csv",
		programmeLabel(f.Programme),
		f.Range.Start.Format(domain.DateLayout),
		f.Range.End.Format(domain.DateLayout))
}

func programmeLabel(p domain.Programme) string {
	switch p {
	case domain.OGX:
		return "oGX"
	case domain.ICX:
		return "iCX"
	}
	return string(p)
}
