using Luxus.Connect.Infra.Crosscutting.Notifications;

namespace Luxus.Connect.Infra.Crosscutting.Constants;

public class Notifications
{
    public readonly struct Shared
    {
        public static readonly Notification UNEXPECTED_ERROR = new(nameof(UNEXPECTED_ERROR), "We're sorry... An unexpected problem has occurred. Please wait, our team is already working to resolve it as soon as possible.");
        public static readonly Notification RESOURCE_NOT_FOUND = new(nameof(RESOURCE_NOT_FOUND), $"The requested resource was not found");
        public static readonly Notification DOMAIN_VIOLATION = new(nameof(DOMAIN_VIOLATION), $"An business rule violation has occurred.");
        public static readonly Notification REQUEST_VALIDATION = new(nameof(REQUEST_VALIDATION), $"The send data is not valid.");
        public static readonly Notification SERVICE_UNAVAILABLE = new(nameof(SERVICE_UNAVAILABLE), $"One or more services may be unavailable");
        public static readonly Notification SAVING_DATA_FAILURE = new(nameof(SAVING_DATA_FAILURE), "Opss... An error occurred while saving the data");
        public static readonly Notification ORGANIZATION_ID_REQUIRED = new(nameof(ORGANIZATION_ID_REQUIRED), "Organization ID is required.");

        public static Notification UnexpectedError(string message) => new(nameof(UNEXPECTED_ERROR), message);
    }

    public readonly struct Providers
    {
        public static readonly Notification PROVIDER_NAME_REQUIRED = new(nameof(PROVIDER_NAME_REQUIRED), "Provider name is required.");
        public static readonly Notification PROVIDER_NAME_MAX_LENGTH = new(nameof(PROVIDER_NAME_MAX_LENGTH), "Provider name must not exceed 100 characters.");
        public static readonly Notification PROVIDER_SLUG_REQUIRED = new(nameof(PROVIDER_SLUG_REQUIRED), "Provider slug is required.");
        public static readonly Notification PROVIDER_SLUG_MAX_LENGTH = new(nameof(PROVIDER_SLUG_MAX_LENGTH), "Provider slug must not exceed 50 characters.");
        public static readonly Notification PROVIDER_SLUG_DUPLICATED = new(nameof(PROVIDER_SLUG_DUPLICATED), "An provider with this slug already exists.");
        public static readonly Notification PROVIDER_NOT_FOUND = new(nameof(PROVIDER_NOT_FOUND), "Provider was not found.");
    }

    public readonly struct PhoneLines
    {
        public static readonly Notification PHONE_LINE_ID_REQUIRED = new(nameof(PHONE_LINE_ID_REQUIRED), "Phone line id is required.");
        public static readonly Notification PHONE_LINE_PROVIDER_REQUIRED = new(nameof(PHONE_LINE_PROVIDER_REQUIRED), "Provider is required.");
        public static readonly Notification PHONE_LINE_NUMBER_DUPLICATED = new(nameof(PHONE_LINE_NUMBER_DUPLICATED), "A phone line with this number already exists.");
        public static readonly Notification PHONE_LINE_NOT_FOUND = new(nameof(PHONE_LINE_NOT_FOUND), "Phone line was not found.");
        public static readonly Notification PHONE_LINE_PROVIDER_NOT_FOUND = new(nameof(PHONE_LINE_PROVIDER_NOT_FOUND), "Provider was not found.");
        public static readonly Notification PHONE_LINE_CUSTOMER_PROVIDER_MISMATCH = new(
            nameof(PHONE_LINE_CUSTOMER_PROVIDER_MISMATCH),
            "Phone line and customer must belong to the same provider.");
        public static readonly Notification PHONE_LINE_ACTIVE_CUSTOMER_LINK_NOT_FOUND = new(
            nameof(PHONE_LINE_ACTIVE_CUSTOMER_LINK_NOT_FOUND),
            "No active customer link was found for this phone line.");
        public static readonly Notification PHONE_LINE_CUSTOMER_TRANSFER_SAME_CUSTOMER = new(
            nameof(PHONE_LINE_CUSTOMER_TRANSFER_SAME_CUSTOMER),
            "Phone line transfer requires a different target customer.");

        public static readonly Notification PHONE_LINE_PLAN_REQUIRED_WITH_CUSTOMER = new(
            nameof(PHONE_LINE_PLAN_REQUIRED_WITH_CUSTOMER),
            "Com cliente vinculado, o plano é obrigatório.");

        public static readonly Notification PHONE_LINE_INITIAL_SERVICE_REQUIRED_WITH_CUSTOMER = new(
            nameof(PHONE_LINE_INITIAL_SERVICE_REQUIRED_WITH_CUSTOMER),
            "Com cliente vinculado, é obrigatório informar o serviço do plano (assinatura inicial) na linha.");

        public static readonly Notification PHONE_LINE_PLAN_SERVICE_NOT_IN_PLAN = new(
            nameof(PHONE_LINE_PLAN_SERVICE_NOT_IN_PLAN),
            "O serviço do plano informado não pertence ao plano da linha.");

        public static readonly Notification PHONE_LINE_TRANSITION_REQUIRES_CUSTOMER = new(
            nameof(PHONE_LINE_TRANSITION_REQUIRES_CUSTOMER),
            "Linha em transição exige cliente vinculado.");

        public static readonly Notification PHONE_LINE_TRANSITION_REQUIRES_PLAN = new(
            nameof(PHONE_LINE_TRANSITION_REQUIRES_PLAN),
            "Linha em transição exige plano.");

        public static readonly Notification PHONE_LINE_TRANSITION_REQUIRES_INITIAL_SERVICE = new(
            nameof(PHONE_LINE_TRANSITION_REQUIRES_INITIAL_SERVICE),
            "Linha em transição exige serviço inicial (assinatura) no plano.");

        public static readonly Notification PHONE_LINE_TRANSITION_SUB_STATUS_INVALID = new(
            nameof(PHONE_LINE_TRANSITION_SUB_STATUS_INVALID),
            "Para linha em transição informe o subtipo: portabilidade, TT ou PP.");

        public static readonly Notification PHONE_LINE_TRANSITION_SUB_STATUS_WITHOUT_FLAG = new(
            nameof(PHONE_LINE_TRANSITION_SUB_STATUS_WITHOUT_FLAG),
            "Subtipo de transição só se aplica quando start_in_transition é verdadeiro.");
    }

    public readonly struct ProviderServices
    {
        public static readonly Notification PROVIDER_SERVICE_NAME_REQUIRED = new(nameof(PROVIDER_SERVICE_NAME_REQUIRED), "Service name is required.");
        public static readonly Notification PROVIDER_SERVICE_NAME_MAX_LENGTH = new(nameof(PROVIDER_SERVICE_NAME_MAX_LENGTH), "Service name must not exceed 256 characters.");
        public static readonly Notification PROVIDER_SERVICE_PROVIDER_REQUIRED = new(nameof(PROVIDER_SERVICE_PROVIDER_REQUIRED), "Provider is required.");
        public static readonly Notification PROVIDER_SERVICE_PLAN_REQUIRED = new(nameof(PROVIDER_SERVICE_PLAN_REQUIRED), "Plan is required.");
        public static readonly Notification PROVIDER_SERVICE_NAME_DUPLICATED = new(nameof(PROVIDER_SERVICE_NAME_DUPLICATED), "A service with this name already exists for this provider.");
        public static readonly Notification PROVIDER_SERVICE_NOT_FOUND = new(nameof(PROVIDER_SERVICE_NOT_FOUND), "Provider service was not found.");
        public static readonly Notification PROVIDER_SERVICE_PROVIDER_NOT_FOUND = new(nameof(PROVIDER_SERVICE_PROVIDER_NOT_FOUND), "Provider was not found.");
    }

    public readonly struct BillingCycles
    {
        public static readonly Notification BILLING_CYCLE_CODE_REQUIRED = new(nameof(BILLING_CYCLE_CODE_REQUIRED), "Billing cycle code is required.");
        public static readonly Notification BILLING_CYCLE_NAME_REQUIRED = new(nameof(BILLING_CYCLE_NAME_REQUIRED), "Billing cycle name is required.");
        public static readonly Notification BILLING_CYCLE_NOT_FOUND = new(nameof(BILLING_CYCLE_NOT_FOUND), "Billing cycle was not found.");
        public static readonly Notification BILLING_CYCLE_CONSOLIDATED = new(nameof(BILLING_CYCLE_CONSOLIDATED), "Billing cycle is consolidated and cannot be changed.");
    }

    public readonly struct LineServices
    {
        public static readonly Notification LINE_SERVICE_LINE_REQUIRED = new(nameof(LINE_SERVICE_LINE_REQUIRED), "Phone line is required.");
        public static readonly Notification LINE_SERVICE_SERVICE_REQUIRED = new(nameof(LINE_SERVICE_SERVICE_REQUIRED), "Provider service is required.");
        public static readonly Notification LINE_SERVICE_START_DATE_REQUIRED = new(nameof(LINE_SERVICE_START_DATE_REQUIRED), "Start date is required.");
        public static readonly Notification LINE_SERVICE_DUPLICATED = new(nameof(LINE_SERVICE_DUPLICATED), "This service is already linked to this phone line.");
        public static readonly Notification LINE_SERVICE_NOT_FOUND = new(nameof(LINE_SERVICE_NOT_FOUND), "Line service was not found.");
        public static readonly Notification LINE_SERVICE_PHONE_LINE_NOT_FOUND = new(nameof(LINE_SERVICE_PHONE_LINE_NOT_FOUND), "Phone line was not found.");
        public static readonly Notification LINE_SERVICE_PROVIDER_SERVICE_NOT_FOUND = new(nameof(LINE_SERVICE_PROVIDER_SERVICE_NOT_FOUND), "Provider service was not found.");
    }

    public readonly struct ContractingCompanies
    {
        public static readonly Notification CONTRACTING_COMPANY_NOT_FOUND = new(
            nameof(CONTRACTING_COMPANY_NOT_FOUND),
            "Empresa contratante não encontrada.");

        public static readonly Notification CONTRACTING_COMPANY_PROVIDER_MISMATCH = new(
            nameof(CONTRACTING_COMPANY_PROVIDER_MISMATCH),
            "A empresa contratante não pertence à operadora informada.");

        public static readonly Notification CONTRACTING_COMPANY_LEGAL_NAME_REQUIRED = new(
            nameof(CONTRACTING_COMPANY_LEGAL_NAME_REQUIRED),
            "Razão social da empresa contratante é obrigatória.");

        public static readonly Notification CONTRACTING_COMPANY_LEGAL_NAME_MAX_LENGTH = new(
            nameof(CONTRACTING_COMPANY_LEGAL_NAME_MAX_LENGTH),
            "Razão social não pode exceder 512 caracteres.");

        public static readonly Notification CONTRACTING_COMPANY_TAX_ID_REQUIRED = new(
            nameof(CONTRACTING_COMPANY_TAX_ID_REQUIRED),
            "CNPJ da empresa contratante é obrigatório (14 dígitos).");

        public static readonly Notification CONTRACTING_COMPANY_TAX_ID_DUPLICATED = new(
            nameof(CONTRACTING_COMPANY_TAX_ID_DUPLICATED),
            "Já existe empresa contratante com este CNPJ para esta operadora.");
    }

    public readonly struct Customers
    {
        public static readonly Notification CUSTOMER_CONTRACTING_COMPANY_MISMATCH = new(
            nameof(CUSTOMER_CONTRACTING_COMPANY_MISMATCH),
            "O cliente já está vinculado a outra empresa contratante; ajuste o vínculo ou use o contratante correto na importação.");

        public static readonly Notification CUSTOMER_DOCUMENT_OTHER_PROVIDER = new(
            nameof(CUSTOMER_DOCUMENT_OTHER_PROVIDER),
            "Já existe cadastro com este documento em outra operadora.");

        public static readonly Notification CUSTOMER_NAME_REQUIRED = new(nameof(CUSTOMER_NAME_REQUIRED), "Customer name is required.");
        public static readonly Notification CUSTOMER_NAME_MAX_LENGTH = new(nameof(CUSTOMER_NAME_MAX_LENGTH), "Customer name must not exceed 256 characters.");
        public static readonly Notification CUSTOMER_RESPONSIBLE_SALESPERSON_USER_ID_MAX_LENGTH = new(
            nameof(CUSTOMER_RESPONSIBLE_SALESPERSON_USER_ID_MAX_LENGTH),
            "Identificador do vendedor responsável não pode exceder 256 caracteres.");
        public static readonly Notification CUSTOMER_DOCUMENT_REQUIRED = new(nameof(CUSTOMER_DOCUMENT_REQUIRED), "Customer document (CPF/CNPJ) is required.");
        public static readonly Notification CUSTOMER_DOCUMENT_MAX_LENGTH = new(nameof(CUSTOMER_DOCUMENT_MAX_LENGTH), "Customer document must not exceed 20 characters.");
        public static readonly Notification CUSTOMER_LEGAL_NAME_REQUIRED_FOR_PJ = new(nameof(CUSTOMER_LEGAL_NAME_REQUIRED_FOR_PJ), "Legal name is required for PJ customers.");
        public static readonly Notification CUSTOMER_PROVIDER_REQUIRED = new(nameof(CUSTOMER_PROVIDER_REQUIRED), "Provider is required.");
        public static readonly Notification CUSTOMER_CONTRACTING_COMPANY_REQUIRED = new(
            nameof(CUSTOMER_CONTRACTING_COMPANY_REQUIRED),
            "Empresa contratante é obrigatória.");
        public static readonly Notification CUSTOMER_DOCUMENT_DUPLICATED = new(nameof(CUSTOMER_DOCUMENT_DUPLICATED), "A customer with this document already exists.");
        public static readonly Notification CUSTOMER_NOT_FOUND = new(nameof(CUSTOMER_NOT_FOUND), "Customer was not found.");

        public static readonly Notification CUSTOMER_ID_REQUIRED = new(nameof(CUSTOMER_ID_REQUIRED), "Customer id is required.");

        public static readonly Notification CUSTOMER_ATTACHMENT_NOT_FOUND = new(
            nameof(CUSTOMER_ATTACHMENT_NOT_FOUND),
            "Anexo do cliente não encontrado.");

        public static readonly Notification CUSTOMER_ATTACHMENT_ID_REQUIRED = new(
            nameof(CUSTOMER_ATTACHMENT_ID_REQUIRED),
            "Identificador do anexo é obrigatório.");

        public static readonly Notification CUSTOMER_ATTACHMENT_TITLE_MAX_LENGTH = new(
            nameof(CUSTOMER_ATTACHMENT_TITLE_MAX_LENGTH),
            "Título do anexo não pode exceder 256 caracteres.");

        public static readonly Notification CUSTOMER_ATTACHMENT_ORIGINAL_FILE_NAME_REQUIRED = new(
            nameof(CUSTOMER_ATTACHMENT_ORIGINAL_FILE_NAME_REQUIRED),
            "Nome original do arquivo é obrigatório.");

        public static readonly Notification CUSTOMER_ATTACHMENT_SIZE_BYTES_TOO_LARGE = new(
            nameof(CUSTOMER_ATTACHMENT_SIZE_BYTES_TOO_LARGE),
            "Tamanho do arquivo excede o limite permitido (256 MB).");

        public static readonly Notification CUSTOMER_ATTACHMENT_CONTENT_TYPE_MAX_LENGTH = new(
            nameof(CUSTOMER_ATTACHMENT_CONTENT_TYPE_MAX_LENGTH),
            "Content-Type não pode exceder 128 caracteres.");

        public static readonly Notification CUSTOMER_BILLING_READINESS_CONTEXT_NOT_FOUND = new(
            nameof(CUSTOMER_BILLING_READINESS_CONTEXT_NOT_FOUND),
            "Cliente ou mês de processamento não encontrado, ou não pertencem à mesma operadora.");

        public static readonly Notification CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_REQUIRED = new(
            nameof(CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_REQUIRED),
            "Justificativa da liberação manual é obrigatória (mínimo 10 caracteres).");

        public static readonly Notification CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_MAX_LENGTH = new(
            nameof(CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_MAX_LENGTH),
            "Justificativa não pode exceder 4000 caracteres.");

        public static readonly Notification CUSTOMER_PROCESSING_MONTH_PROVIDER_MISMATCH = new(
            nameof(CUSTOMER_PROCESSING_MONTH_PROVIDER_MISMATCH),
            "O cliente e o mês de processamento devem pertencer à mesma operadora.");

        public static readonly Notification CUSTOMER_MANUAL_RELEASE_ALREADY_EXISTS = new(
            nameof(CUSTOMER_MANUAL_RELEASE_ALREADY_EXISTS),
            "Já existe liberação manual registrada para este cliente neste mês de processamento.");
    }

    public readonly struct Invoices
    {
        public static readonly Notification INVOICE_DUPLICATE_SAME_PROCESSING_MONTH = new(
            nameof(INVOICE_DUPLICATE_SAME_PROCESSING_MONTH),
            "Já existe fatura com a mesma operadora, empresa contratante, conta e vencimento neste mês de processamento.");

        public static readonly Notification INVOICE_DUPLICATE_OTHER_PROCESSING_MONTH = new(
            nameof(INVOICE_DUPLICATE_OTHER_PROCESSING_MONTH),
            "Esta fatura (mesma operadora, empresa contratante, conta e vencimento) já foi registrada em outro mês de processamento.");

        public static readonly Notification INVOICE_PROCESSING_MONTH_REQUIRED = new(nameof(INVOICE_PROCESSING_MONTH_REQUIRED), "Processing month is required.");

        public static readonly Notification INVOICE_CYCLE_REQUIRED = new(nameof(INVOICE_CYCLE_REQUIRED), "Billing cycle is required.");
        public static readonly Notification INVOICE_NUMBER_REQUIRED = new(nameof(INVOICE_NUMBER_REQUIRED), "Invoice number is required.");
        public static readonly Notification INVOICE_NUMBER_MAX_LENGTH = new(nameof(INVOICE_NUMBER_MAX_LENGTH), "Invoice number must not exceed 64 characters.");
        public static readonly Notification INVOICE_ITEMS_REQUIRED = new(nameof(INVOICE_ITEMS_REQUIRED), "At least one invoice line item is required.");
        public static readonly Notification INVOICE_ACCOUNT_NUMBER_DUPLICATED_IN_CYCLE = new(nameof(INVOICE_ACCOUNT_NUMBER_DUPLICATED_IN_CYCLE), "An invoice with this number already exists for the billing cycle.");
        public static readonly Notification INVOICE_BILLING_CYCLE_NOT_FOUND = new(nameof(INVOICE_BILLING_CYCLE_NOT_FOUND), "Billing cycle was not found.");
        public static readonly Notification INVOICE_PHONE_LINE_NOT_FOUND = new(nameof(INVOICE_PHONE_LINE_NOT_FOUND), "Phone line was not found.");
        public static readonly Notification INVOICE_PARENT_NOT_FOUND = new(nameof(INVOICE_PARENT_NOT_FOUND), "Parent invoice was not found.");
        public static readonly Notification INVOICE_PROVIDER_SERVICE_NOT_FOUND = new(nameof(INVOICE_PROVIDER_SERVICE_NOT_FOUND), "Provider service was not found.");
        public static readonly Notification INVOICE_LINE_SERVICE_NOT_FOUND = new(nameof(INVOICE_LINE_SERVICE_NOT_FOUND), "The phone line does not have this provider service subscribed.");
        public static readonly Notification INVOICE_LINE_REQUIRED_FOR_SERVICE_ITEM = new(nameof(INVOICE_LINE_REQUIRED_FOR_SERVICE_ITEM), "Invoice line items that reference a service must include a phone line.");
        public static readonly Notification INVOICE_NOT_FOUND = new(nameof(INVOICE_NOT_FOUND), "Invoice was not found.");
        public static readonly Notification INVOICE_CUSTOMER_ID_REQUIRED = new(nameof(INVOICE_CUSTOMER_ID_REQUIRED), "Customer is required.");

        public static readonly Notification INVOICE_CUSTOMER_NOT_FOUND = new(nameof(INVOICE_CUSTOMER_NOT_FOUND), "Customer was not found.");

        /// <summary>§3.1 — linha na fatura importada sem destino válido (ex.: ativa/aguardando/em transição sem cliente).</summary>
        public static readonly Notification INVOICE_IMPORTED_LINE_ORPHAN_DESTINATION = new(
            nameof(INVOICE_IMPORTED_LINE_ORPHAN_DESTINATION),
            "Linha órfã: existe na fatura sem vínculo a cliente ativo nem situação de estoque permitida (regra estrutural §3.1).");

        public static readonly Notification INVOICE_IMPORTED_LINE_CUSTOMER_MISMATCH = new(
            nameof(INVOICE_IMPORTED_LINE_CUSTOMER_MISMATCH),
            "Linha na fatura com cliente diferente do titular da fatura importada (inconsistência de vínculo).");
    }

    public readonly struct ProcessingMonths
    {
        public static readonly Notification PROCESSING_MONTH_PROVIDER_REQUIRED = new(nameof(PROCESSING_MONTH_PROVIDER_REQUIRED), "Provider is required.");
        public static readonly Notification PROCESSING_MONTH_YEAR_INVALID = new(nameof(PROCESSING_MONTH_YEAR_INVALID), "Year must be between 2000 and 2100.");
        public static readonly Notification PROCESSING_MONTH_MONTH_INVALID = new(nameof(PROCESSING_MONTH_MONTH_INVALID), "Month must be between 1 and 12.");
        public static readonly Notification PROCESSING_MONTH_DISPLAY_NAME_REQUIRED = new(nameof(PROCESSING_MONTH_DISPLAY_NAME_REQUIRED), "Display name is required.");
        public static readonly Notification PROCESSING_MONTH_DISPLAY_NAME_MAX_LENGTH = new(nameof(PROCESSING_MONTH_DISPLAY_NAME_MAX_LENGTH), "Display name must not exceed 128 characters.");
        public static readonly Notification PROCESSING_MONTH_NOT_FOUND = new(nameof(PROCESSING_MONTH_NOT_FOUND), "Processing month was not found.");
        public static readonly Notification PROCESSING_MONTH_PROVIDER_MISMATCH = new(nameof(PROCESSING_MONTH_PROVIDER_MISMATCH), "Processing month does not belong to the specified provider.");
        public static readonly Notification PROCESSING_MONTH_DUPLICATE = new(nameof(PROCESSING_MONTH_DUPLICATE), "A processing month already exists for this provider and calendar month.");
        public static readonly Notification PROCESSING_MONTH_ID_REQUIRED = new(nameof(PROCESSING_MONTH_ID_REQUIRED), "Processing month id is required.");
        public static readonly Notification PROCESSING_MONTH_ALREADY_CLOSED = new(
            nameof(PROCESSING_MONTH_ALREADY_CLOSED),
            "Este mês de processamento já está encerrado.");

        public static readonly Notification PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_REQUIRED = new(
            nameof(PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_REQUIRED),
            "Justificativa do encerramento em contingência é obrigatória (mínimo 10 caracteres).");

        public static readonly Notification PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_MAX_LENGTH = new(
            nameof(PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_MAX_LENGTH),
            "Justificativa não pode exceder 4000 caracteres.");
        public static readonly Notification PROCESSING_MONTH_NOT_OPEN = new(
            nameof(PROCESSING_MONTH_NOT_OPEN),
            "O mês de processamento está encerrado; não é possível registrar novas faturas, importações ou solicitações de importação.");

        public static readonly Notification PROCESSING_MONTH_RETROACTIVE_CHANGE_BLOCKED = new(
            nameof(PROCESSING_MONTH_RETROACTIVE_CHANGE_BLOCKED),
            "O intervalo de datas intersecta um mês de processamento já encerrado; alterações retroativas de vigência não são permitidas (§11.3).");
    }

    public readonly struct ObjectStorage
    {
        public static readonly Notification PRESIGNED_EXPIRES_IN_SECONDS_INVALID = new(
            nameof(PRESIGNED_EXPIRES_IN_SECONDS_INVALID),
            "expires_in_seconds must be between 60 and 604800 (7 days).");

        public static readonly Notification OBJECT_KEY_INVALID = new(
            nameof(OBJECT_KEY_INVALID),
            "object_key contains invalid characters.");
    }

    public readonly struct InvoiceImports
    {
        public static readonly Notification PROVIDER_ID_REQUIRED = new(nameof(PROVIDER_ID_REQUIRED), "Provider id is required.");
        public static readonly Notification PROCESSING_MONTH_ID_REQUIRED = new(nameof(PROCESSING_MONTH_ID_REQUIRED), "Processing month id is required.");
        public static readonly Notification IMPORT_REQUEST_ID_REQUIRED = new(nameof(STORAGE_BUCKET_REQUIRED), "Invoice import id is required.");
        public static readonly Notification STORAGE_BUCKET_REQUIRED = new(nameof(STORAGE_BUCKET_REQUIRED), "Storage bucket is required.");
        public static readonly Notification STORAGE_BUCKET_MAX_LENGTH = new(nameof(STORAGE_BUCKET_MAX_LENGTH), "Storage bucket must not exceed 256 characters.");
        public static readonly Notification STORAGE_OBJECT_KEY_REQUIRED = new(nameof(STORAGE_OBJECT_KEY_REQUIRED), "Storage object key is required.");
        public static readonly Notification STORAGE_OBJECT_KEY_MAX_LENGTH = new(nameof(STORAGE_OBJECT_KEY_MAX_LENGTH), "Storage object key must not exceed 2048 characters.");
        public static readonly Notification ORIGINAL_FILE_NAME_MAX_LENGTH = new(nameof(ORIGINAL_FILE_NAME_MAX_LENGTH), "Original file name must not exceed 512 characters.");
        public static readonly Notification IMPORT_REQUEST_NOT_FOUND = new(nameof(IMPORT_REQUEST_NOT_FOUND), "Invoice import request was not found.");
        public static readonly Notification IMPORT_REQUEST_NOT_PENDING = new(nameof(IMPORT_REQUEST_NOT_PENDING), "Invoice import request is not pending.");
        public static readonly Notification PROVIDER_NOT_FOUND = new(nameof(PROVIDER_NOT_FOUND), "Provider not found");

        public static readonly Notification REFERENCE_MONTH_INVALID = new(
            nameof(REFERENCE_MONTH_INVALID),
            "Mês de referência (010D) inválido ou ausente; esperado AAAAMM.");

        public static readonly Notification PROCESSING_MONTH_NOT_FOUND_FOR_FILE = new(
            nameof(PROCESSING_MONTH_NOT_FOUND_FOR_FILE),
            "Não existe mês de processamento cadastrado para esta operadora e referência (AAAAMM) do arquivo.");

        public static readonly Notification CONTRACTING_COMPANY_NOT_FOUND_FOR_FILE = new(
            nameof(CONTRACTING_COMPANY_NOT_FOUND_FOR_FILE),
            "Não foi encontrada empresa contratante com o CNPJ do arquivo para esta operadora.");

        public static readonly Notification CPF_REQUIRES_EXISTING_CUSTOMER_FOR_IMPORT = new(
            nameof(CPF_REQUIRES_EXISTING_CUSTOMER_FOR_IMPORT),
            "Para documento CPF no arquivo é necessário cliente já cadastrado na operadora para determinar a empresa contratante.");

        public static readonly Notification CUSTOMER_DOCUMENT_INVALID_FOR_IMPORT = new(
            nameof(CUSTOMER_DOCUMENT_INVALID_FOR_IMPORT),
            "Documento do cliente (011D) inválido para importação: esperado CNPJ (14 dígitos) ou CPF (11 dígitos).");
    }
}
