package sicredi

import (
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/config"
)

type Config struct {
	Enabled            bool
	Sandbox            bool
	Production         bool
	APIKey             string
	Username           string
	Password           string
	Cooperativa        string
	Posto              string
	CodigoBeneficiario string
	WebhookToken       string
	PublicAPIURL       string
}

func ConfigFrom(cfg config.Config) Config {
	username := strings.TrimSpace(cfg.SicrediUsername)
	if username == "" && cfg.SicrediCodigoBeneficiario != "" && cfg.SicrediCooperativa != "" {
		username = strings.TrimSpace(cfg.SicrediCodigoBeneficiario) + strings.TrimSpace(cfg.SicrediCooperativa)
	}
	return Config{
		Enabled:            cfg.SicrediEnabled,
		Sandbox:            cfg.SicrediSandbox,
		Production:         cfg.IsProduction(),
		APIKey:             strings.TrimSpace(cfg.SicrediAPIKey),
		Username:           username,
		Password:           cfg.SicrediPassword,
		Cooperativa:        strings.TrimSpace(cfg.SicrediCooperativa),
		Posto:              strings.TrimSpace(cfg.SicrediPosto),
		CodigoBeneficiario: strings.TrimSpace(cfg.SicrediCodigoBeneficiario),
		WebhookToken:       strings.TrimSpace(cfg.SicrediWebhookToken),
		PublicAPIURL:       strings.TrimSpace(cfg.SicrediPublicAPIURL),
	}
}

func (c Config) EnabledAndConfigured() bool {
	if !c.Enabled {
		return false
	}
	return c.APIKey != "" && c.Username != "" && c.Password != "" &&
		c.Cooperativa != "" && c.Posto != "" && c.CodigoBeneficiario != ""
}

func (c Config) AuthURL() string {
	if c.Sandbox {
		return "https://api-parceiro.sicredi.com.br/sb/auth/openapi/token"
	}
	return "https://api-parceiro.sicredi.com.br/auth/openapi/token"
}

func (c Config) BoletoURL() string {
	if c.Sandbox {
		return "https://api-parceiro.sicredi.com.br/sb/cobranca/boleto/v1/boletos"
	}
	return "https://api-parceiro.sicredi.com.br/cobranca/boleto/v1/boletos"
}

func (c Config) LiquidadosDiaURL() string {
	return c.BoletoURL() + "/liquidados/dia"
}

func (c Config) PdfURL() string {
	return c.BoletoURL() + "/pdf"
}

func (c Config) WebhookContratoURL() string {
	if c.Sandbox {
		return "https://api-parceiro.sicredi.com.br/sb/cobranca/boleto/v1/webhook/contrato"
	}
	return "https://api-parceiro.sicredi.com.br/cobranca/boleto/v1/webhook/contrato"
}

func (c Config) WebhookContratosURL() string {
	if c.Sandbox {
		return "https://api-parceiro.sicredi.com.br/sb/cobranca/boleto/v1/webhook/contratos"
	}
	return "https://api-parceiro.sicredi.com.br/cobranca/boleto/v1/webhook/contratos"
}

func (c Config) BoletoByNossoNumeroURL(nossoNumero string) string {
	return c.BoletoURL() + "/" + strings.TrimSpace(nossoNumero)
}
