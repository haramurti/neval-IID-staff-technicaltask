// Package app wires everything together: picks the data source from env,
// builds the service, and exposes HTTP handlers.
//
// It is shared by the local server (cmd/server) and, later, the Vercel
// functions, so both run exactly the same code.
package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/graph"
	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/repository/expa"
	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/repository/mock"
	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/service"
)

// App holds the ready-to-use service.
type App struct {
	Funnels *service.FunnelService
}

// NewFromEnv reads config from environment variables:
//
//	DATA_SOURCE        "mock" (default) or "expa"
//	EXPA_ACCESS_TOKEN  required when DATA_SOURCE=expa (never logged)
func NewFromEnv() (*App, error) {
	repo, err := newRepository(os.Getenv("DATA_SOURCE"), os.Getenv("EXPA_ACCESS_TOKEN"))
	if err != nil {
		return nil, err
	}
	return &App{Funnels: service.NewFunnelService(repo)}, nil
}

func newRepository(source, token string) (service.PerformanceRepository, error) {
	switch source {
	case "", "mock":
		return mock.New()
	case "expa":
		if token == "" {
			return nil, errors.New("DATA_SOURCE=expa needs EXPA_ACCESS_TOKEN to be set")
		}
		return expa.New(token), nil
	default:
		return nil, fmt.Errorf("unknown DATA_SOURCE %q (use mock or expa)", source)
	}
}

// GraphQLHandler serves the /graphql endpoint (POST with a JSON body).
func (a *App) GraphQLHandler() http.Handler {
	schema := graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{Funnels: a.Funnels},
	})
	srv := handler.New(schema)
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{}) // lets the playground show the schema
	return srv
}

// PlaygroundHandler is a small in-browser GraphQL IDE, handy for demos.
func PlaygroundHandler(graphqlPath string) http.Handler {
	return playground.Handler("LC Dashboard GraphQL", graphqlPath)
}
