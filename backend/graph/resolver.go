package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"context"

	"github.com/haramurti/neval-IID-staff-technicaltask/backend/internal/domain"
)

// FunnelGetter is what the resolvers need from the service layer.
// *service.FunnelService satisfies it; tests can pass a fake.
type FunnelGetter interface {
	GetFunnel(ctx context.Context, p domain.Programme, start, end string) (domain.Funnel, error)
}

// Resolver is the root resolver. It only holds dependencies.
type Resolver struct {
	Funnels FunnelGetter
}
