// Package expa will load funnel data from the real EXPA GraphQL API.
//
// STATUS: STUB. We do not have an EXPA access token yet, so the real
// schema could not be introspected. Field names must NOT be guessed.
package expa

import (
	"context"
	"errors"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/domain"
	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/service"
)

// Endpoint is the EXPA GraphQL API.
const Endpoint = "https://gis-api.aiesec.org/graphql"

// ErrNotImplemented is returned until the real query is written.
var ErrNotImplemented = errors.New("expa repository not implemented yet")

// Repository talks to EXPA. The token is only ever held here, on the server.
type Repository struct {
	token string
}

// New creates the repository. token comes from the EXPA_ACCESS_TOKEN env var.
func New(token string) *Repository {
	return &Repository{token: token}
}

// ListFunnelRecords must return the same domain records as the mock repo.
//
// TODO(expa): once a token is available:
//  1. Introspect the schema at Endpoint (send the token in the
//     Authorization header) and find the query for applications/people
//     of an LC (AIESEC in Universitas Brawijaya), filtered by programme
//     (oGX: outgoing, iCX: incoming) and date range.
//  2. Request, per application: id, product (GV/GTa/GTe) and the date it
//     reached each phase (SU, APL, ACC, APD, RE, FIN, CO).
//  3. POST the query with net/http, map every item to domain.FunnelRecord.
//  4. Never log the token or put it in returned error messages.
func (r *Repository) ListFunnelRecords(_ context.Context, _ service.Filter) ([]domain.FunnelRecord, error) {
	return nil, ErrNotImplemented
}

// Source is shown in the UI badge.
func (r *Repository) Source() string { return "expa" }
