// Command server runs the whole dashboard: GraphQL API, CSV export and the
// static frontend, on one port. The same binary runs locally and on Vercel.
//
//	cd backend && go run ./cmd/server
//	open http://localhost:8080
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/app"
	"github.com/haramurti/neval-IID-staff-technicaltask/backend/web"
)

func main() {
	a, err := app.NewFromEnv()
	if err != nil {
		log.Fatalf("startup: %v", err)
	}

	port := getenv("PORT", "8080") // Vercel sets PORT for us

	mux := http.NewServeMux()
	mux.Handle("POST /graphql", a.GraphQLHandler())
	mux.Handle("GET /export.csv", a.ExportCSVHandler())
	mux.Handle("GET /playground", app.PlaygroundHandler("/graphql"))
	mux.Handle("GET /", frontendHandler())

	log.Printf("data source: %s", getenv("DATA_SOURCE", "mock")) // never log the token
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// frontendHandler serves the page.
//   - Local dev: straight from ../frontend on disk, so edits show on refresh.
//   - Deployed: from the copy embedded in the binary (web/dist).
func frontendHandler() http.Handler {
	dir := getenv("FRONTEND_DIR", "../frontend")
	if _, err := os.Stat(dir + "/index.html"); err == nil {
		log.Printf("frontend: serving from disk (%s)", dir)
		return http.FileServer(http.Dir(dir))
	}
	if web.HasIndex() {
		log.Printf("frontend: serving embedded copy")
		return http.FileServerFS(web.Files())
	}
	log.Printf("frontend: NOT FOUND (no %s and nothing embedded)", dir)
	return http.NotFoundHandler()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
