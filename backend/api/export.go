// Vercel serverless function: /api/export (vercel.json maps /export.csv here).
package handler

import (
	"net/http"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/app"
)

// Export is the function Vercel calls for each request.
func Export(w http.ResponseWriter, r *http.Request) {
	a, err := app.Shared()
	if err != nil {
		app.StartupFailed(w)
		return
	}
	a.ExportCSVHandler().ServeHTTP(w, r)
}
