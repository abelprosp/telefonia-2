package services

import (
	"context"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/sicredi"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

// sicrediForOrg resolves the Sicredi client for a tenant.
// Organization credentials always win. Global env credentials are used only
// for the legacy Luxus organization when the tenant has not configured Sicredi.
func (s *Service) sicrediForOrg(ctx context.Context, orgID string) SicrediBoletoIssuer {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return nil
	}

	production := false
	if s.Sicredi != nil {
		production = s.Sicredi.Config().Production
	}
	cfg, ok, err := s.Store.SicrediConfigForOrg(ctx, orgID, production)
	if err == nil && ok {
		return sicredi.NewClient(cfg)
	}
	// Org explicitly enabled but incomplete → do not fall back.
	if err == nil && cfg.Enabled {
		return sicredi.NewClient(cfg)
	}
	// Strict isolation: never reuse Luxus/global Sicredi for other tenants.
	if orgID == store.DefaultLuxusOrgID && s.Sicredi != nil && s.Sicredi.Enabled() {
		return s.Sicredi
	}
	return nil
}

func (s *Service) sicrediForRequest(ctx context.Context) SicrediBoletoIssuer {
	orgID, err := orgFrom(ctx)
	if err != nil || strings.TrimSpace(orgID) == "" {
		return nil
	}
	return s.sicrediForOrg(ctx, orgID)
}
