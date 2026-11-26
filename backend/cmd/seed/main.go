// Command seed generates realistic DUMMY funnel data for AIESEC in UB and
// writes it to internal/repository/mock/seed.json.
//
// It uses a fixed random seed, so running it again gives the same file.
//
//	cd backend && go run ./cmd/seed
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"
)

const outFile = "internal/repository/mock/seed.json"

// Data window: Jan 2026 until the task deadline.
var (
	windowStart = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	windowEnd   = time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC)
)

// phases in funnel order, and the chance of moving from the previous phase.
// SU 100% -> APL 40% -> ACC 25% -> APD 15% -> RE 12% -> FIN ~10% -> CO ~9%
var steps = []struct {
	phase   string
	chance  float64 // probability to reach this phase from the previous one
	minDays int     // days after the previous phase
	maxDays int
}{
	{"APL", 0.40, 1, 20},
	{"ACC", 0.625, 3, 21},
	{"APD", 0.60, 2, 14},
	{"RE", 0.80, 20, 60},
	{"FIN", 0.85, 30, 56},
	{"CO", 0.90, 3, 14},
}

type programmeConfig struct {
	programme string
	signUps   int
	products  []string
	weights   []float64 // share of each product
}

var programmes = []programmeConfig{
	{"OGX", 1200, []string{"oGV", "oGTa", "oGTe"}, []float64{0.6, 0.25, 0.15}},
	{"ICX", 500, []string{"iGV", "iGTa", "iGTe"}, []float64{0.6, 0.25, 0.15}},
}

// record is the JSON shape stored in seed.json.
type record struct {
	ID         string            `json:"id"`
	Programme  string            `json:"programme"`
	Product    string            `json:"product"`
	PhaseDates map[string]string `json:"phaseDates"` // phase -> "YYYY-MM-DD"
}

func main() {
	rng := rand.New(rand.NewPCG(42, 2026)) // fixed seed = reproducible
	days := int(windowEnd.Sub(windowStart).Hours()/24) + 1

	var records []record
	for _, pc := range programmes {
		for i := 0; i < pc.signUps; i++ {
			date := windowStart.AddDate(0, 0, rng.IntN(days))
			rec := record{
				ID:         fmt.Sprintf("%s-%04d", pc.programme, i+1),
				Programme:  pc.programme,
				Product:    pick(rng, pc.products, pc.weights),
				PhaseDates: map[string]string{"SU": date.Format("2006-01-02")},
			}

			for _, s := range steps {
				if rng.Float64() >= s.chance {
					break // person stopped here
				}
				date = date.AddDate(0, 0, s.minDays+rng.IntN(s.maxDays-s.minDays+1))
				if date.After(windowEnd) {
					break // has not happened yet
				}
				rec.PhaseDates[s.phase] = date.Format("2006-01-02")
			}
			records = append(records, rec)
		}
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outFile, append(data, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %d records to %s", len(records), outFile)
}

// pick returns one item using the given weights (weights sum to 1).
func pick(rng *rand.Rand, items []string, weights []float64) string {
	x := rng.Float64()
	for i, w := range weights {
		if x < w {
			return items[i]
		}
		x -= w
	}
	return items[len(items)-1]
}
