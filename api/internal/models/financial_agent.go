package models

import "time"

type FinancialAgentLookupInput struct {
	Query          string `json:"query"`
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Document       string `json:"document"`
	LineNumber     string `json:"line_number"`
	InvoiceNumber  string `json:"invoice_number"`
	WhatsAppNumber string `json:"whatsapp_number"`
}

type FinancialAgentCustomer struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	LegalName        *string  `json:"legal_name,omitempty"`
	CpfCnpj          string   `json:"cpf_cnpj"`
	Active           bool     `json:"active"`
	BillingEmail     *string  `json:"billing_email,omitempty"`
	PhoneLineNumbers []string `json:"phone_line_numbers"`
	MatchReason      string   `json:"match_reason"`
	OpenInvoiceCount int      `json:"open_invoice_count"`
	OverdueCount     int      `json:"overdue_count"`
	OverdueBalance   float64  `json:"overdue_balance"`
}

type FinancialAgentLookupResponse struct {
	Query   string                   `json:"query"`
	Count   int                      `json:"count"`
	Items   []FinancialAgentCustomer `json:"items"`
	Message string                   `json:"message,omitempty"`
}

type FinancialAgentListInvoicesInput struct {
	CustomerID     string `json:"customer_id"`
	Query          string `json:"query"`
	InvoiceNumber  string `json:"invoice_number"`
	DueDate        string `json:"due_date"`
	Status         string `json:"status"`
	OverdueOnly    bool   `json:"overdue_only"`
	WhatsAppNumber string `json:"whatsapp_number"`
}

type FinancialAgentInvoice struct {
	ID                    string     `json:"id"`
	CustomerID            string     `json:"customer_id"`
	CustomerName          string     `json:"customer_name"`
	InvoiceNumber         string     `json:"invoice_number"`
	IssueDate             time.Time  `json:"issue_date"`
	DueDate               time.Time  `json:"due_date"`
	Amount                float64    `json:"amount"`
	Status                string     `json:"status"`
	Overdue               bool       `json:"overdue"`
	DaysOverdue           int        `json:"days_overdue"`
	PhoneLineNumber       *string    `json:"phone_line_number,omitempty"`
	AccountsReceivableID  *string    `json:"accounts_receivable_id,omitempty"`
	SicrediNossoNumero    *string    `json:"sicredi_nosso_numero,omitempty"`
	SicrediLinhaDigitavel *string    `json:"sicredi_linha_digitavel,omitempty"`
	SicrediPixQrCode      *string    `json:"sicredi_pix_qr_code,omitempty"`
	SicrediPixTxID        *string    `json:"sicredi_pix_tx_id,omitempty"`
	SicrediBoletoStatus   *string    `json:"sicredi_boleto_status,omitempty"`
	SicrediPaidAt         *time.Time `json:"sicredi_paid_at,omitempty"`
	HasBoletoPDF          bool       `json:"has_boleto_pdf"`
	Paid                  bool       `json:"paid"`
}

type FinancialAgentListInvoicesResponse struct {
	Count int                     `json:"count"`
	Items []FinancialAgentInvoice `json:"items"`
}

type FinancialAgentInvoicePackageInput struct {
	InvoiceID      string `json:"invoice_id"`
	CustomerID     string `json:"customer_id"`
	Query          string `json:"query"`
	WhatsAppNumber string `json:"whatsapp_number"`
}

type FinancialAgentInvoicePackage struct {
	Invoice        FinancialAgentInvoice `json:"invoice"`
	WhatsAppText   string                `json:"whatsapp_text"`
	PixCopiaCola   string                `json:"pix_copia_cola,omitempty"`
	LinhaDigitavel string                `json:"linha_digitavel,omitempty"`
	BoletoPDFURL   string                `json:"boleto_pdf_url,omitempty"`
	InvoiceHTMLURL string                `json:"invoice_html_url,omitempty"`
	HasBoletoPDF   bool                  `json:"has_boleto_pdf"`
	CompanyName    string                `json:"company_name"`
}

type FinancialAgentDelinquencyInput struct {
	CustomerID     string `json:"customer_id"`
	Query          string `json:"query"`
	WhatsAppNumber string `json:"whatsapp_number"`
	SyncSicredi    bool   `json:"sync_sicredi"`
}

type FinancialAgentDelinquencyInvoice struct {
	FinancialAgentInvoice
	LateFeeAmount    float64 `json:"late_fee_amount"`
	InterestAmount   float64 `json:"interest_amount"`
	TotalWithCharges float64 `json:"total_with_charges"`
	SystemStatus     string  `json:"system_status"`
	SicrediStatus    string  `json:"sicredi_status"`
	SicrediSituacao  string  `json:"sicredi_situacao,omitempty"`
	Delinquent       bool    `json:"delinquent"`
}

type FinancialAgentDelinquencyResponse struct {
	Customer         *FinancialAgentCustomer            `json:"customer,omitempty"`
	Delinquent       bool                               `json:"delinquent"`
	OverdueCount     int                                `json:"overdue_count"`
	OverdueBalance   float64                            `json:"overdue_balance"`
	TotalWithCharges float64                            `json:"total_with_charges"`
	SicrediConnected bool                               `json:"sicredi_connected"`
	SicrediSyncNote  string                             `json:"sicredi_sync_note,omitempty"`
	Items            []FinancialAgentDelinquencyInvoice `json:"items"`
	Summary          string                             `json:"summary"`
}

type FinancialAgentInterestInput struct {
	InvoiceID string  `json:"invoice_id"`
	Amount    float64 `json:"amount"`
	DueDate   string  `json:"due_date"`
	AsOf      string  `json:"as_of"`
}

type FinancialAgentInterestResponse struct {
	InvoiceID           string    `json:"invoice_id,omitempty"`
	InvoiceNumber       string    `json:"invoice_number,omitempty"`
	OriginalAmount      float64   `json:"original_amount"`
	DueDate             time.Time `json:"due_date"`
	AsOf                time.Time `json:"as_of"`
	DaysOverdue         int       `json:"days_overdue"`
	Overdue             bool      `json:"overdue"`
	LateFeePercentage   float64   `json:"late_fee_percentage"`
	InterestRateMonthly float64   `json:"interest_rate_monthly"`
	LateFeeAmount       float64   `json:"late_fee_amount"`
	InterestAmount      float64   `json:"interest_amount"`
	TotalAmount         float64   `json:"total_amount"`
	Explanation         string    `json:"explanation"`
}

type FinancialAgentSyncSicrediInput struct {
	InvoiceID  string `json:"invoice_id"`
	CustomerID string `json:"customer_id"`
}

type FinancialAgentVerifyReceiptInput struct {
	InvoiceID      string  `json:"invoice_id"`
	CustomerID     string  `json:"customer_id"`
	Query          string  `json:"query"`
	WhatsAppNumber string  `json:"whatsapp_number"`
	PayerName      string  `json:"payer_name"`
	Amount         float64 `json:"amount"`
	PaidAt         string  `json:"paid_at"`
	PixTxID        string  `json:"pix_tx_id"`
	PixEndToEnd    string  `json:"pix_end_to_end"`
	BankName       string  `json:"bank_name"`
	NossoNumero    string  `json:"nosso_numero"`
	LinhaDigitavel string  `json:"linha_digitavel"`
	InvoiceNumber  string  `json:"invoice_number"`
	ReceiptText    string  `json:"receipt_text"`
	ForceConfirm   bool    `json:"force_confirm"`
}

type FinancialAgentVerifyReceiptResponse struct {
	Matched           bool                            `json:"matched"`
	Confidence        int                             `json:"confidence"`
	ConfidenceLabel   string                          `json:"confidence_label"`
	CanAutoConfirm    bool                            `json:"can_auto_confirm"`
	Confirmed         bool                            `json:"confirmed"`
	NeedsHumanReview  bool                            `json:"needs_human_review"`
	Reasons           []string                        `json:"reasons"`
	Invoice           *FinancialAgentInvoice          `json:"invoice,omitempty"`
	Interest          *FinancialAgentInterestResponse `json:"interest,omitempty"`
	SicrediSituacao   string                          `json:"sicredi_situacao,omitempty"`
	SicrediLiquidated bool                            `json:"sicredi_liquidated"`
	Summary           string                          `json:"summary"`
}

type FinancialAgentConfirmPaymentInput struct {
	InvoiceID   string  `json:"invoice_id"`
	Amount      float64 `json:"amount"`
	PaymentDate string  `json:"payment_date"`
	Reference   string  `json:"reference"`
	Notes       string  `json:"notes"`
	Force       bool    `json:"force"`
}

type FinancialAgentConfirmPaymentResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	InvoiceID     string  `json:"invoice_id"`
	InvoiceNumber string  `json:"invoice_number"`
	Amount        float64 `json:"amount"`
	Source        string  `json:"source"`
}

type FinancialAgentHealthResponse struct {
	Enabled          bool   `json:"enabled"`
	OrganizationSet  bool   `json:"organization_set"`
	SicrediEnabled   bool   `json:"sicredi_enabled"`
	SicrediConnected bool   `json:"sicredi_connected"`
	Message          string `json:"message"`
}

type FinancialAgentSettingsInput struct {
	Enabled           *bool   `json:"enabled,omitempty"`
	EvolutionAPIURL   *string `json:"evolution_api_url,omitempty"`
	EvolutionAPIKey   *string `json:"evolution_api_key,omitempty"`
	EvolutionInstance *string `json:"evolution_instance,omitempty"`
	N8nWebhookURL     *string `json:"n8n_webhook_url,omitempty"`
}

type FinancialAgentSettingsResponse struct {
	Enabled              bool      `json:"enabled"`
	EvolutionAPIURL      string    `json:"evolution_api_url"`
	EvolutionAPIKeySet   bool      `json:"evolution_api_key_set"`
	EvolutionInstance    string    `json:"evolution_instance"`
	N8nWebhookURL        string    `json:"n8n_webhook_url"`
	UpdatedAt            time.Time `json:"updated_at,omitempty"`
}

type FinancialAgentWhatsAppStatus struct {
	Configured    bool   `json:"configured"`
	State         string `json:"state"`
	Connected     bool   `json:"connected"`
	QRCode        string `json:"qr_code,omitempty"`
	PairingCode   string `json:"pairing_code,omitempty"`
	InstanceName  string `json:"instance_name,omitempty"`
	ProfileName   string `json:"profile_name,omitempty"`
	OwnerJID      string `json:"owner_jid,omitempty"`
	Message       string `json:"message,omitempty"`
}

type FinancialAgentEvent struct {
	ID             string    `json:"id"`
	EventType      string    `json:"event_type"`
	EventLabel     string    `json:"event_label"`
	WhatsAppNumber string    `json:"whatsapp_number,omitempty"`
	CustomerName   string    `json:"customer_name,omitempty"`
	InvoiceNumber  string    `json:"invoice_number,omitempty"`
	Success        bool      `json:"success"`
	Summary        string    `json:"summary"`
	CreatedAt      time.Time `json:"created_at"`
}

type FinancialAgentPanelStats struct {
	TotalToday    int `json:"total_today"`
	SuccessToday  int `json:"success_today"`
	FailedToday   int `json:"failed_today"`
	LookupsToday  int `json:"lookups_today"`
	ReceiptsToday int `json:"receipts_today"`
	PaymentsToday int `json:"payments_today"`
}

type FinancialAgentPanelResponse struct {
	ToolsReady          bool                            `json:"tools_ready"`
	OrganizationSet     bool                            `json:"organization_set"`
	SicrediEnabled      bool                            `json:"sicredi_enabled"`
	SicrediConnected    bool                            `json:"sicredi_connected"`
	Settings            FinancialAgentSettingsResponse  `json:"settings"`
	WhatsApp            FinancialAgentWhatsAppStatus    `json:"whatsapp"`
	Stats               FinancialAgentPanelStats        `json:"stats"`
	Message             string                          `json:"message"`
}
