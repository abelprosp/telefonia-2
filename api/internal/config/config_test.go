package config

import (
	"net/url"
	"testing"
)

func TestLocalAuthRequiresOnlySecretAndFrontendOrigin(t *testing.T) {
	c := Config{DatabaseURL: "postgres://local-test", AuthSecret: "01234567890123456789012345678901", CORSOrigins: []string{"http://localhost:5173"}}
	if err := c.Validate(); err != nil {
		t.Fatalf("local auth incorrectly requires Keycloak: %v", err)
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
	c := Config{AuthSecret: "01234567890123456789012345678901", CORSOrigins: []string{"https://app.example"}, Environment: "Production", DatabaseURL: "postgres://db", KeycloakRealm: "luxus", KeycloakResource: "connect-cli", KeycloakAuthServerURL: "http://keycloak/auth", KeycloakPublicAuthServerURL: "https://auth.example/auth", ObjectStoragePublicURL: "https://storage.example", ObjectStorageAccessKeyID: "application", ObjectStorageSecretKey: "dedicated-secret"}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
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
