package parceiros

import (
	"encoding/base64"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// safeExternalID is also used as a path segment for stored attachments, so it
// must never contain separators, traversal segments or leading dots.
var safeExternalID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

const (
	Origin          = "luxus_parceiros"
	maxName         = 256
	maxNote         = 4000
	maxDecodedBytes = 12_000_000
)

type Sale struct {
	ID                string   `json:"id"`
	Protocol          string   `json:"protocol"`
	TaskProtocol      string   `json:"taskProtocol"`
	SoldAt            string   `json:"soldAt"`
	ConcludedAt       string   `json:"concludedAt"`
	Value             *float64 `json:"value"`
	CommissionValue   *float64 `json:"commissionValue"`
	OperatorName      string   `json:"operatorName"`
	PlanName          string   `json:"planName"`
	LineNumber        string   `json:"lineNumber"`
	SimType           string   `json:"simType"`
	IsVirginChip      bool     `json:"isVirginChip"`
	Iccid             string   `json:"iccid"`
	DeviceImei        string   `json:"deviceImei"`
	DeviceEid         string   `json:"deviceEid"`
	DevicePlatform    string   `json:"devicePlatform"`
	IsPortability     bool     `json:"isPortability"`
	PortabilityNumber string   `json:"portabilityNumber"`
	DonorOperator     string   `json:"donorOperator"`
	IsUpgrade         bool     `json:"isUpgrade"`
	UpgradeOfProtocol string   `json:"upgradeOfProtocol"`
	ContractFormat    string   `json:"contractFormat"`
	PartnerName       string   `json:"partnerName"`
	BranchName        string   `json:"branchName"`
	Notes             string   `json:"notes"`
}

type Client struct {
	Name          string `json:"name"`
	Document      string `json:"document"`
	DocumentType  string `json:"documentType"`
	RG            string `json:"rg"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	AddressNumber string `json:"addressNumber"`
	Complement    string `json:"complement"`
	Neighborhood  string `json:"neighborhood"`
	City          string `json:"city"`
	State         string `json:"state"`
	ZipCode       string `json:"zipCode"`
}

type Attachment struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	MimeType      string `json:"mimeType"`
	ContentBase64 string `json:"contentBase64"`
}

type SaleIntake struct {
	Sale        Sale         `json:"sale"`
	Client      Client       `json:"client"`
	Attachments []Attachment `json:"attachments"`
}

type Prepared struct {
	ExternalSaleID    string
	Protocol          string
	TaskProtocol      string
	CustomerName      string
	CustomerType      string
	DocumentDigits    string
	DocumentKind      string
	LookupDocument    bool
	ReviewPending     bool
	RawDocument       string
	RG                string
	Phone             string
	Email             string
	RawEmail          string
	RawAddress        string
	StructuredAddress *Address
	LineDigits        string
	SimType           string
	Iccid             string
	DeviceImei        string
	DeviceEid         string
	DevicePlatform    string
	SoldAt            *time.Time
	ConcludedAt       *time.Time
	Warnings          []string
}

type Address struct {
	Street       string
	Number       string
	Neighborhood string
	City         string
	State        string
	ZipCode      string
	Complement   string
}

type IntakeResult struct {
	CustomerID          string   `json:"customerId"`
	CustomerCreated     bool     `json:"customerCreated"`
	PhoneLineID         *string  `json:"phoneLineId"`
	PhoneLineCreated    bool     `json:"phoneLineCreated"`
	AttachmentsAccepted int      `json:"attachmentsAccepted"`
	AttachmentsSkipped  int      `json:"attachmentsSkipped"`
	Warnings            []string `json:"warnings"`
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func Prepare(input SaleIntake) (Prepared, error) {
	var prepared Prepared
	prepared.ExternalSaleID = strings.TrimSpace(input.Sale.ID)
	prepared.Protocol = strings.TrimSpace(input.Sale.Protocol)
	prepared.TaskProtocol = strings.TrimSpace(input.Sale.TaskProtocol)
	prepared.CustomerName = strings.TrimSpace(input.Client.Name)
	if prepared.ExternalSaleID == "" || prepared.Protocol == "" || prepared.CustomerName == "" {
		return prepared, &ValidationError{Message: "Informe sale.id, sale.protocol e o nome do cliente."}
	}
	if len(prepared.ExternalSaleID) > 64 || len(prepared.Protocol) > 64 {
		return prepared, &ValidationError{Message: "sale.id e sale.protocol precisam ter no máximo 64 caracteres."}
	}
	if !safeExternalID.MatchString(prepared.ExternalSaleID) {
		return prepared, &ValidationError{Message: "sale.id deve conter apenas letras, números, ponto, hífen ou sublinhado, sem barras ou '..'."}
	}
	if utf8.RuneCountInString(prepared.CustomerName) > maxName {
		prepared.Warnings = append(prepared.Warnings, "Nome do cliente passou de 256 caracteres e foi encurtado na ficha. O texto completo fica no payload da venda.")
		prepared.CustomerName = string([]rune(prepared.CustomerName)[:maxName])
	}

	rawDocument := strings.TrimSpace(input.Client.Document)
	digits := digitsOnly(rawDocument)
	switch {
	case len(digits) == 11:
		prepared.DocumentDigits = digits
		prepared.DocumentKind = "cpf"
		prepared.CustomerType = "pf"
		prepared.LookupDocument = true
	case len(digits) == 14:
		prepared.DocumentDigits = digits
		prepared.DocumentKind = "cnpj"
		prepared.CustomerType = "pj"
		prepared.LookupDocument = true
	default:
		prepared.CustomerType = "pf"
		prepared.ReviewPending = true
		prepared.RawDocument = rawDocument
		if rawDocument != "" {
			prepared.Warnings = append(prepared.Warnings, "Documento fora do padrão de CPF ou CNPJ. A venda entrou e o cliente ficou pendente de revisão.")
		} else {
			prepared.Warnings = append(prepared.Warnings, "Venda sem CPF. O cliente ficou pendente de revisão.")
		}
	}
	if strings.EqualFold(strings.TrimSpace(input.Client.DocumentType), "CNPJ") && prepared.CustomerType == "pf" && prepared.LookupDocument {
		prepared.Warnings = append(prepared.Warnings, "documentType veio como CNPJ, mas o número tem tamanho de CPF.")
	}

	prepared.RG = strings.TrimSpace(input.Client.RG)
	prepared.Phone = phoneOrBlank(input.Client.Phone, &prepared.Warnings)
	prepared.Email, prepared.RawEmail = emailOrBlank(input.Client.Email, &prepared.Warnings)
	prepared.RawAddress, prepared.StructuredAddress = addressOf(input.Client)
	prepared.LineDigits = digitsOnly(input.Sale.LineNumber)
	if prepared.LineDigits == "" {
		prepared.Warnings = append(prepared.Warnings, "Venda sem número de linha.")
	} else if len(prepared.LineDigits) < 8 || len(prepared.LineDigits) > 20 {
		prepared.Warnings = append(prepared.Warnings, "Número da linha fora do tamanho aceito. Ele fica guardado na venda e a linha não foi aberta.")
		prepared.LineDigits = ""
	}
	prepared.SimType = simTypeOf(input.Sale.SimType)
	prepared.Iccid = strings.TrimSpace(input.Sale.Iccid)
	prepared.DeviceImei = strings.TrimSpace(input.Sale.DeviceImei)
	prepared.DeviceEid = strings.TrimSpace(input.Sale.DeviceEid)
	prepared.DevicePlatform = strings.ToUpper(strings.TrimSpace(input.Sale.DevicePlatform))
	if prepared.DeviceImei != "" && len(digitsOnly(prepared.DeviceImei)) != 15 {
		prepared.Warnings = append(prepared.Warnings, "IMEI diferente de 15 dígitos. O valor foi guardado como veio.")
	}
	if prepared.DeviceEid != "" && len(digitsOnly(prepared.DeviceEid)) != 32 {
		prepared.Warnings = append(prepared.Warnings, "EID diferente de 32 dígitos. O valor foi guardado como veio.")
	}
	if prepared.DevicePlatform != "" && prepared.DevicePlatform != "IOS" && prepared.DevicePlatform != "ANDROID" {
		prepared.Warnings = append(prepared.Warnings, "Plataforma do aparelho diferente de IOS ou ANDROID. O valor foi guardado como veio.")
	}
	prepared.SoldAt = parseTime(input.Sale.SoldAt, "soldAt", &prepared.Warnings)
	prepared.ConcludedAt = parseTime(input.Sale.ConcludedAt, "concludedAt", &prepared.Warnings)
	if prepared.Warnings == nil {
		prepared.Warnings = []string{}
	}
	return prepared, nil
}

func phoneOrBlank(value string, warnings *[]string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.EqualFold(trimmed, "sem telefone") {
		return ""
	}
	digits := digitsOnly(trimmed)
	if len(digits) < 10 || len(digits) > 15 {
		*warnings = append(*warnings, "Telefone fora de 10 a 15 dígitos. Ficou em branco na ficha e o texto original permanece na venda.")
		return ""
	}
	return digits
}

func emailOrBlank(value string, warnings *[]string) (valid, raw string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", ""
	}
	parsed, err := mail.ParseAddress(trimmed)
	if err != nil || parsed.Address != trimmed || len(trimmed) > 256 {
		*warnings = append(*warnings, "E-mail inválido. Ficou guardado no texto original e não entrou como e-mail de contato.")
		return "", trimmed
	}
	return trimmed, ""
}

func addressOf(client Client) (string, *Address) {
	parts := []string{
		strings.TrimSpace(client.Address),
		strings.TrimSpace(client.AddressNumber),
		strings.TrimSpace(client.Complement),
		strings.TrimSpace(client.Neighborhood),
		strings.TrimSpace(client.City),
		strings.TrimSpace(client.State),
		strings.TrimSpace(client.ZipCode),
	}
	var rawParts []string
	for _, part := range parts {
		if part != "" {
			rawParts = append(rawParts, part)
		}
	}
	raw := strings.Join(rawParts, ", ")
	zip := digitsOnly(client.ZipCode)
	state := strings.ToUpper(strings.TrimSpace(client.State))
	street := strings.TrimSpace(client.Address)
	number := strings.TrimSpace(client.AddressNumber)
	neighborhood := strings.TrimSpace(client.Neighborhood)
	city := strings.TrimSpace(client.City)
	complement := strings.TrimSpace(client.Complement)
	if street == "" || number == "" || neighborhood == "" || city == "" || !validState(state) || len(zip) != 8 ||
		utf8.RuneCountInString(street) > 256 || utf8.RuneCountInString(number) > 16 ||
		utf8.RuneCountInString(neighborhood) > 128 || utf8.RuneCountInString(city) > 128 ||
		utf8.RuneCountInString(complement) > 128 {
		return raw, nil
	}
	return raw, &Address{
		Street: street, Number: number, Neighborhood: neighborhood,
		City: city, State: state, ZipCode: zip, Complement: complement,
	}
}

func simTypeOf(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "CHIP":
		return "PHYSICAL"
	case "ESIM":
		return "ESIM"
	default:
		return "UNKNOWN"
	}
}

func parseTime(value, field string, warnings *[]string) *time.Time {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		parsed, err = time.Parse("2006-01-02", trimmed)
	}
	if err != nil {
		*warnings = append(*warnings, field+" não é uma data válida. A venda entrou sem essa data.")
		return nil
	}
	return &parsed
}

func digitsOnly(value string) string {
	var b strings.Builder
	for _, r := range value {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func DecodeAttachment(contentBase64 string) ([]byte, string, bool) {
	trimmed := strings.TrimSpace(contentBase64)
	if trimmed == "" {
		return nil, "Anexo sem conteúdo.", false
	}
	if i := strings.Index(trimmed, ","); i >= 0 && strings.Contains(strings.ToLower(trimmed[:i]), "base64") {
		trimmed = trimmed[i+1:]
	}
	data, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(trimmed)
	}
	if err != nil {
		return nil, "Anexo com conteúdo que não é base64. O arquivo não entrou.", false
	}
	if len(data) > maxDecodedBytes {
		return nil, "Anexo maior que 12 MB. O arquivo não entrou.", false
	}
	return data, "", true
}

func validState(value string) bool {
	switch value {
	case "AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA", "MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN", "RS", "RO", "RR", "SC", "SP", "SE", "TO":
		return true
	default:
		return false
	}
}
