package notifications

import "strings"

type Notification struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Param   *string `json:"param"`
}

func N(code, message string) Notification {
	return Notification{Code: code, Message: message, Param: nil}
}

func NP(code, message, param string) Notification {
	return Notification{Code: code, Message: message, Param: &param}
}

// Shared
func SharedUnexpectedError(detail string) Notification {
	msg := "Ocorreu um erro inesperado. Tente novamente."
	d := strings.ToLower(strings.TrimSpace(detail))
	switch {
	case strings.Contains(d, "column") && strings.Contains(d, "does not exist"):
		msg = "Schema do banco desatualizado (coluna ausente). Reinicie a API para aplicar as migrações."
	case strings.Contains(d, "relation") && strings.Contains(d, "does not exist"):
		msg = "Schema do banco desatualizado (tabela ausente). Aplique as migrações."
	case strings.Contains(d, "violates foreign key"):
		msg = "Não foi possível gravar: referência inválida no banco."
	case strings.Contains(d, "duplicate key"):
		msg = "Registro duplicado. Atualize a página e tente novamente."
	}
	return N("UNEXPECTED_ERROR", msg)
}

var (
	SharedResourceNotFound     = N("RESOURCE_NOT_FOUND", "The requested resource was not found")
	SharedDomainViolation      = N("DOMAIN_VIOLATION", "An business rule violation has occurred.")
	SharedOrganizationRequired = N("ORGANIZATION_ID_REQUIRED", "Organization ID is required.")
)

// Providers
var (
	ProviderNameRequired       = N("PROVIDER_NAME_REQUIRED", "Provider name is required.")
	ProviderNameMaxLength      = N("PROVIDER_NAME_MAX_LENGTH", "Provider name must not exceed 100 characters.")
	ProviderSlugRequired       = N("PROVIDER_SLUG_REQUIRED", "Provider slug is required.")
	ProviderSlugMaxLength      = N("PROVIDER_SLUG_MAX_LENGTH", "Provider slug must not exceed 50 characters.")
	ProviderSlugDuplicated     = N("PROVIDER_SLUG_DUPLICATED", "An provider with this slug already exists.")
	ProviderNotFound           = N("PROVIDER_NOT_FOUND", "Provider was not found.")
	ProviderPlanCodeRequired   = N("PROVIDER_PLAN_CODE_REQUIRED", "Plan code is required.")
	ProviderPlanCodeMaxLength  = N("PROVIDER_PLAN_CODE_MAX_LENGTH", "Plan code must not exceed 64 characters.")
	ProviderPlanNameRequired   = N("PROVIDER_PLAN_NAME_REQUIRED", "Plan name is required.")
	ProviderPlanNameMaxLength  = N("PROVIDER_PLAN_NAME_MAX_LENGTH", "Plan name must not exceed 256 characters.")
	ProviderPlanCodeDuplicated = N("PROVIDER_PLAN_CODE_DUPLICATED", "A plan with this code already exists for the operator.")
	ProviderPlanNotFound       = N("PROVIDER_PLAN_NOT_FOUND", "Plan was not found.")
)

// Customers
var (
	CustomerNameRequired               = N("CUSTOMER_NAME_REQUIRED", "Customer name is required.")
	CustomerNameMaxLength              = N("CUSTOMER_NAME_MAX_LENGTH", "Customer name must not exceed 256 characters.")
	CustomerDocumentRequired           = N("CUSTOMER_DOCUMENT_REQUIRED", "Customer document is required.")
	CustomerDocumentMaxLength          = N("CUSTOMER_DOCUMENT_MAX_LENGTH", "Customer document must not exceed 20 characters.")
	CustomerLegalNameRequiredForPJ     = N("CUSTOMER_LEGAL_NAME_REQUIRED_FOR_PJ", "Legal name is required for PJ customers.")
	CustomerDocumentDuplicated         = N("CUSTOMER_DOCUMENT_DUPLICATED", "A customer with this document already exists.")
	CustomerNotFound                   = N("CUSTOMER_NOT_FOUND", "Customer was not found.")
	CustomerBillingReadinessNotFound   = N("CUSTOMER_BILLING_READINESS_CONTEXT_NOT_FOUND", "Billing readiness context was not found.")
	CustomerProcessingMonthMismatch    = N("CUSTOMER_PROCESSING_MONTH_PROVIDER_MISMATCH", "Customer and processing month provider mismatch.")
	CustomerManualReleaseAlready       = N("CUSTOMER_MANUAL_RELEASE_ALREADY_EXISTS", "Manual release already exists for this customer and processing month.")
	CustomerManualReleaseJustMin       = N("CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_MIN_LENGTH", "Justification must be at least 10 characters.")
	CustomerManualReleaseJustMax       = N("CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_MAX_LENGTH", "Justification must not exceed 4000 characters.")
	CustomerAttachmentOriginalRequired = N("CUSTOMER_ATTACHMENT_ORIGINAL_FILE_NAME_REQUIRED", "Original file name is required.")
	CustomerAttachmentBucketRequired   = N("CUSTOMER_ATTACHMENT_STORAGE_BUCKET_REQUIRED", "Storage bucket is required.")
	CustomerAttachmentKeyRequired      = N("CUSTOMER_ATTACHMENT_STORAGE_OBJECT_KEY_REQUIRED", "Storage object key is required.")
)

// PhoneLines
var (
	PhoneLineNotFound                   = N("PHONE_LINE_NOT_FOUND", "Phone line was not found.")
	PhoneLineNumberRequired             = N("PHONE_LINE_NUMBER_REQUIRED", "Phone line number is required.")
	PhoneLineNumberDuplicated           = N("PHONE_LINE_NUMBER_DUPLICATED", "A phone line with this number already exists.")
	PhoneLineProviderAccountNotFound    = N("PHONE_LINE_PROVIDER_ACCOUNT_NOT_FOUND", "Provider account was not found for this operator.")
	PhoneLineProviderPlanInvalid        = N("PHONE_LINE_PROVIDER_PLAN_INVALID", "Selected plan does not belong to the operator.")
	PhoneLineActiveCustomerLinkNotFound = N("PHONE_LINE_ACTIVE_CUSTOMER_LINK_NOT_FOUND", "No active customer link was found for this phone line.")
	PhoneLineCustomerTransferSame       = N("PHONE_LINE_CUSTOMER_TRANSFER_SAME_CUSTOMER", "Phone line transfer requires a different target customer.")
	PhoneLineOperationPendingExists     = N("PHONE_LINE_OPERATION_PENDING_EXISTS", "There is already a pending operation request for this phone line.")
	PhoneLineOperationNotFound          = N("PHONE_LINE_OPERATION_NOT_FOUND", "Phone line operation request was not found.")
	PhoneLineOperationAlreadyReviewed   = N("PHONE_LINE_OPERATION_ALREADY_REVIEWED", "This operation request has already been reviewed.")
	PartnerCustomerAccessDenied         = N("PARTNER_CUSTOMER_ACCESS_DENIED", "You do not have access to this customer.")
	PartnerPhoneLineAccessDenied        = N("PARTNER_PHONE_LINE_ACCESS_DENIED", "You do not have access to this phone line.")
	PhoneLineClassificationInvalid      = N("PHONE_LINE_CLASSIFICATION_INVALID", "Classificação inválida. Use normal, titular ou dependent.")
	PhoneLineTitularRequired            = N("PHONE_LINE_TITULAR_REQUIRED", "Linha dependente deve estar vinculada a um titular.")
	PhoneLineTitularSelf                = N("PHONE_LINE_TITULAR_SELF", "Uma linha não pode ser titular de si mesma.")
	PhoneLineTitularNotFound            = N("PHONE_LINE_TITULAR_NOT_FOUND", "Linha titular não encontrada.")
	PhoneLineServiceRequiredToLink      = N("PHONE_LINE_SERVICE_REQUIRED_TO_LINK", "Não é possível vincular a linha a um cliente sem ao menos um serviço ativo na composição.")
	PhoneLineServiceTypeDuplicated      = N("PHONE_LINE_SERVICE_TYPE_DUPLICATED", "Já existe serviço do mesmo tipo ativo nesta linha no período informado.")
	PhoneLineServiceNotFound            = N("PHONE_LINE_SERVICE_NOT_FOUND", "Serviço da linha não encontrado.")
	PhoneLineTransitionSubStatusInvalid = N("PHONE_LINE_TRANSITION_SUBSTATUS_INVALID", "Substatus de transição inválido.")
)

// Device stock
var (
	DeviceStockBrandRequired  = N("DEVICE_STOCK_BRAND_REQUIRED", "Device brand is required.")
	DeviceStockBrandMaxLength = N("DEVICE_STOCK_BRAND_MAX_LENGTH", "Device brand must not exceed 128 characters.")
	DeviceStockModelRequired  = N("DEVICE_STOCK_MODEL_REQUIRED", "Device model is required.")
	DeviceStockModelMaxLength = N("DEVICE_STOCK_MODEL_MAX_LENGTH", "Device model must not exceed 256 characters.")
	DeviceStockSkuMaxLength   = N("DEVICE_STOCK_SKU_MAX_LENGTH", "SKU must not exceed 64 characters.")
	DeviceStockSkuDuplicated  = N("DEVICE_STOCK_SKU_DUPLICATED", "A device with this SKU already exists.")
	DeviceStockImeiDuplicated = N("DEVICE_STOCK_IMEI_DUPLICATED", "A device with this IMEI already exists.")
	DeviceStockImeiInvalid    = N("DEVICE_STOCK_IMEI_INVALID", "IMEI must have 15 digits.")
	DeviceStockNotFound       = N("DEVICE_STOCK_NOT_FOUND", "Device stock item was not found.")
	DeviceStockStatusInvalid  = N("DEVICE_STOCK_STATUS_INVALID", "Invalid device stock status.")
)

// Customer devices
var (
	CustomerDeviceDescriptionRequired  = N("CUSTOMER_DEVICE_DESCRIPTION_REQUIRED", "Device description is required.")
	CustomerDeviceMonthlyAmountInvalid = N("CUSTOMER_DEVICE_MONTHLY_AMOUNT_INVALID", "Monthly amount must be zero or greater.")
	CustomerDeviceNotFound             = N("CUSTOMER_DEVICE_NOT_FOUND", "Customer device link was not found.")
	CustomerDeviceAlreadyLinked        = N("CUSTOMER_DEVICE_ALREADY_LINKED", "This device is already linked to an active customer.")
)

// Financial
var (
	FinancialDateRequired             = N("FINANCIAL_DATE_REQUIRED", "Date is required.")
	FinancialDateInvalid              = N("FINANCIAL_DATE_INVALID", "Date must be in YYYY-MM-DD format.")
	FinancialDescriptionRequired      = N("FINANCIAL_DESCRIPTION_REQUIRED", "Description is required.")
	FinancialAmountInvalid            = N("FINANCIAL_AMOUNT_INVALID", "Amount must be greater than zero.")
	FinancialPayableNotFound          = N("FINANCIAL_PAYABLE_NOT_FOUND", "Account payable was not found.")
	FinancialReceivableNotFound       = N("FINANCIAL_RECEIVABLE_NOT_FOUND", "Account receivable was not found.")
	FinancialPartnerSaleNotFound      = N("FINANCIAL_PARTNER_SALE_NOT_FOUND", "Partner sale record was not found.")
	FinancialPartnerSaleStatusInvalid = N("FINANCIAL_PARTNER_SALE_STATUS_INVALID", "Invalid partner sale status.")
	FinancialCommissionPercentInvalid = N("FINANCIAL_COMMISSION_PERCENT_INVALID", "Commission percent must be between 0 and 100.")
	FinancialPayableFromInvoiceExists = N("FINANCIAL_PAYABLE_FROM_INVOICE_EXISTS", "An account payable already exists for this provider invoice.")
)

// Sales & contracts
var (
	SaleNotFound                   = N("SALE_NOT_FOUND", "Sale was not found.")
	SaleStatusInvalid              = N("SALE_STATUS_INVALID", "Sale status does not allow this operation.")
	SaleItemsRequired              = N("SALE_ITEMS_REQUIRED", "At least one line item is required to confirm the sale.")
	SaleLineItemInvalid            = N("SALE_LINE_ITEM_INVALID", "Invalid sale line item.")
	ContractTemplateNotFound       = N("CONTRACT_TEMPLATE_NOT_FOUND", "Contract template was not found.")
	ContractTemplateNameRequired   = N("CONTRACT_TEMPLATE_NAME_REQUIRED", "Contract template name is required.")
	ContractTemplateCodeRequired   = N("CONTRACT_TEMPLATE_CODE_REQUIRED", "Contract template code is required.")
	ContractTemplateBodyRequired   = N("CONTRACT_TEMPLATE_BODY_REQUIRED", "Contract template body is required.")
	ContractTemplateCodeDuplicated = N("CONTRACT_TEMPLATE_CODE_DUPLICATED", "A contract template with this code already exists.")
)

// BillingCycles
var (
	BillingCycleCodeRequired  = N("BILLING_CYCLE_CODE_REQUIRED", "Billing cycle code is required.")
	BillingCycleCodeMaxLength = N("BILLING_CYCLE_CODE_MAX_LENGTH", "Billing cycle code must not exceed 20 characters.")
	BillingCycleNameRequired  = N("BILLING_CYCLE_NAME_REQUIRED", "Billing cycle name is required.")
	BillingCycleNameMaxLength = N("BILLING_CYCLE_NAME_MAX_LENGTH", "Billing cycle name must not exceed 100 characters.")
	BillingCycleNotFound      = N("BILLING_CYCLE_NOT_FOUND", "Billing cycle was not found.")
	BillingCycleConsolidated  = N("BILLING_CYCLE_CONSOLIDATED", "Billing cycle is consolidated and cannot be changed.")
)

// ProcessingMonths
var (
	ProcessingMonthNotFound             = N("PROCESSING_MONTH_NOT_FOUND", "Processing month was not found.")
	ProcessingMonthProviderMismatch     = N("PROCESSING_MONTH_PROVIDER_MISMATCH", "Processing month provider mismatch.")
	ProcessingMonthDuplicate            = N("PROCESSING_MONTH_DUPLICATE", "Processing month already exists for this provider, year and month.")
	ProcessingMonthAlreadyClosed        = N("PROCESSING_MONTH_ALREADY_CLOSED", "Processing month is already closed.")
	ProcessingMonthNotOpen              = N("PROCESSING_MONTH_NOT_OPEN", "Processing month is not open.")
	ProcessingMonthYearInvalid          = N("PROCESSING_MONTH_YEAR_INVALID", "Year must be between 2000 and 2100.")
	ProcessingMonthMonthInvalid         = N("PROCESSING_MONTH_MONTH_INVALID", "Month must be between 1 and 12.")
	ProcessingMonthDisplayNameRequired  = N("PROCESSING_MONTH_DISPLAY_NAME_REQUIRED", "Display name is required.")
	ProcessingMonthDisplayNameMaxLength = N("PROCESSING_MONTH_DISPLAY_NAME_MAX_LENGTH", "Display name must not exceed 128 characters.")
	ProcessingMonthRetroactiveBlocked   = N("PROCESSING_MONTH_RETROACTIVE_CHANGE_BLOCKED", "Retroactive change blocked by closed processing month.")
	ProcessingMonthContingencyJustMin   = N("PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_MIN_LENGTH", "Justification must be at least 10 characters.")
	ProcessingMonthContingencyJustMax   = N("PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_MAX_LENGTH", "Justification must not exceed 4000 characters.")
)

// Invoices
var (
	InvoiceNotFound                      = N("INVOICE_NOT_FOUND", "Invoice was not found.")
	InvoiceDuplicateSameProcessingMonth  = N("INVOICE_DUPLICATE_SAME_PROCESSING_MONTH", "Invoice duplicate for same processing month.")
	InvoiceDuplicateOtherProcessingMonth = N("INVOICE_DUPLICATE_OTHER_PROCESSING_MONTH", "Esta fatura já foi importada em outro mês de processamento.")
	InvoiceDuplicateFileHash             = N("INVOICE_DUPLICATE_FILE_HASH", "Este arquivo já foi importado (SHA-256 idêntico). Use o fluxo explícito de fatura substituta se for uma substituição.")
	InvoiceImportedLineOrphanDestination = N("INVOICE_IMPORTED_LINE_ORPHAN_DESTINATION", "Imported line has incompatible status.")
)

// InvoiceImports
var (
	ImportProviderIDRequired           = N("PROVIDER_ID_REQUIRED", "Provider id is required.")
	ImportProcessingMonthIDRequired    = N("PROCESSING_MONTH_ID_REQUIRED", "Processing month id is required.")
	ImportStorageBucketRequired        = N("STORAGE_BUCKET_REQUIRED", "Storage bucket is required.")
	ImportStorageBucketMaxLength       = N("STORAGE_BUCKET_MAX_LENGTH", "Storage bucket must not exceed 256 characters.")
	ImportStorageObjectKeyRequired     = N("STORAGE_OBJECT_KEY_REQUIRED", "Storage object key is required.")
	ImportStorageObjectKeyMaxLength    = N("STORAGE_OBJECT_KEY_MAX_LENGTH", "Storage object key must not exceed 2048 characters.")
	ImportRequestNotFound              = N("IMPORT_REQUEST_NOT_FOUND", "Import request was not found.")
	ImportRequestNotPending            = N("IMPORT_REQUEST_NOT_PENDING", "Import request is not pending.")
	ImportCustomerDocumentInvalid      = N("CUSTOMER_DOCUMENT_INVALID_FOR_IMPORT", "Customer document invalid for import.")
	ImportCPFRequiresExistingCustomer  = N("CPF_REQUIRES_EXISTING_CUSTOMER_FOR_IMPORT", "CPF requires existing customer for import.")
	ImportContractingCompanyNotFound   = N("CONTRACTING_COMPANY_NOT_FOUND_FOR_FILE", "Contracting company not found for file.")
	CustomerContractingCompanyMismatch = N("CUSTOMER_CONTRACTING_COMPANY_MISMATCH", "Customer contracting company mismatch.")
	ImportPDFNotParsed                 = N("IMPORT_PDF_NOT_PARSED", "PDF ainda não parseado. Importe o arquivo TXT equivalente da operadora.")
	ImportHeaderMissing                = N("IMPORT_HEADER_MISSING", "Registro 010D não encontrado no arquivo.")
	ImportCompetenceMismatch           = N("IMPORT_COMPETENCE_MISMATCH", "Competência da fatura diverge da competência selecionada. Importação bloqueada.")
	ImportCompetenceUndetermined       = N("IMPORT_COMPETENCE_UNDETERMINED", "Não foi possível determinar a competência da fatura com segurança. Importação bloqueada.")
	ImportFileInvalid                  = N("IMPORT_FILE_INVALID", "Arquivo de fatura inválido.")
	ImportFileReadError                = N("IMPORT_FILE_READ_ERROR", "Erro ao ler o arquivo de fatura.")
	ImportFormatUnsupported            = N("IMPORT_FORMAT_UNSUPPORTED", "Formato de arquivo não suportado.")
	ImportDataInconsistent             = N("IMPORT_DATA_INCONSISTENT", "Dados da fatura inconsistentes.")
	ImportProcessingFailed             = N("IMPORT_PROCESSING_FAILED", "Falha no processamento da importação.")
	ImportCompletedSuccess             = N("IMPORT_COMPLETED_SUCCESS", "Importação concluída com sucesso.")
)

// Exceedances & fidelity
var (
	ExceedanceTermRequired      = N("EXCEEDANCE_TERM_REQUIRED", "Informe o termo de excedente.")
	ExceedanceTermNotFound      = N("EXCEEDANCE_TERM_NOT_FOUND", "Termo de excedente não encontrado.")
	ExceedanceChargeTypeInvalid = N("EXCEEDANCE_CHARGE_TYPE_INVALID", "Tipo de cobrança inválido. Use mirrored ou tabulated.")
	LineFidelityNotFound        = N("LINE_FIDELITY_NOT_FOUND", "Fidelidade da linha não encontrada.")
	LineFidelityMonthsInvalid   = N("LINE_FIDELITY_MONTHS_INVALID", "Prazo de fidelidade deve ser maior que zero.")
	LineFidelityRenewalInvalid  = N("LINE_FIDELITY_RENEWAL_INVALID", "Período de renovação deve ser maior que zero quando a renovação automática está ativa.")
)

// ObjectStorage
var (
	ObjectStorageUnavailable = N("OBJECT_STORAGE_UNAVAILABLE", "Object storage is not configured.")
	PresignedExpiresInvalid  = N("PRESIGNED_EXPIRES_IN_SECONDS_INVALID", "Expires in seconds must be between 60 and 604800.")
	ObjectKeyInvalid         = N("OBJECT_KEY_INVALID", "Object key is invalid.")
	TicketAttachmentRequired = N("TICKET_ATTACHMENT_REQUIRED", "Informe o arquivo ou a chave do anexo.")
	TicketAttachmentNotFound = N("TICKET_ATTACHMENT_NOT_FOUND", "Anexo do ticket não encontrado.")
	TicketMessageEmpty       = N("TICKET_MESSAGE_REQUIRED", "Informe a mensagem ou anexe um arquivo.")
	TicketFilenameInvalid    = N("TICKET_ATTACHMENT_FILENAME_INVALID", "Nome de arquivo inválido.")
)

// Billing & email
var (
	BillingEmailTemplateNotFound        = N("BILLING_EMAIL_TEMPLATE_NOT_FOUND", "Modelo de e-mail de fatura não encontrado. Verifique os templates em Configurações.")
	BillingEmailTemplateCodeDuplicated  = N("BILLING_EMAIL_TEMPLATE_CODE_DUPLICATED", "Já existe um template com este código.")
	BillingEmailTemplateFieldsRequired  = N("BILLING_EMAIL_TEMPLATE_FIELDS_REQUIRED", "Nome, assunto e corpo do template são obrigatórios.")
	BillingDocumentNotFound             = N("BILLING_DOCUMENT_NOT_FOUND", "Fatura do cliente não encontrada.")
	BillingDocumentFieldsRequired       = N("BILLING_DOCUMENT_FIELDS_REQUIRED", "Assunto e corpo da fatura são obrigatórios.")
	BillingDocumentCancelled            = N("BILLING_DOCUMENT_CANCELLED", "Faturas canceladas não podem ser enviadas.")
	BillingDocumentAlreadyPaid          = N("BILLING_DOCUMENT_ALREADY_PAID", "Esta fatura já está marcada como paga.")
	BillingDocumentManualPayBlocked     = N("BILLING_DOCUMENT_MANUAL_PAY_BLOCKED", "Não é possível dar baixa manual nesta fatura.")
	SaleAlreadyPaid                     = N("SALE_ALREADY_PAID", "Esta venda já está marcada como paga.")
	SaleManualPayBlocked                = N("SALE_MANUAL_PAY_BLOCKED", "Somente vendas confirmadas (ou em rascunho) podem receber baixa presencial.")
	PaymentMethodInvalid                = N("PAYMENT_METHOD_INVALID", "Forma de pagamento inválida. Use cash, pix_presencial ou card_presencial.")
	BillingCustomerEmailRequired        = N("BILLING_CUSTOMER_EMAIL_REQUIRED", "Informe o e-mail de cobrança do cliente.")
	BillingEmailNotConfigured           = N("BILLING_EMAIL_NOT_CONFIGURED", "SMTP não configurado. Defina SMTP_HOST e variáveis relacionadas.")
	BillingReceivableNotOverdue         = N("BILLING_RECEIVABLE_NOT_OVERDUE", "Somente contas vencidas podem receber lembrete de cobrança.")
	BillingLayoutTemplateNotFound       = N("BILLING_LAYOUT_TEMPLATE_NOT_FOUND", "Modelo de layout de fatura não encontrado. Verifique os templates em Configurações.")
	BillingLayoutTemplateCodeDuplicated = N("BILLING_LAYOUT_TEMPLATE_CODE_DUPLICATED", "Já existe um layout com este código.")
	BillingLayoutTemplateFieldsRequired = N("BILLING_LAYOUT_TEMPLATE_FIELDS_REQUIRED", "Nome e configuração do layout são obrigatórios.")
	BillingDocumentCreateFailed         = N("BILLING_DOCUMENT_CREATE_FAILED", "Não foi possível gravar a fatura. Verifique se as migrações do banco estão atualizadas.")
	BillingGenerateFailed               = N("BILLING_GENERATE_FAILED", "Não foi possível gerar a fatura.")
	SicrediNotConfigured                = N("SICREDI_NOT_CONFIGURED", "Cobrança Sicredi não está configurada. Preencha API Key, usuário/senha e dados do beneficiário em Configurações.")
	SicrediBoletoAlreadyIssued          = N("SICREDI_BOLETO_ALREADY_ISSUED", "Já existe boleto Sicredi emitido para esta fatura.")
	SicrediAPIKeyInvalid                = N("SICREDI_API_KEY_INVALID", "API Key Sicredi inválida. Não use chaves OpenAI (sk-proj). Use a chave do portal Sicredi Parceiro.")
	FinancialAgentNotConfigured         = N("FINANCIAL_AGENT_NOT_CONFIGURED", "Financial agent API key is not configured. Set FINANCIAL_AGENT_API_KEY.")
	FinancialAgentOrgRequired           = N("FINANCIAL_AGENT_ORG_REQUIRED", "Financial agent organization is not configured. Set FINANCIAL_AGENT_ORG_ID.")
	FinancialAgentQueryRequired         = N("FINANCIAL_AGENT_QUERY_REQUIRED", "Informe nome, telefone, CPF/CNPJ, número da linha ou número da fatura.")
	FinancialAgentReceiptRequired       = N("FINANCIAL_AGENT_RECEIPT_REQUIRED", "Informe os dados extraídos do comprovante (valor e data, PIX ou número da fatura).")
	FinancialAgentConfirmDenied         = N("FINANCIAL_AGENT_CONFIRM_DENIED", "Pagamento não pode ser confirmado automaticamente. Encaminhe para revisão humana.")
	FinancialAgentWhatsAppNotConfigured = N("FINANCIAL_AGENT_WHATSAPP_NOT_CONFIGURED", "Informe a URL, a chave e o nome da instância da Evolution API.")
	FinancialAgentDisabled              = N("FINANCIAL_AGENT_DISABLED", "O agente financeiro está desabilitado nesta organização.")
)
