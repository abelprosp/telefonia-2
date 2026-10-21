package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AuthSecret                      string
	IntakeAPIKey                    string
	IntakeOrgID                     string
	DatabaseURL                     string
	RabbitMQURL                     string
	ObjectStorageServiceURL         string
	ObjectStoragePublicURL          string
	ObjectStorageAccessKeyID        string
	ObjectStorageSecretKey          string
	ObjectStorageBucketName         string
	ObjectStorageMaxObjectBytes     int64
	ObjectStorageMaxDiskObjectBytes int64
	CORSOrigins                     []string
	Port                            string
	Environment                     string
	SMTPHost                        string
	SMTPPort                        int
	SMTPUser                        string
	SMTPPassword                    string
	SMTPFrom                        string
	SMTPTLS                         bool
	SicrediEnabled                  bool
	SicrediSandbox                  bool
	SicrediAPIKey                   string
	SicrediUsername                 string
	SicrediPassword                 string
	SicrediCooperativa              string
	SicrediPosto                    string
	SicrediCodigoBeneficiario       string
	SicrediWebhookToken             string
	SicrediPublicAPIURL             string
	SicrediAutoRegisterWebhook      bool
	MonitoringTestEnabled           bool
	ImportEnforceStateMachine       bool
	FinancialSFTPHost               string
	FinancialSFTPPort               int
	FinancialSFTPUser               string
	FinancialSFTPPassword           string
	FinancialSFTPPath               string
	ZapSignAPIToken                 string
	ZapSignBaseURL                  string
	ZapSignSandbox                  bool
	ZapSignWebhookToken             string
	FinancialAgentAPIKey            string
	FinancialAgentOrgID             string
	FinancialAgentPublicURL         string
	ParceirosIntakeAPIKey           string
	ParceirosIntakeOrgID            string
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
		if objectStorageServiceURL == "" {
			objectStorageServiceURL = "http://minio:9000"
		}
	}

	return Config{
		AuthSecret:                      os.Getenv("AUTH_SECRET"),
		IntakeAPIKey:                    strings.TrimSpace(os.Getenv("TELEFONIA_INTAKE_API_KEY")),
		IntakeOrgID:                     strings.TrimSpace(os.Getenv("TELEFONIA_INTAKE_ORG_ID")),
		DatabaseURL:                     dbURL,
		RabbitMQURL:                     os.Getenv("RABBITMQ_URL"),
		ObjectStorageServiceURL:         objectStorageServiceURL,
		ObjectStoragePublicURL:          objectStoragePublicURL,
		ObjectStorageAccessKeyID:        os.Getenv("OBJECT_STORAGE_ACCESS_KEY_ID"),
		ObjectStorageSecretKey:          os.Getenv("OBJECT_STORAGE_SECRET_ACCESS_KEY"),
		ObjectStorageBucketName:         firstNonEmpty(os.Getenv("OBJECT_STORAGE_BUCKET_NAME"), "luxus-connect"),
		ObjectStorageMaxObjectBytes:     GetEnvInt64("OBJECT_STORAGE_MAX_OBJECT_BYTES", 512*1024*1024),
		ObjectStorageMaxDiskObjectBytes: GetEnvInt64("OBJECT_STORAGE_MAX_DISK_OBJECT_BYTES", 2*1024*1024*1024),
		CORSOrigins:                     origins,
		Port:                            port,
		Environment:                     env,
		SMTPHost:                        strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:                        GetEnvInt("SMTP_PORT", 587),
		SMTPUser:                        strings.TrimSpace(os.Getenv("SMTP_USER")),
		SMTPPassword:                    os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:                        strings.TrimSpace(os.Getenv("SMTP_FROM")),
		SMTPTLS:                         strings.EqualFold(os.Getenv("SMTP_TLS"), "true"),
		SicrediEnabled:                  strings.EqualFold(os.Getenv("SICREDI_ENABLED"), "true"),
		SicrediSandbox:                  !strings.EqualFold(os.Getenv("SICREDI_SANDBOX"), "false"),
		SicrediAPIKey:                   strings.TrimSpace(os.Getenv("SICREDI_API_KEY")),
		SicrediUsername:                 strings.TrimSpace(os.Getenv("SICREDI_USERNAME")),
		SicrediPassword:                 os.Getenv("SICREDI_PASSWORD"),
		SicrediCooperativa:              strings.TrimSpace(os.Getenv("SICREDI_COOPERATIVA")),
		SicrediPosto:                    strings.TrimSpace(os.Getenv("SICREDI_POSTO")),
		SicrediCodigoBeneficiario:       strings.TrimSpace(os.Getenv("SICREDI_CODIGO_BENEFICIARIO")),
		SicrediWebhookToken:             strings.TrimSpace(os.Getenv("SICREDI_WEBHOOK_TOKEN")),
		SicrediPublicAPIURL:             sicrediPublicURL,
		SicrediAutoRegisterWebhook:      strings.EqualFold(os.Getenv("SICREDI_AUTO_REGISTER_WEBHOOK"), "true"),
		MonitoringTestEnabled:           strings.EqualFold(os.Getenv("MONITORING_TEST_ENABLED"), "true"),
		ImportEnforceStateMachine:       GetEnvBool("IMPORT_ENFORCE_STATE_MACHINE", true),
		FinancialSFTPHost:               strings.TrimSpace(os.Getenv("FINANCIAL_SFTP_HOST")),
		FinancialSFTPPort:               GetEnvInt("FINANCIAL_SFTP_PORT", 22),
		FinancialSFTPUser:               strings.TrimSpace(os.Getenv("FINANCIAL_SFTP_USER")),
		FinancialSFTPPassword:           os.Getenv("FINANCIAL_SFTP_PASSWORD"),
		FinancialSFTPPath:               strings.TrimSpace(firstNonEmpty(os.Getenv("FINANCIAL_SFTP_PATH"), "/inbound")),
		ZapSignAPIToken:                 strings.TrimSpace(os.Getenv("ZAPSIGN_API_TOKEN")),
		ZapSignBaseURL:                  strings.TrimSpace(os.Getenv("ZAPSIGN_BASE_URL")),
		ZapSignSandbox:                  strings.EqualFold(os.Getenv("ZAPSIGN_SANDBOX"), "true"),
		ZapSignWebhookToken:             strings.TrimSpace(os.Getenv("ZAPSIGN_WEBHOOK_TOKEN")),
		FinancialAgentAPIKey:            strings.TrimSpace(os.Getenv("FINANCIAL_AGENT_API_KEY")),
		FinancialAgentOrgID:             strings.TrimSpace(os.Getenv("FINANCIAL_AGENT_ORG_ID")),
		FinancialAgentPublicURL:         strings.TrimRight(firstNonEmpty(os.Getenv("FINANCIAL_AGENT_PUBLIC_URL"), sicrediPublicURL), "/"),
		ParceirosIntakeAPIKey:           strings.TrimSpace(os.Getenv("PARCEIROS_INTAKE_API_KEY")),
		ParceirosIntakeOrgID:            strings.TrimSpace(os.Getenv("PARCEIROS_INTAKE_ORG_ID")),
	}
}

func (c Config) FinancialSFTPConfigured() bool {
	return c.FinancialSFTPHost != "" && c.FinancialSFTPUser != "" && c.FinancialSFTPPassword != ""
}

func (c Config) IsProduction() bool {
	return strings.EqualFold(c.Environment, "Production")
}

// Validate fails closed rather than replacing invalid endpoints or accepting
// development credentials in production.
func (c Config) Validate() error {
	if (c.IntakeAPIKey == "") != (c.IntakeOrgID == "") {
		return fmt.Errorf("TELEFONIA_INTAKE_API_KEY and TELEFONIA_INTAKE_ORG_ID must be configured together")
	}
	if c.IntakeAPIKey != "" && (len(c.IntakeAPIKey) < 32 || len(c.IntakeOrgID) > 36) {
		return fmt.Errorf("intake requires a key of at least 32 characters and a valid organization ID")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if len(c.AuthSecret) < 32 {
		return fmt.Errorf("AUTH_SECRET requires at least 32 characters")
	}
	if len(c.CORSOrigins) == 0 {
		return fmt.Errorf("CORS_ORIGINS must contain the frontend origin")
	}
	// Wildcard CORS is never acceptable because responses allow credentials.
	for _, origin := range c.CORSOrigins {
		if strings.Contains(origin, "*") {
			return fmt.Errorf("CORS_ORIGINS must not contain wildcards")
		}
	}
	if !c.IsProduction() {
		return nil
	}
	for name, raw := range map[string]string{
		"OBJECT_STORAGE_PUBLIC_URL": c.ObjectStoragePublicURL,
	} {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return fmt.Errorf("%s must be a public HTTPS URL", name)
		}
		if name == "OBJECT_STORAGE_PUBLIC_URL" && (u.Path != "" && u.Path != "/" || net.ParseIP(u.Hostname()) != nil || u.Hostname() == "localhost" || u.Hostname() == "minio") {
			return fmt.Errorf("OBJECT_STORAGE_PUBLIC_URL must use a public DNS name without a path")
		}
	}
	if c.ObjectStorageAccessKeyID == "" || c.ObjectStorageSecretKey == "" || c.ObjectStorageAccessKeyID == "minioadmin" || c.ObjectStorageSecretKey == "minioadmin" {
		return fmt.Errorf("production object storage requires dedicated application credentials")
	}
	for _, origin := range c.CORSOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme != "https" || u.Host == "" || strings.Contains(origin, "*") || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return fmt.Errorf("CORS_ORIGINS must contain exact HTTPS origins")
		}
	}
	return nil
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

	u := &url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, port), Path: "/" + database, User: url.UserPassword(user, password)}
	query := url.Values{"sslmode": []string{sslMode}}
	u.RawQuery = query.Encode()
	return u.String()
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

func GetEnvInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
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
