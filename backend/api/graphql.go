// Vercel serverless function: every .go file in api/ becomes one endpoint.
// This one is /api/graphql (vercel.json also maps /graphql here).
//
// It runs the exact same code as the local server (internal/app).
package handler

import (
	"net/http"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/app"
)

// GraphQL is the function Vercel calls for each request.
func GraphQL(w http.ResponseWriter, r *http.Request) {
	a, err := app.Shared()
	if err != nil {
		app.StartupFailed(w)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return
	}
	a.GraphQLHandler().ServeHTTP(w, r)
}
