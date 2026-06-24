package notifications

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
func SharedUnexpectedError(msg string) Notification {
	return N("UNEXPECTED_ERROR", msg)
}

var (
	SharedResourceNotFound     = N("RESOURCE_NOT_FOUND", "The requested resource was not found")
	SharedDomainViolation      = N("DOMAIN_VIOLATION", "An business rule violation has occurred.")
	SharedOrganizationRequired = N("ORGANIZATION_ID_REQUIRED", "Organization ID is required.")
)

// Providers
var (
	ProviderNameRequired    = N("PROVIDER_NAME_REQUIRED", "Provider name is required.")
	ProviderNameMaxLength   = N("PROVIDER_NAME_MAX_LENGTH", "Provider name must not exceed 100 characters.")
	ProviderSlugRequired    = N("PROVIDER_SLUG_REQUIRED", "Provider slug is required.")
	ProviderSlugMaxLength   = N("PROVIDER_SLUG_MAX_LENGTH", "Provider slug must not exceed 50 characters.")
	ProviderSlugDuplicated  = N("PROVIDER_SLUG_DUPLICATED", "An provider with this slug already exists.")
	ProviderNotFound        = N("PROVIDER_NOT_FOUND", "Provider was not found.")
	ProviderPlanCodeRequired    = N("PROVIDER_PLAN_CODE_REQUIRED", "Plan code is required.")
	ProviderPlanCodeMaxLength   = N("PROVIDER_PLAN_CODE_MAX_LENGTH", "Plan code must not exceed 64 characters.")
	ProviderPlanNameRequired    = N("PROVIDER_PLAN_NAME_REQUIRED", "Plan name is required.")
	ProviderPlanNameMaxLength   = N("PROVIDER_PLAN_NAME_MAX_LENGTH", "Plan name must not exceed 256 characters.")
	ProviderPlanCodeDuplicated  = N("PROVIDER_PLAN_CODE_DUPLICATED", "A plan with this code already exists for the operator.")
	ProviderPlanNotFound        = N("PROVIDER_PLAN_NOT_FOUND", "Plan was not found.")
)

// Customers
var (
	CustomerNameRequired              = N("CUSTOMER_NAME_REQUIRED", "Customer name is required.")
	CustomerNameMaxLength             = N("CUSTOMER_NAME_MAX_LENGTH", "Customer name must not exceed 256 characters.")
	CustomerDocumentRequired          = N("CUSTOMER_DOCUMENT_REQUIRED", "Customer document is required.")
	CustomerDocumentMaxLength         = N("CUSTOMER_DOCUMENT_MAX_LENGTH", "Customer document must not exceed 20 characters.")
	CustomerLegalNameRequiredForPJ    = N("CUSTOMER_LEGAL_NAME_REQUIRED_FOR_PJ", "Legal name is required for PJ customers.")
	CustomerDocumentDuplicated        = N("CUSTOMER_DOCUMENT_DUPLICATED", "A customer with this document already exists.")
	CustomerNotFound                  = N("CUSTOMER_NOT_FOUND", "Customer was not found.")
	CustomerBillingReadinessNotFound  = N("CUSTOMER_BILLING_READINESS_CONTEXT_NOT_FOUND", "Billing readiness context was not found.")
	CustomerProcessingMonthMismatch   = N("CUSTOMER_PROCESSING_MONTH_PROVIDER_MISMATCH", "Customer and processing month provider mismatch.")
	CustomerManualReleaseAlready      = N("CUSTOMER_MANUAL_RELEASE_ALREADY_EXISTS", "Manual release already exists for this customer and processing month.")
	CustomerManualReleaseJustMin      = N("CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_MIN_LENGTH", "Justification must be at least 10 characters.")
	CustomerManualReleaseJustMax      = N("CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_MAX_LENGTH", "Justification must not exceed 4000 characters.")
	CustomerAttachmentOriginalRequired = N("CUSTOMER_ATTACHMENT_ORIGINAL_FILE_NAME_REQUIRED", "Original file name is required.")
	CustomerAttachmentBucketRequired  = N("CUSTOMER_ATTACHMENT_STORAGE_BUCKET_REQUIRED", "Storage bucket is required.")
	CustomerAttachmentKeyRequired     = N("CUSTOMER_ATTACHMENT_STORAGE_OBJECT_KEY_REQUIRED", "Storage object key is required.")
)

// PhoneLines
var (
	PhoneLineNotFound                  = N("PHONE_LINE_NOT_FOUND", "Phone line was not found.")
	PhoneLineNumberRequired            = N("PHONE_LINE_NUMBER_REQUIRED", "Phone line number is required.")
	PhoneLineNumberDuplicated          = N("PHONE_LINE_NUMBER_DUPLICATED", "A phone line with this number already exists.")
	PhoneLineProviderAccountNotFound   = N("PHONE_LINE_PROVIDER_ACCOUNT_NOT_FOUND", "Provider account was not found for this operator.")
	PhoneLineProviderPlanInvalid       = N("PHONE_LINE_PROVIDER_PLAN_INVALID", "Selected plan does not belong to the operator.")
	PhoneLineActiveCustomerLinkNotFound = N("PHONE_LINE_ACTIVE_CUSTOMER_LINK_NOT_FOUND", "No active customer link was found for this phone line.")
	PhoneLineCustomerTransferSame      = N("PHONE_LINE_CUSTOMER_TRANSFER_SAME_CUSTOMER", "Phone line transfer requires a different target customer.")
	PhoneLineOperationPendingExists    = N("PHONE_LINE_OPERATION_PENDING_EXISTS", "There is already a pending operation request for this phone line.")
	PhoneLineOperationNotFound         = N("PHONE_LINE_OPERATION_NOT_FOUND", "Phone line operation request was not found.")
	PhoneLineOperationAlreadyReviewed  = N("PHONE_LINE_OPERATION_ALREADY_REVIEWED", "This operation request has already been reviewed.")
	PartnerCustomerAccessDenied        = N("PARTNER_CUSTOMER_ACCESS_DENIED", "You do not have access to this customer.")
	PartnerPhoneLineAccessDenied       = N("PARTNER_PHONE_LINE_ACCESS_DENIED", "You do not have access to this phone line.")
)

// Device stock
var (
	DeviceStockBrandRequired   = N("DEVICE_STOCK_BRAND_REQUIRED", "Device brand is required.")
	DeviceStockBrandMaxLength  = N("DEVICE_STOCK_BRAND_MAX_LENGTH", "Device brand must not exceed 128 characters.")
	DeviceStockModelRequired   = N("DEVICE_STOCK_MODEL_REQUIRED", "Device model is required.")
	DeviceStockModelMaxLength  = N("DEVICE_STOCK_MODEL_MAX_LENGTH", "Device model must not exceed 256 characters.")
	DeviceStockSkuMaxLength    = N("DEVICE_STOCK_SKU_MAX_LENGTH", "SKU must not exceed 64 characters.")
	DeviceStockSkuDuplicated   = N("DEVICE_STOCK_SKU_DUPLICATED", "A device with this SKU already exists.")
	DeviceStockImeiDuplicated  = N("DEVICE_STOCK_IMEI_DUPLICATED", "A device with this IMEI already exists.")
	DeviceStockImeiInvalid     = N("DEVICE_STOCK_IMEI_INVALID", "IMEI must have 15 digits.")
	DeviceStockNotFound        = N("DEVICE_STOCK_NOT_FOUND", "Device stock item was not found.")
	DeviceStockStatusInvalid   = N("DEVICE_STOCK_STATUS_INVALID", "Invalid device stock status.")
)

// Customer devices
var (
	CustomerDeviceDescriptionRequired = N("CUSTOMER_DEVICE_DESCRIPTION_REQUIRED", "Device description is required.")
	CustomerDeviceMonthlyAmountInvalid = N("CUSTOMER_DEVICE_MONTHLY_AMOUNT_INVALID", "Monthly amount must be zero or greater.")
	CustomerDeviceNotFound            = N("CUSTOMER_DEVICE_NOT_FOUND", "Customer device link was not found.")
	CustomerDeviceAlreadyLinked       = N("CUSTOMER_DEVICE_ALREADY_LINKED", "This device is already linked to an active customer.")
)

// Financial
var (
	FinancialDateRequired              = N("FINANCIAL_DATE_REQUIRED", "Date is required.")
	FinancialDateInvalid               = N("FINANCIAL_DATE_INVALID", "Date must be in YYYY-MM-DD format.")
	FinancialDescriptionRequired       = N("FINANCIAL_DESCRIPTION_REQUIRED", "Description is required.")
	FinancialAmountInvalid             = N("FINANCIAL_AMOUNT_INVALID", "Amount must be greater than zero.")
	FinancialPayableNotFound           = N("FINANCIAL_PAYABLE_NOT_FOUND", "Account payable was not found.")
	FinancialReceivableNotFound        = N("FINANCIAL_RECEIVABLE_NOT_FOUND", "Account receivable was not found.")
	FinancialPartnerSaleNotFound       = N("FINANCIAL_PARTNER_SALE_NOT_FOUND", "Partner sale record was not found.")
	FinancialPartnerSaleStatusInvalid  = N("FINANCIAL_PARTNER_SALE_STATUS_INVALID", "Invalid partner sale status.")
	FinancialCommissionPercentInvalid  = N("FINANCIAL_COMMISSION_PERCENT_INVALID", "Commission percent must be between 0 and 100.")
	FinancialPayableFromInvoiceExists  = N("FINANCIAL_PAYABLE_FROM_INVOICE_EXISTS", "An account payable already exists for this provider invoice.")
)

// Sales & contracts
var (
	SaleNotFound                  = N("SALE_NOT_FOUND", "Sale was not found.")
	SaleStatusInvalid             = N("SALE_STATUS_INVALID", "Sale status does not allow this operation.")
	SaleItemsRequired             = N("SALE_ITEMS_REQUIRED", "At least one line item is required to confirm the sale.")
	SaleLineItemInvalid           = N("SALE_LINE_ITEM_INVALID", "Invalid sale line item.")
	ContractTemplateNotFound      = N("CONTRACT_TEMPLATE_NOT_FOUND", "Contract template was not found.")
	ContractTemplateNameRequired  = N("CONTRACT_TEMPLATE_NAME_REQUIRED", "Contract template name is required.")
	ContractTemplateCodeRequired  = N("CONTRACT_TEMPLATE_CODE_REQUIRED", "Contract template code is required.")
	ContractTemplateBodyRequired  = N("CONTRACT_TEMPLATE_BODY_REQUIRED", "Contract template body is required.")
	ContractTemplateCodeDuplicated = N("CONTRACT_TEMPLATE_CODE_DUPLICATED", "A contract template with this code already exists.")
)

// BillingCycles
var (
	BillingCycleCodeRequired   = N("BILLING_CYCLE_CODE_REQUIRED", "Billing cycle code is required.")
	BillingCycleNameRequired   = N("BILLING_CYCLE_NAME_REQUIRED", "Billing cycle name is required.")
	BillingCycleNotFound       = N("BILLING_CYCLE_NOT_FOUND", "Billing cycle was not found.")
	BillingCycleConsolidated   = N("BILLING_CYCLE_CONSOLIDATED", "Billing cycle is consolidated and cannot be changed.")
)

// ProcessingMonths
var (
	ProcessingMonthNotFound              = N("PROCESSING_MONTH_NOT_FOUND", "Processing month was not found.")
	ProcessingMonthProviderMismatch      = N("PROCESSING_MONTH_PROVIDER_MISMATCH", "Processing month provider mismatch.")
	ProcessingMonthDuplicate             = N("PROCESSING_MONTH_DUPLICATE", "Processing month already exists for this provider, year and month.")
	ProcessingMonthAlreadyClosed         = N("PROCESSING_MONTH_ALREADY_CLOSED", "Processing month is already closed.")
	ProcessingMonthNotOpen               = N("PROCESSING_MONTH_NOT_OPEN", "Processing month is not open.")
	ProcessingMonthYearInvalid           = N("PROCESSING_MONTH_YEAR_INVALID", "Year must be between 2000 and 2100.")
	ProcessingMonthMonthInvalid          = N("PROCESSING_MONTH_MONTH_INVALID", "Month must be between 1 and 12.")
	ProcessingMonthDisplayNameRequired   = N("PROCESSING_MONTH_DISPLAY_NAME_REQUIRED", "Display name is required.")
	ProcessingMonthDisplayNameMaxLength  = N("PROCESSING_MONTH_DISPLAY_NAME_MAX_LENGTH", "Display name must not exceed 128 characters.")
	ProcessingMonthRetroactiveBlocked    = N("PROCESSING_MONTH_RETROACTIVE_CHANGE_BLOCKED", "Retroactive change blocked by closed processing month.")
	ProcessingMonthContingencyJustMin    = N("PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_MIN_LENGTH", "Justification must be at least 10 characters.")
	ProcessingMonthContingencyJustMax    = N("PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_MAX_LENGTH", "Justification must not exceed 4000 characters.")
)

// Invoices
var (
	InvoiceNotFound                      = N("INVOICE_NOT_FOUND", "Invoice was not found.")
	InvoiceDuplicateSameProcessingMonth  = N("INVOICE_DUPLICATE_SAME_PROCESSING_MONTH", "Invoice duplicate for same processing month.")
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
)

// ObjectStorage
var (
	ObjectStorageUnavailable = N("OBJECT_STORAGE_UNAVAILABLE", "Object storage is not configured.")
	PresignedExpiresInvalid    = N("PRESIGNED_EXPIRES_IN_SECONDS_INVALID", "Expires in seconds must be between 60 and 604800.")
	ObjectKeyInvalid           = N("OBJECT_KEY_INVALID", "Object key is invalid.")
)

// Billing & email
var (
	BillingEmailTemplateNotFound       = N("BILLING_EMAIL_TEMPLATE_NOT_FOUND", "Invoice email template was not found.")
	BillingEmailTemplateCodeDuplicated = N("BILLING_EMAIL_TEMPLATE_CODE_DUPLICATED", "Template code already exists.")
	BillingEmailTemplateFieldsRequired = N("BILLING_EMAIL_TEMPLATE_FIELDS_REQUIRED", "Template name, subject and body are required.")
	BillingDocumentNotFound            = N("BILLING_DOCUMENT_NOT_FOUND", "Customer billing document was not found.")
	BillingDocumentFieldsRequired      = N("BILLING_DOCUMENT_FIELDS_REQUIRED", "Recipient, subject and body are required.")
	BillingDocumentCancelled           = N("BILLING_DOCUMENT_CANCELLED", "Cancelled billing documents cannot be sent.")
	BillingCustomerEmailRequired       = N("BILLING_CUSTOMER_EMAIL_REQUIRED", "Customer billing email is required.")
	BillingEmailNotConfigured          = N("BILLING_EMAIL_NOT_CONFIGURED", "SMTP is not configured. Set SMTP_HOST and related env vars.")
	BillingReceivableNotOverdue        = N("BILLING_RECEIVABLE_NOT_OVERDUE", "Only overdue receivables can receive collection reminders.")
	BillingLayoutTemplateNotFound      = N("BILLING_LAYOUT_TEMPLATE_NOT_FOUND", "Invoice layout template was not found.")
	BillingLayoutTemplateCodeDuplicated = N("BILLING_LAYOUT_TEMPLATE_CODE_DUPLICATED", "Layout template code already exists.")
	BillingLayoutTemplateFieldsRequired = N("BILLING_LAYOUT_TEMPLATE_FIELDS_REQUIRED", "Layout name and configuration are required.")
	SicrediNotConfigured              = N("SICREDI_NOT_CONFIGURED", "Sicredi billing API is not configured.")
	SicrediBoletoAlreadyIssued        = N("SICREDI_BOLETO_ALREADY_ISSUED", "A Sicredi boleto was already issued for this invoice.")
)
