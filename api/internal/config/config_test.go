package config

import (
	"net/url"
	"testing"
)

func TestLocalAuthRequiresOnlySecretAndFrontendOrigin(t *testing.T) {
	c := Config{DatabaseURL: "postgres://local-test", AuthSecret: "01234567890123456789012345678901", CORSOrigins: []string{"http://localhost:5173"}, WebAuthnRPID: "localhost"}
	if err := c.Validate(); err != nil {
		t.Fatalf("local auth unexpectedly requires an external identity provider: %v", err)
	}
	invalid := c
	invalid.AuthSecret = ""
	if invalid.Validate() == nil {
		t.Fatal("missing encryption key accepted")
	}
	invalid = c
	invalid.CORSOrigins = nil
	if invalid.Validate() == nil {
		t.Fatal("missing CSRF origins accepted")
	}
	for _, rpID := range []string{"", "https://localhost", "localhost:5173", "localhost/path", "evil.example"} {
		invalid = c
		invalid.WebAuthnRPID = rpID
		if invalid.Validate() == nil {
			t.Errorf("invalid WebAuthn RP ID accepted: %q", rpID)
		}
	}
	invalid = c
	invalid.CORSOrigins = []string{"https://telefonia.example"}
	invalid.WebAuthnRPID = "other.example"
	if invalid.Validate() == nil {
		t.Fatal("RP ID unrelated to CORS origin accepted")
	}
}

func TestLoadDerivesWebAuthnRPIDFromFirstOrigin(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "https://app.telefonia.example")
	t.Setenv("WEBAUTHN_RP_ID", "")
	t.Setenv("WEBAUTHN_RP_NAME", "")
	c := Load()
	if c.WebAuthnRPID != "app.telefonia.example" || c.WebAuthnRPName != "Telefonia" {
		t.Fatalf("unexpected WebAuthn defaults: %q %q", c.WebAuthnRPID, c.WebAuthnRPName)
	}
}

func TestDatabaseCredentialsRoundTrip(t *testing.T) {
	password := "p@ss:/?#% word"
	raw := NormalizeDatabaseURL("Host=localhost;Username=user@name;Password=" + password + ";Database=connect")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := u.User.Password()
	if got != password || u.User.Username() != "user@name" || u.Host != "localhost:5432" {
		t.Fatal("credentials altered by URL construction")
	}
}

func TestProductionRejectsInvalidStorageInsteadOfChangingDomain(t *testing.T) {
	c := Config{AuthSecret: "01234567890123456789012345678901", CORSOrigins: []string{"https://app.example"}, WebAuthnRPID: "app.example", Environment: "Production", DatabaseURL: "postgres://db", ObjectStoragePublicURL: "https://storage.example", ObjectStorageAccessKeyID: "application", ObjectStorageSecretKey: "dedicated-secret"}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, rpID := range []string{"localhost", "127.0.0.1", "::1"} {
		invalid := c
		invalid.WebAuthnRPID = rpID
		if invalid.Validate() == nil {
			t.Errorf("production accepted local WebAuthn RP ID: %s", rpID)
		}
	}
	for _, raw := range []string{"", "http://storage.example", "https://127.0.0.1", "https://storage.example/storage"} {
		invalid := c
		invalid.ObjectStoragePublicURL = raw
		if invalid.Validate() == nil {
			t.Errorf("accepted invalid URL: %s", raw)
		}
		if invalid.ObjectStoragePublicURL != raw {
			t.Fatal("configuration silently rewritten")
		}
	}
	c.ObjectStorageSecretKey = "minioadmin"
	if c.Validate() == nil {
		t.Fatal("default production credentials accepted")
	}
}
