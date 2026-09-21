package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luxus-connect/telefonia/api/internal/config"
)

func TestValidateTokenIssuerAudience_ok(t *testing.T) {
	m := &Middleware{cfg: config.Config{
		KeycloakPublicAuthServerURL: "https://auth.example.com/auth",
		KeycloakAuthServerURL:       "http://keycloak:8080/auth",
		KeycloakRealm:               "luxus",
		KeycloakResource:            "connect-cli",
	}}
	claims := jwt.MapClaims{
		"iss": "https://auth.example.com/auth/realms/luxus",
		"azp": "connect-cli",
		"aud": "account",
	}
	if err := m.validateTokenIssuerAudience(claims); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTokenIssuerAudience_rejectsWrongAudience(t *testing.T) {
	m := &Middleware{cfg: config.Config{
		KeycloakPublicAuthServerURL: "https://auth.example.com/auth",
		KeycloakRealm:               "luxus",
		KeycloakResource:            "connect-cli",
	}}
	claims := jwt.MapClaims{
		"iss": "https://auth.example.com/auth/realms/luxus",
		"azp": "other-client",
	}
	if err := m.validateTokenIssuerAudience(claims); err == nil {
		t.Fatal("expected audience rejection")
	}
}

func TestValidateTokenIssuerAudience_rejectsWrongIssuer(t *testing.T) {
	m := &Middleware{cfg: config.Config{
		KeycloakPublicAuthServerURL: "https://auth.example.com/auth",
		KeycloakAuthServerURL:       "http://keycloak:8080/auth",
		KeycloakRealm:               "luxus",
		KeycloakResource:            "connect-cli",
	}}
	claims := jwt.MapClaims{
		"iss": "https://evil.example/realms/luxus",
		"azp": "connect-cli",
	}
	if err := m.validateTokenIssuerAudience(claims); err == nil {
		t.Fatal("expected issuer rejection")
	}
}

func TestValidateTokenIssuerAudience_acceptsInternalIssuer(t *testing.T) {
	m := &Middleware{cfg: config.Config{
		KeycloakPublicAuthServerURL: "https://auth.example.com/auth",
		KeycloakAuthServerURL:       "http://keycloak:8080/auth",
		KeycloakRealm:               "luxus",
		KeycloakResource:            "connect-cli",
	}}
	claims := jwt.MapClaims{
		"iss": "http://keycloak:8080/auth/realms/luxus",
		"azp": "connect-cli",
	}
	if err := m.validateTokenIssuerAudience(claims); err != nil {
		t.Fatal(err)
	}
}
