package vivo

type deserializeFunc func(line string, trim bool) any

func buildDeserializerRegistry() map[string]deserializeFunc {
	return map[string]deserializeFunc{
		"010D": deserializeLine010DHeader,
		"011D": deserializeLine011DCustomer,
		"020D": deserializeLine020DPayment,
		"014D": deserializeLine014DFiscalPlaceholder,
		"015D": deserializeLine015DTariffExcessSummary,
		"016D": deserializeLine016DPreviousPeriodUsage,
		"017J": deserializeLine017JFiscalMetadata,
		"017D": deserializeLine017DFiscalMetadata,
		"050W": deserializeLine050WServicesHeader,
		"050I": deserializeLine050IPlanSummary,
		"050H": deserializeLine050HService,
		"050G": deserializeLine050GService,
		"051W": deserializeLine051WUsageHeader,
		"051D": deserializeLine051DUsage,
		"052W": deserializeLine052WExtraUsageHeader,
		"052E": deserializeLine052EExtraLocation,
		"052D": deserializeLine052DExtraUsageDetail,
		"052C": deserializeLine052CExtraSubtotal,
		"059A": deserializeLine059AExtraTotal,
		"110T": deserializeInvoiceAccountLinesSummary,
		"110D": deserializeLine110DAccountLineDetail,
		"115T": deserializeInvoiceFranchiseSectionHeader,
		"115D": deserializeInvoiceFranchiseLineDetail,
		"565U": deserializeInvoiceFiscalNfcTotals,
		"565S": deserializeInvoiceFiscalNfcItem,
		"565Q": deserializeInvoiceFiscalNfcCustomer,
	}
}

func deserializeLine010DHeader(line string, trim bool) any {
	return &Line010DHeader{
		LineRecord:            bindCommon(line, trim),
		ReferenceMonth:        bindString(line, 205, 6, trim),
		IssueDate:             bindDate(line, 229, 8, "yyyyMMdd", trim),
		DueDate:               bindDate(line, 247, 8, "yyyyMMdd", trim),
		BillingStartDate:      bindDate(line, 353, 8, "yyyyMMdd", trim),
		BillingEndDate:        bindDate(line, 361, 8, "yyyyMMdd", trim),
		SubtotalServices:      bindDecimal(line, 341, 12, trim),
		SubtotalUsageExceeded: bindDecimal(line, 372, 12, trim),
		TotalAmount:           bindDecimal(line, 540, 12, trim),
		FiscalReferenceCode:   bindString(line, 1770, 13, trim),
	}
}

func deserializeLine011DCustomer(line string, trim bool) any {
	return &Line011DCustomer{
		LineRecord:        bindCommon(line, trim),
		Name:              bindString(line, 115, 83, trim),
		LegalName:         bindString(line, 199, 94, trim),
		Document:          bindString(line, 1846, 18, trim),
		Street:            bindString(line, 958, 19, trim),
		Number:            bindString(line, 1090, 44, trim),
		Neighborhood:      bindString(line, 1039, 52, trim),
		ZipCode:           bindString(line, 558, 9, trim),
		City:              bindString(line, 977, 61, trim),
		State:             bindString(line, 1134, 159, trim),
		Country:           bindString(line, 1293, 135, trim),
		StateRegistration: bindString(line, 876, 32, trim),
	}
}

func deserializeLine020DPayment(line string, trim bool) any {
	return &Line020DPayment{
		LineRecord:    bindCommon(line, trim),
		DigitableLine: bindString(line, 178, 51, trim),
		PixQrCode:     bindString(line, 488, 999, trim),
	}
}

func deserializeLine014DFiscalPlaceholder(line string, trim bool) any {
	return &Line014DFiscalPlaceholder{
		LineRecord: bindCommon(line, trim),
		Payload:    bindString(line, 178, 1600, trim),
	}
}

func deserializeLine015DTariffExcessSummary(line string, trim bool) any {
	return &Line015DTariffExcessSummary{
		LineRecord: bindCommon(line, trim),
		RawPayload: bindString(line, 178, 1600, trim),
	}
}

func deserializeLine016DPreviousPeriodUsage(line string, trim bool) any {
	return &Line016DPreviousPeriodUsage{
		LineRecord: bindCommon(line, trim),
		Quantity:   bindDecimal(line, 365, 14, trim),
		Amount:     bindDecimal(line, 379, 9, trim),
		TariffCode: bindString(line, 398, 8, trim),
	}
}

func deserializeLine017JFiscalMetadata(line string, trim bool) any {
	return &Line017JFiscalMetadata{
		LineRecord: bindCommon(line, trim),
		Payload:    bindString(line, 178, 1600, trim),
	}
}

func deserializeLine017DFiscalMetadata(line string, trim bool) any {
	return &Line017DFiscalMetadata{
		LineRecord: bindCommon(line, trim),
		Payload:    bindString(line, 178, 1600, trim),
	}
}

func deserializeLine050WServicesHeader(line string, trim bool) any {
	return &Line050WServicesHeader{
		LineRecord:  bindCommon(line, trim),
		PlanCode:    bindString(line, 159, 18, trim),
		Description: bindString(line, 178, 999, trim),
	}
}

func deserializeLine050IPlanSummary(line string, trim bool) any {
	return &Line050IPlanSummary{
		LineRecord: bindCommon(line, trim),
		PlanCode:   bindString(line, 159, 19, trim),
		PlanName:   bindString(line, 178, 40, trim),
		Subtotal:   bindDecimal(line, 230, 16, trim),
		Flags:      bindString(line, 252, 10, trim),
	}
}

func deserializeLine050HService(line string, trim bool) any {
	return &Line050HService{
		LineRecord:  bindCommon(line, trim),
		PlanCode:    bindString(line, 159, 19, trim),
		ServiceName: bindString(line, 178, 41, trim),
		Flags:       bindString(line, 331, 1, trim),
		Quantity:    bindInt(line, 219, 10, trim),
		Franchise:   bindDecimal(line, 229, 16, trim),
		Used:        bindDecimal(line, 261, 16, trim),
		Unity:       bindString(line, 313, 4, trim),
		Total:       bindDecimal(line, 281, 18, trim),
	}
}

func deserializeLine050GService(line string, trim bool) any {
	return &Line050GService{
		LineRecord:  bindCommon(line, trim),
		PlanCode:    bindString(line, 159, 19, trim),
		ServiceName: bindString(line, 178, 41, trim),
		Flags:       bindString(line, 331, 1, trim),
		Quantity:    bindInt(line, 219, 10, trim),
		Franchise:   bindDecimal(line, 229, 16, trim),
		Used:        bindDecimal(line, 261, 16, trim),
		Unity:       bindString(line, 309, 4, trim),
		Total:       bindDecimal(line, 281, 18, trim),
	}
}

func deserializeLine051WUsageHeader(line string, trim bool) any {
	return &Line051WUsageHeader{
		LineRecord:  bindCommon(line, trim),
		Description: bindString(line, 178, 50, trim),
	}
}

func deserializeLine051DUsage(line string, trim bool) any {
	return &Line051DUsage{
		LineRecord:  bindCommon(line, trim),
		PlanCode:    bindString(line, 159, 19, trim),
		ServiceName: bindString(line, 178, 50, trim),
		Unity:       bindString(line, 260, 4, trim),
		Franchise:   bindDecimal(line, 228, 16, trim),
		Used:        bindDecimal(line, 244, 16, trim),
	}
}

func deserializeLine052WExtraUsageHeader(line string, trim bool) any {
	return &Line052WExtraUsageHeader{
		LineRecord:  bindCommon(line, trim),
		Description: bindString(line, 178, 100, trim),
		FooterFlags: bindString(line, 278, 2, trim),
	}
}

func deserializeLine052EExtraLocation(line string, trim bool) any {
	return &Line052EExtraLocation{
		LineRecord: bindCommon(line, trim),
		Location:   bindString(line, 178, 999, trim),
	}
}

func deserializeLine052DExtraUsageDetail(line string, trim bool) any {
	return &Line052DExtraUsageDetail{
		LineRecord:  bindCommon(line, trim),
		Description: bindString(line, 178, 80, trim),
		Quantity:    bindDecimal(line, 258, 14, trim),
		Amount:      bindDecimal(line, 272, 14, trim),
		ServiceCode: bindString(line, 286, 10, trim),
	}
}

func deserializeLine052CExtraSubtotal(line string, trim bool) any {
	return &Line052CExtraSubtotal{
		LineRecord: bindCommon(line, trim),
		Amount:     bindDecimal(line, 178, 14, trim),
	}
}

func deserializeLine059AExtraTotal(line string, trim bool) any {
	return &Line059AExtraTotal{
		LineRecord:  bindCommon(line, trim),
		TotalAmount: bindDecimal(line, 178, 14, trim),
	}
}

func deserializeInvoiceAccountLinesSummary(line string, trim bool) any {
	rec := bindCommon(line, trim)
	return &InvoiceAccountLinesSummary{
		LineRecord:       rec,
		SubtotalServices: bindDecimal(line, 234, 10, trim),
	}
}

func deserializeLine110DAccountLineDetail(line string, trim bool) any {
	return &Line110DAccountLineDetail{
		LineRecord:   bindCommon(line, trim),
		LineSequence: bindString(line, 64, 6, trim),
		PhoneNumber:  bindString(line, 178, 70, trim),
		PlanName:     bindString(line, 248, 25, trim),
		LineTotal:    bindDecimal(line, 318, 14, trim),
	}
}

func deserializeInvoiceFranchiseSectionHeader(line string, trim bool) any {
	return &InvoiceFranchiseSectionHeader{
		LineRecord:       bindCommon(line, trim),
		SectionName:      bindString(line, 178, 54, trim),
		TotalsPayload:    bindString(line, 232, 138, trim),
		SectionReference: bindString(line, 370, 8, trim),
		SubscriberCount:  bindDecimal(line, 402, 5, trim),
		SectionSequence:  bindString(line, 408, 3, trim),
	}
}

func deserializeInvoiceFranchiseLineDetail(line string, trim bool) any {
	rec := bindCommon(line, trim)
	return &InvoiceFranchiseLineDetail{
		LineRecord:         rec,
		ServiceDescription: bindString(line, 182, 40, trim),
		ServiceOrder:       bindDecimal(line, 181, 1, trim),
		PhoneNumber:        bindString(line, 393, 14, trim),
		DetailSequence:     bindString(line, 408, 3, trim),
		UsageAmount:        bindDecimal(line, 222, 14, trim),
		SectionReference:   bindString(line, 65, 8, trim),
	}
}

func deserializeInvoiceFiscalNfcTotals(line string, trim bool) any {
	return &InvoiceFiscalNfcTotals{
		LineRecord:      bindCommon(line, trim),
		FiscalGroupCode: bindString(line, 64, 4, trim),
		Payload:         bindString(line, 145, 228, trim),
	}
}

func deserializeInvoiceFiscalNfcItem(line string, trim bool) any {
	rec := bindCommon(line, trim)
	return &InvoiceFiscalNfcItem{
		LineRecord:       rec,
		FiscalGroupCode:  bindString(line, 64, 4, trim),
		ItemSubqualifier: bindString(line, 124, 2, trim),
		DetailPayload:    bindString(line, 145, 319, trim),
	}
}

func deserializeInvoiceFiscalNfcCustomer(line string, trim bool) any {
	rec := bindCommon(line, trim)
	return &InvoiceFiscalNfcCustomer{
		LineRecord:      rec,
		FiscalGroupCode: bindString(line, 64, 4, trim),
		Payload:         bindString(line, 145, 551, trim),
	}
}
