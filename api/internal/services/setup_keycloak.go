package services

import (
	"context"
	"fmt"
)

type KeycloakSetupStep struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type KeycloakSetupResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Steps   []KeycloakSetupStep `json:"steps"`
}

func (s *Service) SetupKeycloak(ctx context.Context) (*KeycloakSetupResponse, error) {
	steps := make([]KeycloakSetupStep, 0, 3)
	allOK := true

	if s.Keycloak == nil || !s.Keycloak.Enabled() {
		return &KeycloakSetupResponse{
			Success: false,
			Message: "Integração Keycloak desabilitada ou sem credenciais.",
			Steps: []KeycloakSetupStep{{
				Name:    "config",
				OK:      false,
				Message: "Defina KEYCLOAK_AUTH_SERVER_URL, KEYCLOAK_REALM, KEYCLOAK_ADMIN_USERNAME e KEYCLOAK_ADMIN_PASSWORD.",
			}},
		}, nil
	}

	steps = append(steps, KeycloakSetupStep{
		Name:    "config",
		OK:      true,
		Message: fmt.Sprintf("Keycloak configurado (realm: %s).", s.Keycloak.Realm()),
	})

	users, err := s.Keycloak.ListUsers(ctx, "", 1)
	if err != nil {
		allOK = false
		steps = append(steps, KeycloakSetupStep{
			Name:    "connection",
			OK:      false,
			Message: fmt.Sprintf("Falha ao conectar ao Keycloak Admin API: %s", err.Error()),
		})
	} else {
		steps = append(steps, KeycloakSetupStep{
			Name:    "connection",
			OK:      true,
			Message: fmt.Sprintf("Conexão com Keycloak Admin API OK. Usuários encontrados: %d.", len(users)),
		})
	}

	msg := "Integração Keycloak pronta."
	if !allOK {
		msg = "Corrija os itens pendentes para usar o Keycloak."
	}
	return &KeycloakSetupResponse{
		Success: allOK,
		Message: msg,
		Steps:   steps,
	}, nil
}
