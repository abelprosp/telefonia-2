package models

import "time"

// --- Shared ---

type PresignedURLModel struct {
	URL          string    `json:"url"`
	HTTPMethod   string    `json:"http_method"`
	ExpiresAtUTC time.Time `json:"expires_at_utc"`
}

// --- Providers ---

type CreateProviderInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UpdateProviderInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ListProvidersResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Active bool   `json:"active"`
}

type CreateProviderResponse = ListProvidersResponse

type GetProviderPlanServiceResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Active    bool     `json:"active"`
	Recurring bool     `json:"recurring"`
	Price     *float64 `json:"price"`
}

type GetProviderPlanResponse struct {
	ID       string                           `json:"id"`
	Name     string                           `json:"name"`
	Code     string                           `json:"code"`
	Services []GetProviderPlanServiceResponse `json:"services"`
}

type GetProviderResponse struct {
	ID             string                    `json:"id"`
	OrganizationID string                    `json:"organization_id"`
	Name           string                    `json:"name"`
	Slug           string                    `json:"slug"`
	Active         bool                      `json:"active"`
	Plans          []GetProviderPlanResponse `json:"plans"`
}

// --- Customers ---

type CreateCustomerAddressInput struct {
	Street       string  `json:"street"`
	Number       string  `json:"number"`
	Neighborhood string  `json:"neighborhood"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	ZipCode      string  `json:"zip_code"`
	Complement   *string `json:"complement"`
	Country      string  `json:"country"`
}

type CreateCustomerInput struct {
	ProviderID                      string                       `json:"provider_id"`
	Type                            string                       `json:"type"`
	Name                            string                       `json:"name"`
	Document                        string                       `json:"document"`
	LegalName                       *string                      `json:"legal_name"`
	StateRegistration               *string                      `json:"state_registration"`
	BirthOrOpeningDate              *string                      `json:"birth_or_opening_date"`
	ResponsibleSalespersonUserID    *string                      `json:"responsible_salesperson_user_id"`
	Addresses                       []CreateCustomerAddressInput `json:"addresses"`
}

type UpdateCustomerInput struct {
	Name                         string  `json:"name"`
	LegalName                    *string `json:"legal_name"`
	StateRegistration            *string `json:"state_registration"`
	BirthOrOpeningDate           *string `json:"birth_or_opening_date"`
	ResponsibleSalespersonUserID *string `json:"responsible_salesperson_user_id"`
	BillingEmail                 *string `json:"billing_email"`
}

type ListCustomerResponse struct {
	ID                           string     `json:"id"`
	Active                       bool       `json:"active"`
	Type                         string     `json:"type"`
	Name                         string     `json:"name"`
	CpfCnpj                      string     `json:"cpf_cnpj"`
	StateRegistration            *string    `json:"state_registration"`
	LegalName                    *string    `json:"legal_name"`
	BirthOrOpeningDate           *time.Time `json:"birth_or_opening_date"`
	ResponsibleSalespersonUserID *string    `json:"responsible_salesperson_user_id"`
	BillingEmail                 *string    `json:"billing_email"`
}

type CreateCustomerResponse = ListCustomerResponse

type CustomerProviderLinkResponse struct {
	CustomerID   string     `json:"customer_id"`
	ProviderID   string     `json:"provider_id"`
	ProviderName string     `json:"provider_name"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	IsActive     bool       `json:"is_active"`
}

type CustomerPhoneLineLinkResponse struct {
	CustomerID         string     `json:"customer_id"`
	PhoneLineID        string     `json:"phone_line_id"`
	PhoneLineNumber    string     `json:"phone_line_number"`
	PhoneLineStatus    string     `json:"phone_line_status"`
	LineClassification string     `json:"line_classification"`
	StartDate          time.Time  `json:"start_date"`
	EndDate            *time.Time `json:"end_date"`
	IsActive           bool       `json:"is_active"`
}

type CustomerAttachmentResponse struct {
	ID               string     `json:"id"`
	Title            *string    `json:"title"`
	OriginalFileName string     `json:"original_file_name"`
	StorageBucket    string     `json:"storage_bucket"`
	StorageObjectKey string     `json:"storage_object_key"`
	ContentType      *string    `json:"content_type"`
	SizeBytes        *int64     `json:"size_bytes"`
	UploadedAtUTC    time.Time  `json:"uploaded_at_utc"`
}

type RegisterCustomerAttachmentInput struct {
	Title            *string `json:"title"`
	OriginalFileName string  `json:"original_file_name"`
	StorageBucket    string  `json:"storage_bucket"`
	StorageObjectKey string  `json:"storage_object_key"`
	ContentType      *string `json:"content_type"`
	SizeBytes        *int64  `json:"size_bytes"`
}

type ManuallyReleaseCustomerInput struct {
	Justification string `json:"justification"`
}

type BillingReadinessManualReleaseDto struct {
	Justification     string    `json:"justification"`
	ReleasedAt        time.Time `json:"released_at"`
	ReleasedByUserID  string    `json:"released_by_user_id"`
}

type GetCustomerBillingReadinessResponse struct {
	CustomerID                                  string                            `json:"customer_id"`
	ProcessingMonthID                           string                            `json:"processing_month_id"`
	StatusDisplayName                           string                            `json:"status_display_name"`
	IsReleasedForBilling                        bool                              `json:"is_released_for_billing"`
	IsAutomaticallyComplete                     bool                              `json:"is_automatically_complete"`
	IsManuallyReleased                          bool                              `json:"is_manually_released"`
	AutomaticEvaluationUsesCnpjContractingCompanies bool                          `json:"automatic_evaluation_uses_cnpj_contracting_companies"`
	AccountsExpectedForAutomaticRule            int                               `json:"accounts_expected_for_automatic_rule"`
	AccountsWithInvoiceInProcessingMonth        int                               `json:"accounts_with_invoice_in_processing_month"`
	ManualRelease                               *BillingReadinessManualReleaseDto `json:"manual_release"`
}

// --- Phone Lines ---

type ListPhoneLineResponse struct {
	ID                   string     `json:"id"`
	ProviderPlanID       string     `json:"provider_plan_id"`
	ProviderPlanName     string     `json:"provider_plan_name"`
	ProviderAccountID    string     `json:"provider_account_id"`
	ProviderAccountNumber string    `json:"provider_account_number"`
	CostCenterID         *string    `json:"cost_center_id"`
	CostCenterName       *string    `json:"cost_center_name"`
	LastInvoiceID        *string    `json:"last_invoice_id"`
	LastInvoiceNumber    *string    `json:"last_invoice_number"`
	TitularLineID        *string    `json:"titular_line_id"`
	TitularLineNumber    *string    `json:"titular_line_number"`
	Number               string     `json:"number"`
	LineClassification   string     `json:"line_classification"`
	Status               string     `json:"status"`
	TransitionSubStatus  *string    `json:"transition_sub_status"`
	TransitionStartedAt  *time.Time `json:"transition_started_at"`
	ActivationDate       *time.Time `json:"activation_date"`
	CancellationDate     *time.Time `json:"cancellation_date"`
	BaseCost             *float64   `json:"base_cost"`
	CostWithConsumption  *float64   `json:"cost_with_consumption"`
}

type CreateStockPhoneLineInput struct {
	Number                string `json:"number"`
	ProviderID            string `json:"provider_id"`
	ProviderAccountNumber string `json:"provider_account_number"`
	ProviderPlanID        string `json:"provider_plan_id"`
}

type GetPhoneLineServiceResponse struct {
	ID                   string   `json:"id"`
	PhoneLineID          string   `json:"phone_line_id"`
	ProviderPlanServiceID string  `json:"provider_plan_service_id"`
	Name                 string   `json:"name"`
	Code                 string   `json:"code"`
	Recurring            bool     `json:"recurring"`
	Price                *float64 `json:"price"`
	Active               bool     `json:"active"`
}

type GetChildPhoneLineResponse struct {
	ID                   string                      `json:"id"`
	Number               string                      `json:"number"`
	LineClassification   string                      `json:"line_classification"`
	Status               string                      `json:"status"`
	ProviderPlanID       string                      `json:"provider_plan_id"`
	ProviderPlanName     string                      `json:"provider_plan_name"`
	Plan                 *GetProviderPlanResponse    `json:"plan"`
	Services             []GetPhoneLineServiceResponse `json:"services"`
}

type GetPhoneLineResponse struct {
	ListPhoneLineResponse
	Children []GetChildPhoneLineResponse `json:"children"`
	Services []GetPhoneLineServiceResponse `json:"services"`
}

type PhoneLineCustomerLinkResponse struct {
	PhoneLineID      string     `json:"phone_line_id"`
	CustomerID       string     `json:"customer_id"`
	CustomerName     string     `json:"customer_name"`
	CustomerDocument *string    `json:"customer_document"`
	StartDate        time.Time  `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	IsActive         bool       `json:"is_active"`
}

type AssignPhoneLineCustomerInput struct {
	CustomerID string     `json:"customer_id"`
	StartDate  *time.Time `json:"start_date"`
}

type TransferPhoneLineCustomerInput struct {
	CustomerID   string     `json:"customer_id"`
	TransferDate *time.Time `json:"transfer_date"`
}

type UnassignPhoneLineCustomerInput struct {
	EndDate *time.Time `json:"end_date"`
}

// --- Billing Cycles ---

type CreateBillingCycleInput struct {
	ProviderID string    `json:"provider_id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	StartDate  DateInput `json:"start_date"`
	EndDate    DateInput `json:"end_date"`
}

type UpdateBillingCycleInput = CreateBillingCycleInput

type ListBillingCycleResponse struct {
	ID         string     `json:"id"`
	ProviderID string     `json:"provider_id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    time.Time  `json:"end_date"`
	Status     string     `json:"status"`
	ClosedAt   *time.Time `json:"closed_at"`
	ClosedBy   *string    `json:"closed_by"`
}

type GetBillingCycleResponse = ListBillingCycleResponse
type CreateBillingCycleResponse = ListBillingCycleResponse

// --- Processing Months ---

type CreateProcessingMonthInput struct {
	ProviderID  string `json:"provider_id"`
	Year        int    `json:"year"`
	Month       int    `json:"month"`
	DisplayName string `json:"display_name"`
}

type CloseProcessingMonthContingencyInput struct {
	Justification string `json:"justification"`
}

type ListProcessingMonthResponse struct {
	ID                   string     `json:"id"`
	ProviderID           string     `json:"provider_id"`
	Year                 int        `json:"year"`
	Month                int        `json:"month"`
	DisplayName          string     `json:"display_name"`
	Status               string     `json:"status"`
	ClosedAt             *time.Time `json:"closed_at"`
	ClosedBy             *string    `json:"closed_by"`
	ClosedInContingency  bool       `json:"closed_in_contingency"`
}

type GetProcessingMonthResponse struct {
	ListProcessingMonthResponse
	ContingencyJustification *string `json:"contingency_justification"`
}

// --- Provider Invoices ---

type ProviderInvoiceImportRequestInput struct {
	ProviderID        string  `json:"provider_id"`
	ProcessingMonthID string  `json:"processing_month_id"`
	StorageBucket     string  `json:"storage_bucket"`
	StorageObjectKey  string  `json:"storage_object_key"`
	OriginalFileName  *string `json:"original_file_name"`
}

type RequestProviderInvoiceImportResponse struct {
	ID                string     `json:"id"`
	ProcessingMonthID string     `json:"processing_month_id"`
	Status            string     `json:"status"`
	Error             *string    `json:"error"`
	CompletedAt       *time.Time `json:"completed_at"`
}

type ListProviderInvoiceResponse struct {
	ID                      string    `json:"id"`
	ProviderAccountID       string    `json:"provider_account_id"`
	ProviderAccountNumber   string    `json:"provider_account_number"`
	ContractingCompanyID    string    `json:"contracting_company_id"`
	ContractingCompanyName  string    `json:"contracting_company_name"`
	ProviderID              string    `json:"provider_id"`
	ProviderName            string    `json:"provider_name"`
	BillingCycleID          string    `json:"billing_cycle_id"`
	BillingCycleName        string    `json:"billing_cycle_name"`
	ProcessingMonthID       *string   `json:"processing_month_id"`
	CostCenterID            *string   `json:"cost_center_id"`
	ParentInvoiceID         *string   `json:"parent_invoice_id"`
	IssueDate               time.Time `json:"issue_date"`
	DueDate                 time.Time `json:"due_date"`
	TotalAmount             float64   `json:"total_amount"`
	Status                  string    `json:"status"`
	SubtotalServices        float64   `json:"subtotal_services"`
	SubtotalUsage           float64   `json:"subtotal_usage"`
	SubtotalTaxes           float64   `json:"subtotal_taxes"`
	SubtotalDiscounts       float64   `json:"subtotal_discounts"`
	SubtotalInstallments    float64   `json:"subtotal_installments"`
	AccountPayableID        *string   `json:"account_payable_id"`
	AccountPayableStatus    *string   `json:"account_payable_status"`
}

type GetProviderInvoiceItemResponse struct {
	ID             string                           `json:"id"`
	InvoiceID      string                           `json:"invoice_id"`
	ParentID       *string                          `json:"parent_id"`
	Description    string                           `json:"description"`
	Quantity       float64                          `json:"quantity"`
	TotalPrice     float64                          `json:"total_price"`
	ItemType       string                           `json:"item_type"`
	QuotaAmount    *float64                         `json:"quota_amount"`
	ConsumedAmount *float64                         `json:"consumed_amount"`
	Unit           *string                          `json:"unit"`
	Children       []GetProviderInvoiceItemResponse `json:"children"`
}

type GetProviderInvoiceServiceResponse struct {
	ID             string   `json:"id"`
	InvoiceID      string   `json:"invoice_id"`
	PlanID         string   `json:"plan_id"`
	PlanName       string   `json:"plan_name"`
	Description    string   `json:"description"`
	Quantity       float64  `json:"quantity"`
	TotalPrice     float64  `json:"total_price"`
	QuotaAmount    *float64 `json:"quota_amount"`
	ConsumedAmount *float64 `json:"consumed_amount"`
	Unit           *string  `json:"unit"`
}

type GetProviderInvoiceQuotaSharingResponse struct {
	ID             string   `json:"id"`
	InvoiceID      string   `json:"invoice_id"`
	PhoneLineID    string   `json:"phone_line_id"`
	Description    string   `json:"description"`
	ConsumedAmount *float64 `json:"consumed_amount"`
}

type GetProviderPhoneLineResponse struct {
	ID                    string  `json:"id"`
	ProviderPlanID        string  `json:"provider_plan_id"`
	ProviderPlanName      string  `json:"provider_plan_name"`
	ProviderAccountID     string  `json:"provider_account_id"`
	ProviderAccountNumber string  `json:"provider_account_number"`
	CostCenterID          *string `json:"cost_center_id"`
	CostCenterName        *string `json:"cost_center_name"`
	LastInvoiceID         *string `json:"last_invoice_id"`
	LastInvoiceNumber     *string `json:"last_invoice_number"`
	TitularLineID         *string `json:"titular_line_id"`
	TitularLineNumber     *string `json:"titular_line_number"`
	Number                string  `json:"number"`
	LineClassification    string  `json:"line_classification"`
	Status                string  `json:"status"`
	TransitionSubStatus   *string `json:"transition_sub_status"`
}

type GetProviderInvoiceResponse struct {
	ListProviderInvoiceResponse
	Number                   string                                  `json:"number"`
	ProcessingMonthName      *string                                 `json:"processing_month_name"`
	CostCenterName           *string                                 `json:"cost_center_name"`
	PhoneLines               []GetProviderPhoneLineResponse          `json:"phone_lines"`
	ProviderInvoiceItems     []GetProviderInvoiceItemResponse        `json:"provider_invoice_items"`
	ProviderInvoiceServices  []GetProviderInvoiceServiceResponse     `json:"provider_invoice_services"`
	ProviderInvoiceQuotaSharing []GetProviderInvoiceQuotaSharingResponse `json:"provider_invoice_quota_sharing"`
}

// --- Cost Centers ---

type ListCostCenterResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// --- Stats ---

type DashboardStatsResponse struct {
	BillingCyclesCount    int32 `json:"billing_cycles_count"`
	CustomersCount        int32 `json:"customers_count"`
	ProvidersCount        int32 `json:"providers_count"`
	ProviderInvoicesCount int32 `json:"provider_invoices_count"`
	PhoneLinesCount       int32 `json:"phone_lines_count"`
}

// --- Partner ---

type PartnerDashboardStatsResponse struct {
	CustomersCount              int32   `json:"customers_count"`
	PhoneLinesCount             int32   `json:"phone_lines_count"`
	PendingOperationRequests    int32   `json:"pending_operation_requests_count"`
	TotalBaseCost               float64 `json:"total_base_cost"`
	TotalCostWithConsumption    float64 `json:"total_cost_with_consumption"`
}

type CreatePhoneLineOperationRequestInput struct {
	PhoneLineID   string  `json:"phone_line_id"`
	CustomerID    string  `json:"customer_id"`
	OperationType string  `json:"operation_type"`
	Justification *string `json:"justification"`
}

type ReviewPhoneLineOperationRequestInput struct {
	Status     string  `json:"status"`
	AdminNotes *string `json:"admin_notes"`
}

type PhoneLineOperationRequestResponse struct {
	ID                string     `json:"id"`
	PhoneLineID       string     `json:"phone_line_id"`
	PhoneLineNumber   string     `json:"phone_line_number"`
	CustomerID        string     `json:"customer_id"`
	CustomerName      string     `json:"customer_name"`
	OperationType     string     `json:"operation_type"`
	Status            string     `json:"status"`
	Justification     *string    `json:"justification"`
	AdminNotes        *string    `json:"admin_notes"`
	RequestedByUserID string     `json:"requested_by_user_id"`
	ReviewedByUserID  *string    `json:"reviewed_by_user_id"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

type PartnerPhoneLineResponse struct {
	ID                   string   `json:"id"`
	Number               string   `json:"number"`
	Status               string   `json:"status"`
	TransitionSubStatus  *string  `json:"transition_sub_status"`
	CustomerID           *string  `json:"customer_id"`
	CustomerName         *string  `json:"customer_name"`
	ProviderPlanName     string   `json:"provider_plan_name"`
	BaseCost             *float64 `json:"base_cost"`
	CostWithConsumption  *float64 `json:"cost_with_consumption"`
}

// --- Pre-signed URLs ---

// --- Financial ---

type FinancialSummaryResponse struct {
	TotalPayableOpen       float64 `json:"total_payable_open"`
	TotalReceivableOpen    float64 `json:"total_receivable_open"`
	TotalPartnerCommission float64 `json:"total_partner_commission_accrued"`
	PayableOverdueCount    int32   `json:"payable_overdue_count"`
	ReceivableOverdueCount int32   `json:"receivable_overdue_count"`
	// Faturamento operadora (refaturamento)
	ProviderInvoicesCount              int32   `json:"provider_invoices_count"`
	ProviderInvoicesTotalAmount        float64 `json:"provider_invoices_total_amount"`
	ProviderInvoicesWithoutPayableCount int32  `json:"provider_invoices_without_payable_count"`
	OpenProcessingMonthsCount          int32   `json:"open_processing_months_count"`
	BillingDocumentsDraftCount         int32   `json:"billing_documents_draft_count"`
	BillingDocumentsReadyCount         int32   `json:"billing_documents_ready_count"`
	BillingDocumentsSentCount          int32   `json:"billing_documents_sent_count"`
}

type ListAccountPayableResponse struct {
	ID                       string     `json:"id"`
	Description              string     `json:"description"`
	VendorName               string     `json:"vendor_name"`
	ProviderInvoiceID        *string    `json:"provider_invoice_id"`
	PartnerSalespersonUserID *string    `json:"partner_salesperson_user_id"`
	IssueDate                time.Time  `json:"issue_date"`
	DueDate                  time.Time  `json:"due_date"`
	Amount                   float64    `json:"amount"`
	PaidAmount               float64    `json:"paid_amount"`
	Balance                  float64    `json:"balance"`
	Status                   string     `json:"status"`
	Notes                    *string    `json:"notes"`
	CreatedAt                time.Time  `json:"created_at"`
}

type CreateAccountPayableInput struct {
	Description              string  `json:"description"`
	VendorName               string  `json:"vendor_name"`
	ProviderInvoiceID        *string `json:"provider_invoice_id"`
	PartnerSalespersonUserID *string `json:"partner_salesperson_user_id"`
	IssueDate                string  `json:"issue_date"`
	DueDate                  string  `json:"due_date"`
	Amount                   float64 `json:"amount"`
	Notes                    *string `json:"notes"`
}

type UpdateAccountPayableInput struct {
	Description string  `json:"description"`
	VendorName  string  `json:"vendor_name"`
	DueDate     string  `json:"due_date"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	Notes       *string `json:"notes"`
}

type ListAccountReceivableResponse struct {
	ID                string    `json:"id"`
	CustomerID        string    `json:"customer_id"`
	CustomerName      string    `json:"customer_name"`
	Description       string    `json:"description"`
	ProcessingMonthID *string   `json:"processing_month_id"`
	IssueDate         time.Time `json:"issue_date"`
	DueDate           time.Time `json:"due_date"`
	Amount            float64   `json:"amount"`
	ReceivedAmount    float64   `json:"received_amount"`
	Balance           float64   `json:"balance"`
	Status            string    `json:"status"`
	Notes             *string   `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
}

type CreateAccountReceivableInput struct {
	CustomerID        string  `json:"customer_id"`
	Description       string  `json:"description"`
	ProcessingMonthID *string `json:"processing_month_id"`
	IssueDate         string  `json:"issue_date"`
	DueDate           string  `json:"due_date"`
	Amount            float64 `json:"amount"`
	Notes             *string `json:"notes"`
}

type UpdateAccountReceivableInput struct {
	Description string  `json:"description"`
	DueDate     string  `json:"due_date"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	Notes       *string `json:"notes"`
}

type RegisterFinancialPaymentInput struct {
	Amount      float64 `json:"amount"`
	PaymentDate string  `json:"payment_date"`
	Reference   *string `json:"reference"`
	Notes       *string `json:"notes"`
}

type ListPartnerSaleResponse struct {
	ID                string    `json:"id"`
	SalespersonUserID string    `json:"salesperson_user_id"`
	CustomerID        string    `json:"customer_id"`
	CustomerName      string    `json:"customer_name"`
	PhoneLineID       string    `json:"phone_line_id"`
	PhoneLineNumber   string    `json:"phone_line_number"`
	ReferenceMonth    time.Time `json:"reference_month"`
	GrossAmount       float64   `json:"gross_amount"`
	CommissionPercent float64   `json:"commission_percent"`
	CommissionAmount  float64   `json:"commission_amount"`
	Status            string    `json:"status"`
	AccountPayableID  *string   `json:"account_payable_id"`
	CreatedAt         time.Time `json:"created_at"`
}

type SyncPartnerSalesInput struct {
	ReferenceMonth string `json:"reference_month"`
}

type SyncPartnerSalesResponse struct {
	InsertedCount int `json:"inserted_count"`
}

type UpdatePartnerSaleStatusInput struct {
	Status string `json:"status"`
}

type PartnerCommissionSettingsResponse struct {
	DefaultCommissionPercent float64   `json:"default_commission_percent"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type UpdatePartnerCommissionSettingsInput struct {
	DefaultCommissionPercent float64 `json:"default_commission_percent"`
}

type PartnerFinancialSummaryResponse struct {
	TotalGrossSales          float64 `json:"total_gross_sales"`
	TotalCommissionAccrued   float64 `json:"total_commission_accrued"`
	TotalCommissionApproved  float64 `json:"total_commission_approved"`
	TotalCommissionPaid      float64 `json:"total_commission_paid"`
	TotalReceivableFromSales float64 `json:"total_receivable_from_sales"`
	PendingSalesCount        int32   `json:"pending_sales_count"`
}

type CreateAccountPayableFromInvoiceResponse struct {
	ID string `json:"id"`
}

// --- Commercial sales & contracts ---

type ListContractTemplateResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetContractTemplateResponse struct {
	ListContractTemplateResponse
	BodyTemplate string `json:"body_template"`
}

type CreateContractTemplateInput struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	BodyTemplate string `json:"body_template"`
	Active       *bool  `json:"active"`
}

type UpdateContractTemplateInput struct {
	Name         *string `json:"name"`
	Code         *string `json:"code"`
	BodyTemplate *string `json:"body_template"`
	Active       *bool   `json:"active"`
}

type SaleLineItemResponse struct {
	ID           string  `json:"id"`
	LineItemType string  `json:"line_item_type"`
	Description  string  `json:"description"`
	Quantity     float64 `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	TotalPrice   float64 `json:"total_price"`
	PhoneLineID  *string `json:"phone_line_id"`
	DeviceSku    *string `json:"device_sku"`
	SortOrder    int32   `json:"sort_order"`
}

type GeneratedContractResponse struct {
	ID                 string     `json:"id"`
	ContractTemplateID string     `json:"contract_template_id"`
	Status             string     `json:"status"`
	RenderedHTML       *string    `json:"rendered_html"`
	GeneratedAt        *time.Time `json:"generated_at"`
}

type ListSaleResponse struct {
	ID                   string     `json:"id"`
	SaleNumber           string     `json:"sale_number"`
	CustomerID           string     `json:"customer_id"`
	CustomerName         string     `json:"customer_name"`
	SalespersonUserID    string     `json:"salesperson_user_id"`
	ContractTemplateID   *string    `json:"contract_template_id"`
	ContractTemplateName *string    `json:"contract_template_name"`
	Status               string     `json:"status"`
	SoldAt               *time.Time `json:"sold_at"`
	TotalAmount          float64    `json:"total_amount"`
	Notes                *string    `json:"notes"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type GetSaleResponse struct {
	ListSaleResponse
	Items    []SaleLineItemResponse     `json:"items"`
	Contract *GeneratedContractResponse `json:"contract"`
}

type CreateSaleInput struct {
	CustomerID         string                    `json:"customer_id"`
	ContractTemplateID *string                   `json:"contract_template_id"`
	Notes              *string                   `json:"notes"`
	Items              []CreateSaleLineItemInput `json:"items"`
}

type CreateSaleLineItemInput struct {
	LineItemType string  `json:"line_item_type"`
	Description  string  `json:"description"`
	Quantity     float64 `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	PhoneLineID  *string `json:"phone_line_id"`
	DeviceSku    *string `json:"device_sku"`
}

type AddSaleLineItemInput struct {
	LineItemType string  `json:"line_item_type"`
	Description  string  `json:"description"`
	Quantity     float64 `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	PhoneLineID  *string `json:"phone_line_id"`
	DeviceSku    *string `json:"device_sku"`
}

type UpdateSaleInput struct {
	CustomerID         *string `json:"customer_id"`
	ContractTemplateID *string `json:"contract_template_id"`
	Notes              *string `json:"notes"`
}

// --- Organization users (Keycloak) ---

type ListOrganizationUserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	Profile   string `json:"profile"`
	Enabled   bool   `json:"enabled"`
}

type CreateOrganizationUserInput struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	Profile   string `json:"profile"`
}

type UpdateOrganizationUserInput struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
	Profile   *string `json:"profile"`
	Enabled   *bool   `json:"enabled"`
	Password  *string `json:"password"`
}

// --- Customer billing & email ---

type ListInvoiceEmailTemplateResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Code            string    `json:"code"`
	Kind            string    `json:"kind"`
	SubjectTemplate string    `json:"subject_template"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type GetInvoiceEmailTemplateResponse struct {
	ListInvoiceEmailTemplateResponse
	BodyTemplateHtml string `json:"body_template_html"`
}

type CreateInvoiceEmailTemplateInput struct {
	Name             string `json:"name"`
	Code             string `json:"code"`
	Kind             string `json:"kind"`
	SubjectTemplate  string `json:"subject_template"`
	BodyTemplateHtml string `json:"body_template_html"`
	Active           *bool  `json:"active"`
}

type UpdateInvoiceEmailTemplateInput struct {
	Name             string `json:"name"`
	SubjectTemplate  string `json:"subject_template"`
	BodyTemplateHtml string `json:"body_template_html"`
	Active           *bool  `json:"active"`
}

type ListCustomerBillingDocumentResponse struct {
	ID                   string     `json:"id"`
	CustomerID           string     `json:"customer_id"`
	CustomerName         string     `json:"customer_name"`
	AccountsReceivableID *string    `json:"accounts_receivable_id"`
	ProcessingMonthID    *string    `json:"processing_month_id"`
	InvoiceNumber        string     `json:"invoice_number"`
	IssueDate            time.Time  `json:"issue_date"`
	DueDate              time.Time  `json:"due_date"`
	Amount               float64    `json:"amount"`
	Status               string     `json:"status"`
	RecipientEmail       string     `json:"recipient_email"`
	EmailSubject         string     `json:"email_subject"`
	SendCount            int32      `json:"send_count"`
	SentAt               *time.Time `json:"sent_at"`
	LastSentAt           *time.Time `json:"last_sent_at"`
	CreatedAt            time.Time  `json:"created_at"`
}

type GetCustomerBillingDocumentResponse struct {
	ListCustomerBillingDocumentResponse
	EmailBodyHtml string    `json:"email_body_html"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UpdateCustomerBillingDocumentInput struct {
	RecipientEmail string `json:"recipient_email"`
	EmailSubject   string `json:"email_subject"`
	EmailBodyHtml  string `json:"email_body_html"`
	Status         string `json:"status"`
}

type CreateCustomerBillingDocumentFromReceivableResponse struct {
	ID string `json:"id"`
}

type SendCustomerBillingDocumentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CustomerBillingSendLogResponse struct {
	ID             string    `json:"id"`
	RecipientEmail string    `json:"recipient_email"`
	Subject        string    `json:"subject"`
	Success        bool      `json:"success"`
	ErrorMessage   *string   `json:"error_message"`
	SentByUserID   string    `json:"sent_by_user_id"`
	SentAt         time.Time `json:"sent_at"`
}

type OverdueReceivableResponse struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	BillingEmail string    `json:"billing_email"`
	Description  string    `json:"description"`
	DueDate      time.Time `json:"due_date"`
	Balance      float64   `json:"balance"`
	RemindersSent int32    `json:"reminders_sent"`
}

type SendCollectionReminderInput struct {
	AccountsReceivableID string `json:"accounts_receivable_id"`
	ReminderLevel        int    `json:"reminder_level"`
	TemplateCode         string `json:"template_code"`
}

type SendCollectionReminderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CustomerBillingDocumentRow struct {
	ID                   string
	OrganizationID       string
	CustomerID           string
	AccountsReceivableID *string
	ProcessingMonthID    *string
	InvoiceNumber        string
	IssueDate            time.Time
	DueDate              time.Time
	Amount               float64
	Status               string
	RecipientEmail       string
	EmailSubject         string
	EmailBodyHTML        string
	CreatedAt            time.Time
}

// --- Pre-signed URLs (continued) ---

type CreatePresignedUploadURLInput struct {
	BucketName        string  `json:"bucket_name"`
	ObjectKey         string  `json:"object_key"`
	ContentType       *string `json:"content_type"`
	ExpiresInSeconds  *int    `json:"expires_in_seconds"`
}

type CreatePresignedDownloadURLInput struct {
	BucketName       string `json:"bucket_name"`
	ObjectKey        string `json:"object_key"`
	ExpiresInSeconds *int   `json:"expires_in_seconds"`
}
