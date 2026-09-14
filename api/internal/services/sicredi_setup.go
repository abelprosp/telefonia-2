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

	client := s.sicrediForRequest(ctx)
	if client == nil || !client.Enabled() {
		return &models.SicrediProductionSetupResponse{
			Success: false,
			Message: "Integração Sicredi desabilitada ou sem credenciais nesta empresa.",
			Steps: []models.SicrediSetupStep{{
				Name:    "config",
				OK:      false,
				Message: "Ative o Sicredi e informe as credenciais em Configurações.",
			}},
		}, nil
	}

	cfg := client.Config()
	env := "produção"
	if cfg.Sandbox {
		env = "sandbox"
	}
	steps = append(steps, models.SicrediSetupStep{
		Name:    "environment",
		OK:      true,
		Message: fmt.Sprintf("Ambiente: %s", env),
	})

	if !cfg.Sandbox && strings.TrimSpace(cfg.WebhookToken) == "" {
		allOK = false
		steps = append(steps, models.SicrediSetupStep{
			Name:    "webhook_token",
			OK:      false,
			Message: "Configure o token de webhook nas configurações Sicredi da empresa.",
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
			Message: "Configure a URL HTTPS pública da API nas configurações Sicredi.",
		})
	} else {
		steps = append(steps, models.SicrediSetupStep{
			Name:    "public_url",
			OK:      true,
			Message: publicURL,
		})
	}

	if err := client.Ping(ctx); err != nil {
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

		if publicURL != "" && !strings.Contains(publicURL, "localhost") && !strings.Contains(publicURL, "127.0.0.1") {
			webhookOK := false
			if contracts, err := client.ListWebhookContracts(ctx); err == nil {
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
					allOK = false
					steps = append(steps, models.SicrediSetupStep{
						Name:    "webhook_register",
						OK:      false,
						Message: err.Error(),
					})
				} else {
					steps = append(steps, models.SicrediSetupStep{
						Name:    "webhook_register",
						OK:      true,
						Message: "Webhook registrado.",
					})
				}
			} else {
				steps = append(steps, models.SicrediSetupStep{
					Name:    "webhook_register",
					OK:      true,
					Message: "Webhook já registrado.",
				})
			}
		}
	}

	msg := "Setup Sicredi concluído."
	if !allOK {
		msg = "Setup Sicredi incompleto. Revise os passos com falha."
	}
	return &models.SicrediProductionSetupResponse{
		Success: allOK,
		Message: msg,
		Steps:   steps,
	}, nil
}
