package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL              string
	RabbitMQURL              string
	KeycloakRealm            string
	KeycloakAuthServerURL       string
	KeycloakPublicAuthServerURL string
	KeycloakResource            string
	ObjectStorageServiceURL  string
	ObjectStoragePublicURL   string
	ObjectStorageAccessKeyID string
	ObjectStorageSecretKey   string
	CORSOrigins              []string
	Port                     string
	Environment              string
	KeycloakAdminUsername    string
	KeycloakAdminPassword    string
	SMTPHost                 string
	SMTPPort                 int
	SMTPUser                 string
	SMTPPassword             string
	SMTPFrom                 string
	SMTPTLS                  bool
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := NormalizeDatabaseURL(firstNonEmpty(os.Getenv("DATABASE_URL"), os.Getenv("CONNECTION_STRING")))

	cors := strings.Split(os.Getenv("CORS_ORIGINS"), ";")
	var origins []string
	for _, o := range cors {
		if t := strings.TrimSpace(o); t != "" {
			origins = append(origins, t)
		}
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "Development"
	}

	return Config{
		DatabaseURL:              dbURL,
		RabbitMQURL:              os.Getenv("RABBITMQ_URL"),
		KeycloakRealm:            os.Getenv("KEYCLOAK_REALM"),
		KeycloakAuthServerURL:       strings.TrimRight(os.Getenv("KEYCLOAK_AUTH_SERVER_URL"), "/"),
		KeycloakPublicAuthServerURL: strings.TrimRight(firstNonEmpty(os.Getenv("KEYCLOAK_PUBLIC_AUTH_SERVER_URL"), os.Getenv("KEYCLOAK_AUTH_SERVER_URL")), "/"),
		KeycloakResource:            os.Getenv("KEYCLOAK_RESOURCE"),
		ObjectStorageServiceURL:  os.Getenv("OBJECT_STORAGE_SERVICE_URL"),
		ObjectStoragePublicURL:   firstNonEmpty(os.Getenv("OBJECT_STORAGE_PUBLIC_URL"), os.Getenv("OBJECT_STORAGE_SERVICE_URL")),
		ObjectStorageAccessKeyID: os.Getenv("OBJECT_STORAGE_ACCESS_KEY_ID"),
		ObjectStorageSecretKey:   os.Getenv("OBJECT_STORAGE_SECRET_ACCESS_KEY"),
		CORSOrigins:              origins,
		Port:                     port,
		Environment:              env,
		KeycloakAdminUsername:    firstNonEmpty(os.Getenv("KEYCLOAK_ADMIN_USERNAME"), "admin"),
		KeycloakAdminPassword:    os.Getenv("KEYCLOAK_ADMIN_PASSWORD"),
		SMTPHost:                 strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:                 GetEnvInt("SMTP_PORT", 587),
		SMTPUser:                 strings.TrimSpace(os.Getenv("SMTP_USER")),
		SMTPPassword:             os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:                 strings.TrimSpace(os.Getenv("SMTP_FROM")),
		SMTPTLS:                  strings.EqualFold(os.Getenv("SMTP_TLS"), "true"),
	}
}

func (c Config) IsProduction() bool {
	return strings.EqualFold(c.Environment, "Production")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// NormalizeDatabaseURL converts Npgsql-style connection strings to pgx-compatible URLs.
func NormalizeDatabaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "postgres://") || strings.HasPrefix(raw, "postgresql://") {
		return raw
	}
	if !strings.Contains(raw, "=") {
		return raw
	}

	parts := make(map[string]string)
	for _, segment := range strings.Split(raw, ";") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		kv := strings.SplitN(segment, "=", 2)
		if len(kv) != 2 {
			continue
		}
		parts[strings.ToLower(strings.TrimSpace(kv[0]))] = strings.TrimSpace(kv[1])
	}

	host := parts["host"]
	if host == "" {
		host = "localhost"
	}
	user := firstNonEmpty(parts["username"], parts["user"])
	password := parts["password"]
	database := firstNonEmpty(parts["database"], parts["dbname"])
	port := parts["port"]
	if port == "" {
		port = "5432"
	}

	userInfo := user
	if password != "" {
		userInfo = user + ":" + password
	}
	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable", userInfo, host, port, database)
}

func GetEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
