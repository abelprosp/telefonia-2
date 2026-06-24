package invoicelayout

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

type Theme struct {
	PrimaryColor          string `json:"primaryColor"`
	AccentColor           string `json:"accentColor"`
	BorderColor           string `json:"borderColor"`
	HeaderBackground      string `json:"headerBackground"`
	TitleColor            string `json:"titleColor"`
	TextColor             string `json:"textColor"`
	TableHeaderBackground string `json:"tableHeaderBackground"`
	BorderRadius          int    `json:"borderRadius"`
}

type Branding struct {
	LogoDataUrl   string `json:"logoDataUrl"`
	CompanyName   string `json:"companyName"`
	Tagline       string `json:"tagline"`
	DocumentTitle string `json:"documentTitle"`
}

type SectionToggle struct {
	Enabled bool   `json:"enabled"`
	Title   string `json:"title,omitempty"`
}

type Sections struct {
	UserData            SectionToggle `json:"userData"`
	AccountValue        SectionToggle `json:"accountValue"`
	BillingDates        SectionToggle `json:"billingDates"`
	AccountSummary      SectionToggle `json:"accountSummary"`
	DetailedConsumption SectionToggle `json:"detailedConsumption"`
}

type Labels struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	TotalServices  string `json:"totalServices"`
	Discounts      string `json:"discounts"`
	BillingPeriod  string `json:"billingPeriod"`
	ReferenceMonth string `json:"referenceMonth"`
	DueDate        string `json:"dueDate"`
	Description    string `json:"description"`
	Quantity       string `json:"quantity"`
	Type           string `json:"type"`
	UnitPrice      string `json:"unitPrice"`
	Total          string `json:"total"`
	TotalLabel     string `json:"totalLabel"`
}

type Config struct {
	Theme    Theme    `json:"theme"`
	Branding Branding `json:"branding"`
	Sections Sections `json:"sections"`
	Labels   Labels   `json:"labels"`
}

type LineItem struct {
	Description string
	Quantity    string
	Type        string
	UnitPrice   string
	Total       string
}

type RenderData struct {
	CustomerName      string
	CustomerAddress   string
	CustomerPhone     string
	CustomerDocument  string
	InvoiceNumber     string
	InvoiceAmount     string
	InvoiceDueDate    string
	InvoiceIssueDate  string
	ReferenceMonth    string
	PeriodStart       string
	PeriodEnd         string
	ServicesTotal     string
	Discounts         string
	Description       string
	LineItems         []LineItem
	ConsumptionHTML   string
}

func ParseConfig(raw json.RawMessage) (Config, error) {
	var cfg Config
	if len(raw) == 0 {
		return DefaultConfig(), nil
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}
	def := DefaultConfig()
	if cfg.Theme.PrimaryColor == "" {
		cfg.Theme = def.Theme
	}
	if cfg.Branding.CompanyName == "" {
		cfg.Branding = def.Branding
	}
	if cfg.Labels.Name == "" {
		cfg.Labels = def.Labels
	}
	return cfg, nil
}

func DefaultConfig() Config {
	return Config{
		Theme: Theme{
			PrimaryColor:          "#4a4a4a",
			AccentColor:           "#00a0c6",
			BorderColor:           "#222222",
			HeaderBackground:      "#ffffff",
			TitleColor:            "#1a1a1a",
			TextColor:             "#333333",
			TableHeaderBackground: "#f7f7f7",
			BorderRadius:          12,
		},
		Branding: Branding{
			CompanyName:   "LUXUS",
			Tagline:       "SOLUÇÃO EM TELEFONIA",
			DocumentTitle: "Detalhamento da Fatura",
		},
		Sections: Sections{
			UserData:            SectionToggle{Enabled: true, Title: "Dados do Usuário"},
			AccountValue:        SectionToggle{Enabled: true, Title: "VALOR DA SUA CONTA"},
			BillingDates:        SectionToggle{Enabled: true},
			AccountSummary:      SectionToggle{Enabled: true, Title: "Resumo da Conta"},
			DetailedConsumption: SectionToggle{Enabled: true, Title: "Consumo Detalhado"},
		},
		Labels: Labels{
			Name:           "Nome:",
			Address:        "Endereço:",
			Phone:          "Número do telefone:",
			TotalServices:  "Total Serviços:",
			Discounts:      "Descontos:",
			BillingPeriod:  "Período de faturamento:",
			ReferenceMonth: "Mês de referência:",
			DueDate:        "Data de Vencimento:",
			Description:    "Descrição",
			Quantity:       "Quantidade",
			Type:           "Tipo",
			UnitPrice:      "Preço Unitário",
			Total:          "Total",
			TotalLabel:     "Total:",
		},
	}
}

func Render(cfg Config, data RenderData) string {
	t := cfg.Theme
	l := cfg.Labels
	radius := t.BorderRadius
	if radius <= 0 {
		radius = 12
	}
	box := fmt.Sprintf("border:1px solid %s;border-radius:%dpx;padding:14px 16px;margin-bottom:12px;background:%s;",
		t.BorderColor, radius, t.HeaderBackground)

	var b strings.Builder
	b.WriteString(`<div style="font-family:Arial,Helvetica,sans-serif;max-width:820px;margin:0 auto;color:`)
	b.WriteString(t.TextColor)
	b.WriteString(`;font-size:13px;line-height:1.45;">`)

	b.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="margin-bottom:12px;"><tr>`)
	b.WriteString(`<td width="50%" style="vertical-align:top;padding-right:6px;"><div style="`)
	b.WriteString(box)
	b.WriteString(`text-align:center;">`)
	if cfg.Branding.LogoDataUrl != "" {
		b.WriteString(`<img src="`)
		b.WriteString(html.EscapeString(cfg.Branding.LogoDataUrl))
		b.WriteString(`" alt="Logo" style="max-height:72px;max-width:100%;object-fit:contain;margin-bottom:8px;" />`)
	}
	b.WriteString(`<div style="font-size:22px;font-weight:700;color:`)
	b.WriteString(t.PrimaryColor)
	b.WriteString(`;letter-spacing:1px;">`)
	b.WriteString(html.EscapeString(cfg.Branding.CompanyName))
	b.WriteString(`</div><div style="font-size:11px;color:`)
	b.WriteString(t.TextColor)
	b.WriteString(`;">`)
	b.WriteString(html.EscapeString(cfg.Branding.Tagline))
	b.WriteString(`</div></div></td>`)
	b.WriteString(`<td width="50%" style="vertical-align:top;padding-left:6px;"><div style="`)
	b.WriteString(box)
	b.WriteString(`text-align:center;height:100%;"><h1 style="margin:0;font-size:26px;color:`)
	b.WriteString(t.TitleColor)
	b.WriteString(`;">`)
	b.WriteString(html.EscapeString(cfg.Branding.DocumentTitle))
	b.WriteString(`</h1></div></td></tr></table>`)

	if cfg.Sections.UserData.Enabled {
		b.WriteString(`<div style="`)
		b.WriteString(box)
		b.WriteString(`">`)
		if cfg.Sections.UserData.Title != "" {
			b.WriteString(`<div style="font-weight:700;margin-bottom:8px;">`)
			b.WriteString(html.EscapeString(cfg.Sections.UserData.Title))
			b.WriteString(`</div>`)
		}
		b.WriteString(row(l.Name, data.CustomerName))
		b.WriteString(row(l.Address, data.CustomerAddress))
		b.WriteString(row(l.Phone, data.CustomerPhone))
		b.WriteString(`</div>`)
	}

	if cfg.Sections.AccountValue.Enabled {
		b.WriteString(`<div style="`)
		b.WriteString(box)
		b.WriteString(`">`)
		b.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="0">`)
		b.WriteString(amountRow(cfg.Sections.AccountValue.Title, data.InvoiceAmount, true, t))
		b.WriteString(amountRow(l.TotalServices, data.ServicesTotal, false, t))
		b.WriteString(amountRow(l.Discounts, data.Discounts, false, t))
		b.WriteString(`</table></div>`)
	}

	if cfg.Sections.BillingDates.Enabled {
		b.WriteString(`<div style="`)
		b.WriteString(box)
		b.WriteString(`">`)
		b.WriteString(row(l.BillingPeriod, data.PeriodStart+` a `+data.PeriodEnd))
		b.WriteString(row(l.ReferenceMonth, data.ReferenceMonth))
		b.WriteString(row(l.DueDate, data.InvoiceDueDate))
		b.WriteString(`</div>`)
	}

	if cfg.Sections.AccountSummary.Enabled {
		b.WriteString(`<div style="`)
		b.WriteString(box)
		b.WriteString(`">`)
		if cfg.Sections.AccountSummary.Title != "" {
			b.WriteString(`<div style="font-weight:700;margin-bottom:8px;">`)
			b.WriteString(html.EscapeString(cfg.Sections.AccountSummary.Title))
			b.WriteString(`</div>`)
		}
		b.WriteString(renderLineItemsTable(cfg, data))
		b.WriteString(`</div>`)
	}

	if cfg.Sections.DetailedConsumption.Enabled {
		b.WriteString(`<div style="`)
		b.WriteString(box)
		b.WriteString(`">`)
		if cfg.Sections.DetailedConsumption.Title != "" {
			b.WriteString(`<div style="font-weight:700;margin-bottom:8px;">`)
			b.WriteString(html.EscapeString(cfg.Sections.DetailedConsumption.Title))
			b.WriteString(`</div>`)
		}
		if data.ConsumptionHTML != "" {
			b.WriteString(data.ConsumptionHTML)
		} else {
			b.WriteString(defaultConsumptionHTML(t))
		}
		b.WriteString(`</div>`)
	}

	b.WriteString(`</div>`)
	return WrapHTMLDocument(b.String())
}

func row(label, value string) string {
	return fmt.Sprintf(
		`<div style="margin-bottom:4px;"><strong>%s</strong> %s</div>`,
		html.EscapeString(label),
		html.EscapeString(value),
	)
}

func amountRow(label, value string, highlight bool, t Theme) string {
	weight := "normal"
	size := "13px"
	if highlight {
		weight = "700"
		size = "15px"
	}
	return fmt.Sprintf(
		`<tr><td style="padding:4px 0;font-weight:%s;font-size:%s;">%s</td><td style="padding:4px 0;text-align:right;font-weight:%s;font-size:%s;">%s</td></tr>`,
		weight, size, html.EscapeString(label), weight, size, html.EscapeString(value),
	)
}

func renderLineItemsTable(cfg Config, data RenderData) string {
	t := cfg.Theme
	l := cfg.Labels
	items := data.LineItems
	if len(items) == 0 && data.Description != "" {
		items = []LineItem{{
			Description: data.Description,
			Quantity:    "1",
			Type:        "Mensal",
			UnitPrice:   data.InvoiceAmount,
			Total:       data.InvoiceAmount,
		}}
	}
	var b strings.Builder
	b.WriteString(`<table role="presentation" width="100%" cellspacing="0" cellpadding="6" style="border-collapse:collapse;font-size:12px;">`)
	b.WriteString(`<thead><tr style="background:`)
	b.WriteString(t.TableHeaderBackground)
	b.WriteString(`;"><th align="left">`)
	b.WriteString(html.EscapeString(l.Description))
	b.WriteString(`</th><th>`)
	b.WriteString(html.EscapeString(l.Quantity))
	b.WriteString(`</th><th>`)
	b.WriteString(html.EscapeString(l.Type))
	b.WriteString(`</th><th align="right">`)
	b.WriteString(html.EscapeString(l.UnitPrice))
	b.WriteString(`</th><th align="right">`)
	b.WriteString(html.EscapeString(l.Total))
	b.WriteString(`</th></tr></thead><tbody>`)
	for _, item := range items {
		b.WriteString(`<tr><td>`)
		b.WriteString(html.EscapeString(item.Description))
		b.WriteString(`</td><td align="center">`)
		b.WriteString(html.EscapeString(item.Quantity))
		b.WriteString(`</td><td align="center">`)
		b.WriteString(html.EscapeString(item.Type))
		b.WriteString(`</td><td align="right">`)
		b.WriteString(html.EscapeString(item.UnitPrice))
		b.WriteString(`</td><td align="right">`)
		b.WriteString(html.EscapeString(item.Total))
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`<tr><td colspan="4" align="right" style="font-weight:700;padding-top:8px;">`)
	b.WriteString(html.EscapeString(l.TotalLabel))
	b.WriteString(`</td><td align="right" style="font-weight:700;padding-top:8px;">`)
	b.WriteString(html.EscapeString(data.InvoiceAmount))
	b.WriteString(`</td></tr></tbody></table>`)
	return b.String()
}

func defaultConsumptionHTML(t Theme) string {
	return fmt.Sprintf(`<div style="font-size:12px;">
<div style="font-weight:700;margin-bottom:6px;">PACOTE DE TORPEDOS CONSUMIDOS: 0</div>
<table role="presentation" width="100%%" cellspacing="0" cellpadding="4" style="border-collapse:collapse;">
<tr style="background:%s;font-weight:700;"><td>CHAMADAS LOCAIS</td><td align="right">R$</td></tr>
<tr><td style="padding-left:12px;">Fixo</td><td align="right">0,00</td></tr>
<tr><td style="padding-left:12px;">Móvel Outras Operadoras</td><td align="right">0,00</td></tr>
<tr><td style="padding-left:12px;">Móvel Vivo</td><td align="right">0,00</td></tr>
<tr style="background:%s;font-weight:700;"><td>CHAMADAS ESTADUAIS</td><td align="right">R$</td></tr>
<tr><td style="padding-left:12px;">Fixo</td><td align="right">0,00</td></tr>
<tr><td style="padding-left:12px;">Móvel Outras Operadoras</td><td align="right">0,00</td></tr>
<tr><td style="padding-left:12px;">Móvel Vivo</td><td align="right">0,00</td></tr>
<tr style="background:%s;font-weight:700;"><td>TORPEDOS</td><td align="right">R$</td></tr>
<tr><td style="padding-left:12px;">Móvel Vivo</td><td align="right">0,00</td></tr>
</table></div>`, t.TableHeaderBackground, t.TableHeaderBackground, t.TableHeaderBackground)
}
