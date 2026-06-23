package vivo

import "time"

// Line010DHeader registro 010D — cabeçalho da fatura.
type Line010DHeader struct {
	LineRecord
	ReferenceMonth        string
	IssueDate             time.Time
	DueDate               time.Time
	BillingStartDate      time.Time
	BillingEndDate        time.Time
	SubtotalServices      float64
	SubtotalUsageExceeded float64
	TotalAmount           float64
	FiscalReferenceCode   string
}

// Line011DCustomer registro 011D — dados do cliente.
type Line011DCustomer struct {
	LineRecord
	Name              string
	LegalName         string
	Document          string
	Street            string
	Number            string
	Neighborhood      string
	ZipCode           string
	City              string
	State             string
	Country           string
	StateRegistration string
}

// Line020DPayment registro 020D — dados de pagamento.
type Line020DPayment struct {
	LineRecord
	DigitableLine string
	PixQrCode     string
}

// Line014DFiscalPlaceholder registro 014D.
type Line014DFiscalPlaceholder struct {
	LineRecord
	Payload string
}

// Line015DTariffExcessSummary registro 015D.
type Line015DTariffExcessSummary struct {
	LineRecord
	RawPayload string
}

// Line016DPreviousPeriodUsage registro 016D.
type Line016DPreviousPeriodUsage struct {
	LineRecord
	Quantity   float64
	Amount     float64
	TariffCode string
}

// Line017JFiscalMetadata registro 017J.
type Line017JFiscalMetadata struct {
	LineRecord
	Payload string
}

// Line017DFiscalMetadata registro 017D.
type Line017DFiscalMetadata struct {
	LineRecord
	Payload string
}

// Line050WServicesHeader registro 050W.
type Line050WServicesHeader struct {
	LineRecord
	PlanCode    string
	Description string
}

// Line050IPlanSummary registro 050I.
type Line050IPlanSummary struct {
	LineRecord
	PlanCode string
	PlanName string
	Subtotal float64
	Flags    string
}

// Line050HService registro 050H.
type Line050HService struct {
	LineRecord
	PlanCode    string
	ServiceName string
	Flags       string
	Quantity    int
	Franchise   float64
	Used        float64
	Unity       string
	Total       float64
}

// Line050GService registro 050G.
type Line050GService struct {
	LineRecord
	PlanCode    string
	ServiceName string
	Flags       string
	Quantity    int
	Franchise   float64
	Used        float64
	Unity       string
	Total       float64
}

// Line051WUsageHeader registro 051W.
type Line051WUsageHeader struct {
	LineRecord
	Description string
}

// Line051DUsage registro 051D.
type Line051DUsage struct {
	LineRecord
	PlanCode    string
	ServiceName string
	Unity       string
	Franchise   float64
	Used        float64
}

// Line052WExtraUsageHeader registro 052W.
type Line052WExtraUsageHeader struct {
	LineRecord
	Description  string
	FooterFlags  string
}

// Line052EExtraLocation registro 052E.
type Line052EExtraLocation struct {
	LineRecord
	Location string
}

// Line052DExtraUsageDetail registro 052D.
type Line052DExtraUsageDetail struct {
	LineRecord
	Description string
	Quantity    float64
	Amount      float64
	ServiceCode string
}

// Line052CExtraSubtotal registro 052C.
type Line052CExtraSubtotal struct {
	LineRecord
	Amount float64
}

// Line059AExtraTotal registro 059A.
type Line059AExtraTotal struct {
	LineRecord
	TotalAmount float64
}

// InvoiceAccountLinesSummary registro 110T.
type InvoiceAccountLinesSummary struct {
	LineRecord
	SubtotalServices float64
}

// Line110DAccountLineDetail registro 110D — detalhe de linha telefônica.
type Line110DAccountLineDetail struct {
	LineRecord
	LineSequence string
	PhoneNumber  string
	PlanName     string
	LineTotal    float64
}

// InvoiceFranchiseSectionHeader registro 115T.
type InvoiceFranchiseSectionHeader struct {
	LineRecord
	SectionName      string
	TotalsPayload    string
	SectionReference string
	SubscriberCount  float64
	SectionSequence  string
}

// InvoiceFranchiseLineDetail registro 115D.
type InvoiceFranchiseLineDetail struct {
	LineRecord
	ServiceDescription string
	ServiceOrder       float64
	PhoneNumber        string
	DetailSequence     string
	UsageAmount        float64
	SectionReference   string
}

// InvoiceFiscalNfcTotals registro 565U.
type InvoiceFiscalNfcTotals struct {
	LineRecord
	FiscalGroupCode string
	Payload         string
}

// InvoiceFiscalNfcItem registro 565S.
type InvoiceFiscalNfcItem struct {
	LineRecord
	FiscalGroupCode  string
	ItemSubqualifier string
	DetailPayload    string
}

// InvoiceFiscalNfcCustomer registro 565Q.
type InvoiceFiscalNfcCustomer struct {
	LineRecord
	FiscalGroupCode string
	Payload         string
}
