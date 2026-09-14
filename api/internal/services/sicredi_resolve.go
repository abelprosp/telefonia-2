package services

import (
	"context"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/sicredi"
)

// sicrediForOrg resolves the Sicredi client for a tenant.
// Prefer organization-specific credentials; fall back to the global env client
// only when the org has not configured its own Sicredi account.
func (s *Service) sicrediForOrg(ctx context.Context, orgID string) SicrediBoletoIssuer {
	if orgID != "" {
		production := false
		if s.Sicredi != nil {
			production = s.Sicredi.Config().Production
		}
		cfg, ok, err := s.Store.SicrediConfigForOrg(ctx, orgID, production)
		if err == nil && ok {
			return sicredi.NewClient(cfg)
		}
		// Org explicitly enabled but incomplete → do not fall back silently.
		if err == nil && cfg.Enabled {
			return sicredi.NewClient(cfg)
		}
	}
	return s.Sicredi
}

func (s *Service) sicrediForRequest(ctx context.Context) SicrediBoletoIssuer {
	orgID, err := orgFrom(ctx)
	if err != nil || strings.TrimSpace(orgID) == "" {
		return s.Sicredi
	}
	return s.sicrediForOrg(ctx, orgID)
}
