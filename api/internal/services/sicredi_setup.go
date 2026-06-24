package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Service) SetupSicrediProduction(ctx context.Context, input *models.RegisterSicrediWebhookInput) (*models.SicrediProductionSetupResponse, error) {
	steps := make([]models.SicrediSetupStep, 0, 5)
	allOK := true

	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return &models.SicrediProductionSetupResponse{
			Success: false,
			Message: "Integração Sicredi desabilitada ou sem credenciais.",
			Steps: []models.SicrediSetupStep{{
				Name:    "config",
				OK:      false,
				Message: "Defina SICREDI_ENABLED=true e credenciais OAuth.",
			}},
		}, nil
	}

	cfg := s.Sicredi.Config()
	env := "produção"
	if cfg.Sandbox {
		env = "sandbox"
	}
	steps = append(steps, models.SicrediSetupStep{
		Name:    "environment",
		OK:      true,
		Message: fmt.Sprintf("Ambiente: %s", env),
	})

	if cfg.Production && strings.TrimSpace(cfg.WebhookToken) == "" {
		allOK = false
		steps = append(steps, models.SicrediSetupStep{
			Name:    "webhook_token",
			OK:      false,
			Message: "Configure SICREDI_WEBHOOK_TOKEN (token enviado pelo Sicredi no header Authorization).",
		})
	} else {
		steps = append(steps, models.SicrediSetupStep{
			Name:    "webhook_token",
			OK:      true,
			Message: "Token de webhook configurado.",
		})
	}

	publicURL := strings.TrimSpace(cfg.PublicAPIURL)
	if input != nil && strings.TrimSpace(input.PublicAPIURL) != "" {
		publicURL = strings.TrimSpace(input.PublicAPIURL)
	}
	if publicURL == "" || strings.Contains(publicURL, "localhost") || strings.Contains(publicURL, "127.0.0.1") {
		allOK = false
		steps = append(steps, models.SicrediSetupStep{
			Name:    "public_url",
			OK:      false,
			Message: "Configure SICREDI_PUBLIC_API_URL com a URL HTTPS pública da API.",
		})
	} else {
		steps = append(steps, models.SicrediSetupStep{
			Name:    "public_url",
			OK:      true,
			Message: publicURL,
		})
	}

	if err := s.Sicredi.Ping(ctx); err != nil {
		allOK = false
		steps = append(steps, models.SicrediSetupStep{
			Name:    "connection",
			OK:      false,
			Message: err.Error(),
		})
	} else {
		steps = append(steps, models.SicrediSetupStep{
			Name:    "connection",
			OK:      true,
			Message: "OAuth autenticado com sucesso.",
		})

		webhookOK := false
		if publicURL != "" && !strings.Contains(publicURL, "localhost") && !strings.Contains(publicURL, "127.0.0.1") {
			if contracts, err := s.Sicredi.ListWebhookContracts(ctx); err == nil {
				expected := strings.TrimRight(publicURL, "/") + "/v1/webhooks/sicredi"
				for _, c := range contracts {
					if strings.TrimRight(c.URL, "/") == expected {
						webhookOK = true
						break
					}
				}
			}
			if !webhookOK {
				if _, err := s.RegisterSicrediWebhook(ctx, &models.RegisterSicrediWebhookInput{PublicAPIURL: publicURL}); err != nil {
					steps = append(steps, models.SicrediSetupStep{
						Name:    "webhook_register",
						OK:      false,
						Message: err.Error(),
					})
					allOK = false
				} else {
					webhookOK = true
					webhookURL := strings.TrimRight(publicURL, "/") + "/v1/webhooks/sicredi"
					steps = append(steps, models.SicrediSetupStep{
						Name:    "webhook_register",
						OK:      true,
						Message: "Webhook registrado: " + webhookURL,
					})
				}
			} else {
				steps = append(steps, models.SicrediSetupStep{
					Name:    "webhook_register",
					OK:      true,
					Message: "Webhook já registrado no Sicredi.",
				})
			}
		}

		if contracts, err := s.Sicredi.ListWebhookContracts(ctx); err == nil && len(contracts) > 0 {
			steps = append(steps, models.SicrediSetupStep{
				Name:    "webhook_active",
				OK:      true,
				Message: contracts[0].URL,
			})
		} else if !webhookOK {
			allOK = false
			steps = append(steps, models.SicrediSetupStep{
				Name:    "webhook_active",
				OK:      false,
				Message: "Nenhum contrato de webhook ativo no Sicredi.",
			})
		}
	}

	msg := "Integração Sicredi pronta para produção."
	if !allOK {
		msg = "Corrija os itens pendentes antes de usar em produção."
	}
	return &models.SicrediProductionSetupResponse{
		Success: allOK,
		Message: msg,
		Steps:   steps,
	}, nil
}