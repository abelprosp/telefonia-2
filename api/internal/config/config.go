package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL                 string
	RabbitMQURL                 string
	KeycloakRealm               string
	KeycloakAuthServerURL       string
	KeycloakPublicAuthServerURL string
	KeycloakResource            string
	ObjectStorageServiceURL     string
	ObjectStoragePublicURL      string
	ObjectStorageAccessKeyID    string
	ObjectStorageSecretKey      string
	CORSOrigins                 []string
	Port                        string
	Environment                 string
	KeycloakAdminUsername       string
	KeycloakAdminPassword       string
	SMTPHost                    string
	SMTPPort                    int
	SMTPUser                    string
	SMTPPassword                string
	SMTPFrom                    string
	SMTPTLS                     bool
	SicrediEnabled              bool
	SicrediSandbox              bool
	SicrediAPIKey               string
	SicrediUsername             string
	SicrediPassword             string
	SicrediCooperativa          string
	SicrediPosto                string
	SicrediCodigoBeneficiario   string
	SicrediWebhookToken         string
	SicrediPublicAPIURL         string
	SicrediAutoRegisterWebhook  bool
	MonitoringTestEnabled       bool
	ImportEnforceStateMachine   bool
	FinancialSFTPHost           string
	FinancialSFTPPort           int
	FinancialSFTPUser           string
	FinancialSFTPPassword       string
	FinancialSFTPPath           string
	ZapSignAPIToken             string
	ZapSignBaseURL              string
	ZapSignSandbox              bool
	ZapSignWebhookToken         string
	FinancialAgentAPIKey        string
	FinancialAgentOrgID         string
	FinancialAgentPublicURL     string
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
	if len(origins) == 0 {
		if rd := strings.TrimSpace(os.Getenv("RAILWAY_PUBLIC_DOMAIN")); rd != "" {
			origins = append(origins, "https://"+rd)
		}
	}

	sicrediPublicURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SICREDI_PUBLIC_API_URL")), "/")
	if sicrediPublicURL == "" {
		if rd := strings.TrimSpace(os.Getenv("RAILWAY_PUBLIC_DOMAIN")); rd != "" {
			sicrediPublicURL = "https://" + rd
		}
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "Development"
	}

	objectStorageServiceURL := strings.TrimRight(strings.TrimSpace(os.Getenv("OBJECT_STORAGE_SERVICE_URL")), "/")
	objectStoragePublicURL := firstNonEmpty(os.Getenv("OBJECT_STORAGE_PUBLIC_URL"), objectStorageServiceURL)
	if strings.EqualFold(env, "Production") {
		objectStoragePublicURL = sanitizeProductionStoragePublicURL(objectStoragePublicURL)
		if objectStorageServiceURL == "" {
			objectStorageServiceURL = "http://minio:9000"
		}
	}

	return Config{
		DatabaseURL:                 dbURL,
		RabbitMQURL:                 os.Getenv("RABBITMQ_URL"),
		KeycloakRealm:               os.Getenv("KEYCLOAK_REALM"),
		KeycloakAuthServerURL:       strings.TrimRight(os.Getenv("KEYCLOAK_AUTH_SERVER_URL"), "/"),
		KeycloakPublicAuthServerURL: strings.TrimRight(firstNonEmpty(os.Getenv("KEYCLOAK_PUBLIC_AUTH_SERVER_URL"), os.Getenv("KEYCLOAK_AUTH_SERVER_URL")), "/"),
		KeycloakResource:            os.Getenv("KEYCLOAK_RESOURCE"),
		ObjectStorageServiceURL:     objectStorageServiceURL,
		ObjectStoragePublicURL:      objectStoragePublicURL,
		ObjectStorageAccessKeyID:    os.Getenv("OBJECT_STORAGE_ACCESS_KEY_ID"),
		ObjectStorageSecretKey:      os.Getenv("OBJECT_STORAGE_SECRET_ACCESS_KEY"),
		CORSOrigins:                 origins,
		Port:                        port,
		Environment:                 env,
		KeycloakAdminUsername:       firstNonEmpty(os.Getenv("KEYCLOAK_ADMIN_USERNAME"), "admin"),
		KeycloakAdminPassword:       os.Getenv("KEYCLOAK_ADMIN_PASSWORD"),
		SMTPHost:                    strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:                    GetEnvInt("SMTP_PORT", 587),
		SMTPUser:                    strings.TrimSpace(os.Getenv("SMTP_USER")),
		SMTPPassword:                os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:                    strings.TrimSpace(os.Getenv("SMTP_FROM")),
		SMTPTLS:                     strings.EqualFold(os.Getenv("SMTP_TLS"), "true"),
		SicrediEnabled:              strings.EqualFold(os.Getenv("SICREDI_ENABLED"), "true"),
		SicrediSandbox:              !strings.EqualFold(os.Getenv("SICREDI_SANDBOX"), "false"),
		SicrediAPIKey:               strings.TrimSpace(os.Getenv("SICREDI_API_KEY")),
		SicrediUsername:             strings.TrimSpace(os.Getenv("SICREDI_USERNAME")),
		SicrediPassword:             os.Getenv("SICREDI_PASSWORD"),
		SicrediCooperativa:          strings.TrimSpace(os.Getenv("SICREDI_COOPERATIVA")),
		SicrediPosto:                strings.TrimSpace(os.Getenv("SICREDI_POSTO")),
		SicrediCodigoBeneficiario:   strings.TrimSpace(os.Getenv("SICREDI_CODIGO_BENEFICIARIO")),
		SicrediWebhookToken:         strings.TrimSpace(os.Getenv("SICREDI_WEBHOOK_TOKEN")),
		SicrediPublicAPIURL:         sicrediPublicURL,
		SicrediAutoRegisterWebhook:  strings.EqualFold(os.Getenv("SICREDI_AUTO_REGISTER_WEBHOOK"), "true"),
		MonitoringTestEnabled:       strings.EqualFold(os.Getenv("MONITORING_TEST_ENABLED"), "true"),
		ImportEnforceStateMachine:   GetEnvBool("IMPORT_ENFORCE_STATE_MACHINE", true),
		FinancialSFTPHost:           strings.TrimSpace(os.Getenv("FINANCIAL_SFTP_HOST")),
		FinancialSFTPPort:           GetEnvInt("FINANCIAL_SFTP_PORT", 22),
		FinancialSFTPUser:           strings.TrimSpace(os.Getenv("FINANCIAL_SFTP_USER")),
		FinancialSFTPPassword:       os.Getenv("FINANCIAL_SFTP_PASSWORD"),
		FinancialSFTPPath:           strings.TrimSpace(firstNonEmpty(os.Getenv("FINANCIAL_SFTP_PATH"), "/inbound")),
		ZapSignAPIToken:             strings.TrimSpace(os.Getenv("ZAPSIGN_API_TOKEN")),
		ZapSignBaseURL:              strings.TrimSpace(os.Getenv("ZAPSIGN_BASE_URL")),
		ZapSignSandbox:              strings.EqualFold(os.Getenv("ZAPSIGN_SANDBOX"), "true"),
		ZapSignWebhookToken:         strings.TrimSpace(os.Getenv("ZAPSIGN_WEBHOOK_TOKEN")),
		FinancialAgentAPIKey:        strings.TrimSpace(os.Getenv("FINANCIAL_AGENT_API_KEY")),
		FinancialAgentOrgID:         strings.TrimSpace(os.Getenv("FINANCIAL_AGENT_ORG_ID")),
		FinancialAgentPublicURL:     strings.TrimRight(firstNonEmpty(os.Getenv("FINANCIAL_AGENT_PUBLIC_URL"), sicrediPublicURL), "/"),
	}
}

func (c Config) FinancialSFTPConfigured() bool {
	return c.FinancialSFTPHost != "" && c.FinancialSFTPUser != "" && c.FinancialSFTPPassword != ""
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
		// Railway e outros PaaS usam sslmode=require; pgx aceita a URL como está.
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
	sslMode := firstNonEmpty(parts["ssl mode"], parts["sslmode"])
	if sslMode == "" {
		switch strings.ToLower(host) {
		case "localhost", "127.0.0.1", "::1", "postgres":
			sslMode = "disable"
		default:
			sslMode = "require"
		}
	}

	userInfo := user
	if password != "" {
		userInfo = user + ":" + password
	}
	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", userInfo, host, port, database, url.QueryEscape(sslMode))
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

// sanitizeProductionStoragePublicURL forces a browser-reachable HTTPS endpoint.
// VPS .env files often set http://IP:19000 which browsers block as mixed content.
func sanitizeProductionStoragePublicURL(raw string) string {
	const secureDefault = "https://telefonia.redobrai.online"
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		return secureDefault
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return secureDefault
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "minio" {
		return secureDefault
	}
	// Bare IP hosts are almost always internal MinIO bindings and break HTTPS pages.
	if looksLikeIP(host) {
		return secureDefault
	}
	// Path suffixes like /storage break MinIO signature verification.
	if u.Path != "" && u.Path != "/" {
		return secureDefault
	}
	return raw
}

func looksLikeIP(host string) bool {
	if host == "" {
		return false
	}
	for _, r := range host {
		if (r >= '0' && r <= '9') || r == '.' || r == ':' {
			continue
		}
		return false
	}
	return strings.Contains(host, ".") || strings.Contains(host, ":")
}

// GetEnvBool lê um booleano de ambiente. Ausente = def. "false"/"0"/"no" desligam.
func GetEnvBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
