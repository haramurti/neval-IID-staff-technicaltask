// Command server runs the dashboard locally: the GraphQL API and the
// static frontend on one port.
//
//	cd backend && go run ./cmd/server
//	open http://localhost:8080
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/app"
)

func main() {
	a, err := app.NewFromEnv()
	if err != nil {
		log.Fatalf("startup: %v", err)
	}

	port := getenv("PORT", "8080")
	frontendDir := getenv("FRONTEND_DIR", "../frontend")

	mux := http.NewServeMux()
	mux.Handle("POST /graphql", a.GraphQLHandler())
	mux.Handle("GET /playground", app.PlaygroundHandler("/graphql"))
	mux.Handle("GET /", http.FileServer(http.Dir(frontendDir)))

	log.Printf("data source: %s", sourceName())
	log.Printf("listening on http://localhost:%s (frontend from %s)", port, frontendDir)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// sourceName is for the startup log only. It never prints the token.
func sourceName() string {
	return getenv("DATA_SOURCE", "mock")
}
