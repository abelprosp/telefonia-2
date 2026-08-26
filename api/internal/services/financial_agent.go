package services

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/precision"
	"github.com/luxus-connect/telefonia/api/internal/sicredi"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) FinancialAgentHealth(ctx context.Context) (*models.FinancialAgentHealthResponse, error) {
	resp := &models.FinancialAgentHealthResponse{
		Enabled:         true,
		OrganizationSet: true,
		Message:         "Agente financeiro pronto.",
	}
	if s.Sicredi != nil && s.Sicredi.Enabled() {
		resp.SicrediEnabled = true
		if err := s.Sicredi.Ping(ctx); err != nil {
			resp.Message = "Agente pronto. Sicredi habilitado, porém a conexão falhou: " + err.Error()
		} else {
			resp.SicrediConnected = true
		}
	} else {
		resp.Message = "Agente pronto. Integração Sicredi desabilitada."
	}
	return resp, nil
}

func (s *Service) FinancialAgentLookupCustomer(ctx context.Context, input models.FinancialAgentLookupInput) (*models.FinancialAgentLookupResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	query := strings.TrimSpace(firstNonEmpty(input.Query, input.Name, input.Document, input.Phone, input.LineNumber, input.InvoiceNumber, input.WhatsAppNumber))
	if query == "" {
		return nil, httputil.ValidationError(notifications.FinancialAgentQueryRequired)
	}
	rows, err := s.Store.SearchAgentCustomers(ctx, orgID, input)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	asOf := time.Now().UTC()
	items := make([]models.FinancialAgentCustomer, 0, len(rows))
	for _, row := range rows {
		cust, err := s.agentCustomerFromRow(ctx, orgID, row, asOf)
		if err != nil {
			return nil, err
		}
		items = append(items, cust)
	}
	msg := ""
	switch len(items) {
	case 0:
		msg = "Nenhum cliente encontrado. Peça nome completo, CPF/CNPJ, número da linha ou da fatura."
	case 1:
		msg = "Cliente localizado."
	default:
		msg = "Vários clientes encontrados. Confirme o nome ou o CPF/CNPJ."
	}
	return &models.FinancialAgentLookupResponse{Query: query, Count: len(items), Items: items, Message: msg}, nil
}

func (s *Service) agentCustomerFromRow(ctx context.Context, orgID string, row store.AgentCustomerRow, asOf time.Time) (models.FinancialAgentCustomer, error) {
	lines, err := s.Store.ListAgentCustomerLineNumbers(ctx, orgID, row.ID)
	if err != nil {
		return models.FinancialAgentCustomer{}, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if lines == nil {
		lines = []string{}
	}
	openCount, overdueCount, overdueBalance, err := s.Store.CountAgentCustomerInvoiceStats(ctx, orgID, row.ID, asOf)
	if err != nil {
		return models.FinancialAgentCustomer{}, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return models.FinancialAgentCustomer{
		ID:               row.ID,
		Name:             row.Name,
		LegalName:        row.LegalName,
		CpfCnpj:          row.CpfCnpj,
		Active:           row.Active,
		BillingEmail:     row.BillingEmail,
		PhoneLineNumbers: lines,
		MatchReason:      row.MatchReason,
		OpenInvoiceCount: openCount,
		OverdueCount:     overdueCount,
		OverdueBalance:   overdueBalance,
	}, nil
}

func (s *Service) FinancialAgentListInvoices(ctx context.Context, input models.FinancialAgentListInvoicesInput) (*models.FinancialAgentListInvoicesResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	customerID := strings.TrimSpace(input.CustomerID)
	if customerID == "" {
		lookup, err := s.FinancialAgentLookupCustomer(ctx, models.FinancialAgentLookupInput{
			Query:          firstNonEmpty(input.Query, input.InvoiceNumber, input.WhatsAppNumber),
			WhatsAppNumber: input.WhatsAppNumber,
			InvoiceNumber:  input.InvoiceNumber,
		})
		if err != nil {
			return nil, err
		}
		if lookup.Count == 0 {
			return &models.FinancialAgentListInvoicesResponse{Items: []models.FinancialAgentInvoice{}}, nil
		}
		if lookup.Count == 1 {
			customerID = lookup.Items[0].ID
		} else if strings.TrimSpace(input.InvoiceNumber) == "" {
			return &models.FinancialAgentListInvoicesResponse{Items: []models.FinancialAgentInvoice{}}, nil
		}
	}
	var due *time.Time
	if strings.TrimSpace(input.DueDate) != "" {
		parsed, err := parseAgentDate(input.DueDate)
		if err != nil {
			return nil, httputil.ValidationError(notifications.N("REQUEST_VALIDATION", "Data de vencimento inválida. Use AAAA-MM-DD ou DD/MM/AAAA."))
		}
		due = &parsed
	}
	asOf := time.Now().UTC()
	docs, err := s.Store.ListAgentInvoices(ctx, orgID, customerID, strings.TrimSpace(firstNonEmpty(input.InvoiceNumber, "")), strings.TrimSpace(input.Status), due, input.OverdueOnly, asOf, 30)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	items := make([]models.FinancialAgentInvoice, 0, len(docs))
	for _, doc := range docs {
		inv := toAgentInvoice(doc, asOf)
		if customerID == "" && strings.TrimSpace(input.Query) != "" {
			q := strings.ToLower(strings.TrimSpace(input.Query))
			if !strings.Contains(strings.ToLower(inv.CustomerName), q) &&
				!strings.Contains(strings.ToLower(inv.InvoiceNumber), q) {
				continue
			}
		}
		items = append(items, inv)
	}
	return &models.FinancialAgentListInvoicesResponse{Count: len(items), Items: items}, nil
}

func (s *Service) FinancialAgentInvoicePackage(ctx context.Context, input models.FinancialAgentInvoicePackageInput) (*models.FinancialAgentInvoicePackage, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := s.resolveAgentInvoice(ctx, orgID, input.InvoiceID, input.CustomerID, input.Query, input.WhatsAppNumber)
	if err != nil {
		return nil, err
	}
	asOf := time.Now().UTC()
	inv := toAgentInvoice(doc.ListCustomerBillingDocumentResponse, asOf)
	settings, _ := s.Store.GetOrganizationSettings(ctx, orgID)
	company := "Luxus Connect"
	if settings != nil {
		if n := strings.TrimSpace(settings.Company.TradingName); n != "" {
			company = n
		} else if n := strings.TrimSpace(settings.Company.CompanyName); n != "" {
			company = n
		}
	}
	pix := derefStr(inv.SicrediPixQrCode)
	linha := derefStr(inv.SicrediLinhaDigitavel)
	charges := CalculateLateCharges(inv.Amount, inv.DueDate, asOf, agentLateFee(settings), agentInterestMonthly(settings))
	var b strings.Builder
	fmt.Fprintf(&b, "Olá, %s! Aqui é o atendimento financeiro da %s.\n\n", inv.CustomerName, company)
	fmt.Fprintf(&b, "Fatura: %s\nVencimento: %s\nValor original: %s\n", inv.InvoiceNumber, inv.DueDate.Format("02/01/2006"), formatMoneyBR(inv.Amount))
	if charges.Overdue {
		fmt.Fprintf(&b, "Em atraso: %d dia(s)\nMulta: %s\nJuros: %s\nTotal atualizado: %s\n",
			charges.DaysOverdue, formatMoneyBR(charges.LateFeeAmount), formatMoneyBR(charges.InterestAmount), formatMoneyBR(charges.TotalAmount))
	}
	if inv.Paid {
		b.WriteString("\nEsta fatura já consta como paga.\n")
	} else {
		if pix != "" {
			b.WriteString("\nPIX copia e cola:\n")
			b.WriteString(pix)
			b.WriteString("\n")
		}
		if linha != "" {
			b.WriteString("\nLinha digitável do boleto:\n")
			b.WriteString(linha)
			b.WriteString("\n")
		}
		if inv.HasBoletoPDF {
			b.WriteString("\nEm seguida envio o PDF do boleto Sicredi.\n")
		}
	}
	base := strings.TrimRight(s.FinancialAgentPublicURL, "/")
	pkg := &models.FinancialAgentInvoicePackage{
		Invoice:        inv,
		WhatsAppText:   strings.TrimSpace(b.String()),
		PixCopiaCola:   pix,
		LinhaDigitavel: linha,
		HasBoletoPDF:   inv.HasBoletoPDF,
		CompanyName:    company,
	}
	if base != "" {
		pkg.InvoiceHTMLURL = base + "/v1/agent/financial/invoices/" + inv.ID + "/download"
		if inv.HasBoletoPDF {
			pkg.BoletoPDFURL = base + "/v1/agent/financial/invoices/" + inv.ID + "/boleto-pdf"
		}
	}
	return pkg, nil
}

func (s *Service) FinancialAgentCheckDelinquency(ctx context.Context, input models.FinancialAgentDelinquencyInput) (*models.FinancialAgentDelinquencyResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	lookup, err := s.FinancialAgentLookupCustomer(ctx, models.FinancialAgentLookupInput{
		Query:          firstNonEmpty(input.Query, input.CustomerID, input.WhatsAppNumber),
		WhatsAppNumber: input.WhatsAppNumber,
	})
	if err != nil {
		return nil, err
	}
	if lookup.Count == 0 {
		return &models.FinancialAgentDelinquencyResponse{
			Summary: "Não encontrei o cliente para consultar inadimplência.",
			Items:   []models.FinancialAgentDelinquencyInvoice{},
		}, nil
	}
	if lookup.Count > 1 && strings.TrimSpace(input.CustomerID) == "" {
		return &models.FinancialAgentDelinquencyResponse{
			Summary: "Encontrei mais de um cliente. Confirme o nome ou CPF/CNPJ.",
			Items:   []models.FinancialAgentDelinquencyInvoice{},
		}, nil
	}
	cust := lookup.Items[0]
	if id := strings.TrimSpace(input.CustomerID); id != "" {
		for i := range lookup.Items {
			if lookup.Items[i].ID == id {
				cust = lookup.Items[i]
				break
			}
		}
	}
	settings, _ := s.Store.GetOrganizationSettings(ctx, orgID)
	asOf := time.Now().UTC()
	docs, err := s.Store.ListAgentInvoices(ctx, orgID, cust.ID, "", "", nil, false, asOf, 40)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	sicrediOK := s.Sicredi != nil && s.Sicredi.Enabled()
	syncNote := ""
	out := &models.FinancialAgentDelinquencyResponse{
		Customer:         &cust,
		SicrediConnected: sicrediOK,
		Items:            []models.FinancialAgentDelinquencyInvoice{},
	}
	for _, doc := range docs {
		inv := toAgentInvoice(doc, asOf)
		if inv.Paid || inv.Status == "cancelled" {
			continue
		}
		charges := CalculateLateCharges(inv.Amount, inv.DueDate, asOf, agentLateFee(settings), agentInterestMonthly(settings))
		item := models.FinancialAgentDelinquencyInvoice{
			FinancialAgentInvoice: inv,
			LateFeeAmount:         charges.LateFeeAmount,
			InterestAmount:        charges.InterestAmount,
			TotalWithCharges:      charges.TotalAmount,
			SystemStatus:          inv.Status,
			SicrediStatus:         derefStr(inv.SicrediBoletoStatus),
			Delinquent:            inv.Overdue && !inv.Paid,
		}
		if input.SyncSicredi && sicrediOK && derefStr(inv.SicrediNossoNumero) != "" {
			detail, err := s.Sicredi.GetBoleto(ctx, strings.TrimSpace(*inv.SicrediNossoNumero))
			if err != nil {
				syncNote = "Falha ao consultar o Sicredi: " + err.Error()
			} else if detail != nil {
				item.SicrediSituacao = detail.Situacao
				item.SicrediStatus = detail.Situacao
				if sicredi.IsSituacaoLiquidada(detail.Situacao) {
					item.Paid = true
					item.Delinquent = false
					item.Overdue = false
					_, _ = s.SyncSicrediPaymentForDocument(ctx, inv.ID)
				}
			}
		}
		if item.Delinquent {
			out.Delinquent = true
			out.OverdueCount++
			out.OverdueBalance = precision.Round2(out.OverdueBalance + inv.Amount)
			out.TotalWithCharges = precision.Round2(out.TotalWithCharges + item.TotalWithCharges)
			out.Items = append(out.Items, item)
		}
	}
	out.SicrediSyncNote = syncNote
	if !out.Delinquent {
		out.Summary = fmt.Sprintf("%s está em dia. Não há faturas vencidas em aberto no sistema.", cust.Name)
	} else {
		out.Summary = fmt.Sprintf("%s está inadimplente: %d fatura(s) vencida(s), saldo original %s, total com multa e juros %s.",
			cust.Name, out.OverdueCount, formatMoneyBR(out.OverdueBalance), formatMoneyBR(out.TotalWithCharges))
	}
	return out, nil
}

func (s *Service) FinancialAgentCalculateInterest(ctx context.Context, input models.FinancialAgentInterestInput) (*models.FinancialAgentInterestResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := s.Store.GetOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	asOf := time.Now().UTC()
	if strings.TrimSpace(input.AsOf) != "" {
		parsed, err := parseAgentDate(input.AsOf)
		if err != nil {
			return nil, httputil.ValidationError(notifications.N("REQUEST_VALIDATION", "Data de referência inválida."))
		}
		asOf = parsed
	}
	amount := input.Amount
	due := time.Time{}
	invoiceID := strings.TrimSpace(input.InvoiceID)
	invoiceNumber := ""
	if invoiceID != "" {
		doc, err := s.GetCustomerBillingDocument(ctx, invoiceID)
		if err != nil {
			return nil, err
		}
		amount = doc.Amount
		due = doc.DueDate
		invoiceNumber = doc.InvoiceNumber
	} else {
		if amount <= 0 || strings.TrimSpace(input.DueDate) == "" {
			return nil, httputil.ValidationError(notifications.N("REQUEST_VALIDATION", "Informe invoice_id ou amount + due_date."))
		}
		parsed, err := parseAgentDate(input.DueDate)
		if err != nil {
			return nil, httputil.ValidationError(notifications.N("REQUEST_VALIDATION", "Data de vencimento inválida."))
		}
		due = parsed
	}
	charges := CalculateLateCharges(amount, due, asOf, agentLateFee(settings), agentInterestMonthly(settings))
	resp := charges
	resp.InvoiceID = invoiceID
	resp.InvoiceNumber = invoiceNumber
	return &resp, nil
}

func (s *Service) FinancialAgentSyncSicredi(ctx context.Context, input models.FinancialAgentSyncSicrediInput) (*models.SyncSicrediPaymentsResponse, error) {
	if id := strings.TrimSpace(input.InvoiceID); id != "" {
		return s.SyncSicrediPaymentForDocument(ctx, id)
	}
	if cid := strings.TrimSpace(input.CustomerID); cid != "" {
		docs, err := s.Store.ListAgentInvoices(ctx, mustOrg(ctx), cid, "", "", nil, false, time.Now().UTC(), 40)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		agg := &models.SyncSicrediPaymentsResponse{Items: []models.SyncSicrediPaymentItemResult{}}
		for _, doc := range docs {
			if doc.SicrediNossoNumero == nil || strings.TrimSpace(*doc.SicrediNossoNumero) == "" {
				continue
			}
			one, err := s.SyncSicrediPaymentForDocument(ctx, doc.ID)
			if err != nil {
				agg.Items = append(agg.Items, models.SyncSicrediPaymentItemResult{
					DocumentID:    doc.ID,
					InvoiceNumber: doc.InvoiceNumber,
					CustomerName:  doc.CustomerName,
					Status:        "error",
					Message:       err.Error(),
				})
				agg.Checked++
				continue
			}
			agg.Checked += one.Checked
			agg.Paid += one.Paid
			agg.Items = append(agg.Items, one.Items...)
		}
		return agg, nil
	}
	return s.SyncSicrediPayments(ctx, 7)
}

func (s *Service) FinancialAgentVerifyReceipt(ctx context.Context, input models.FinancialAgentVerifyReceiptInput) (*models.FinancialAgentVerifyReceiptResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if input.Amount <= 0 && strings.TrimSpace(input.PixTxID) == "" && strings.TrimSpace(input.PixEndToEnd) == "" &&
		strings.TrimSpace(input.InvoiceID) == "" && strings.TrimSpace(input.InvoiceNumber) == "" &&
		strings.TrimSpace(input.NossoNumero) == "" && strings.TrimSpace(input.ReceiptText) == "" {
		return nil, httputil.ValidationError(notifications.FinancialAgentReceiptRequired)
	}
	asOf := time.Now().UTC()
	var paidAt *time.Time
	if strings.TrimSpace(input.PaidAt) != "" {
		if parsed, err := parseAgentDate(input.PaidAt); err == nil {
			paidAt = &parsed
			asOf = parsed
		}
	}

	var target *models.ListCustomerBillingDocumentResponse
	if id := strings.TrimSpace(input.InvoiceID); id != "" {
		doc, err := s.GetCustomerBillingDocument(ctx, id)
		if err != nil {
			return nil, err
		}
		target = &doc.ListCustomerBillingDocumentResponse
	} else {
		customerID := strings.TrimSpace(input.CustomerID)
		if customerID == "" {
			lookup, err := s.FinancialAgentLookupCustomer(ctx, models.FinancialAgentLookupInput{
				Query:          firstNonEmpty(input.Query, input.InvoiceNumber, input.PayerName, input.WhatsAppNumber, input.NossoNumero),
				WhatsAppNumber: input.WhatsAppNumber,
				InvoiceNumber:  input.InvoiceNumber,
				Name:           input.PayerName,
			})
			if err == nil && lookup.Count == 1 {
				customerID = lookup.Items[0].ID
			}
		}
		docs, err := s.Store.ListAgentInvoices(ctx, orgID, customerID, strings.TrimSpace(input.InvoiceNumber), "", nil, false, time.Now().UTC(), 40)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		bestScore := -1
		for i := range docs {
			inv := toAgentInvoice(docs[i], time.Now().UTC())
			score, _ := scoreReceiptMatch(input, inv, paidAt)
			if score > bestScore {
				bestScore = score
				d := docs[i]
				target = &d
			}
		}
	}
	if target == nil {
		return &models.FinancialAgentVerifyReceiptResponse{
			ConfidenceLabel:  "none",
			NeedsHumanReview: true,
			Reasons:          []string{"Não encontrei fatura correspondente no sistema."},
			Summary:          "Não foi possível casar o comprovante com uma fatura. Encaminhe para o financeiro.",
		}, nil
	}

	inv := toAgentInvoice(*target, time.Now().UTC())
	settings, _ := s.Store.GetOrganizationSettings(ctx, orgID)
	interest := CalculateLateCharges(inv.Amount, inv.DueDate, asOf, agentLateFee(settings), agentInterestMonthly(settings))
	score, reasons := scoreReceiptMatch(input, inv, paidAt)
	if amountsClose(input.Amount, interest.TotalAmount) && input.Amount > 0 && !amountsClose(input.Amount, inv.Amount) {
		score += 15
		reasons = append(reasons, "Valor do comprovante confere com o total atualizado (multa + juros).")
	}

	sicrediLiquidated := false
	sicrediSituacao := ""
	if s.Sicredi != nil && s.Sicredi.Enabled() && derefStr(inv.SicrediNossoNumero) != "" {
		if detail, err := s.Sicredi.GetBoleto(ctx, strings.TrimSpace(*inv.SicrediNossoNumero)); err == nil && detail != nil {
			sicrediSituacao = detail.Situacao
			sicrediLiquidated = sicredi.IsSituacaoLiquidada(detail.Situacao)
			if sicrediLiquidated {
				score += 25
				reasons = append(reasons, "Sicredi confirma boleto liquidado.")
			}
		}
	}

	label := "low"
	if score >= 70 {
		label = "high"
	} else if score >= 45 {
		label = "medium"
	}
	pixMatch := pixIDsMatch(input, inv)
	amountOK := input.Amount > 0 && (amountsClose(input.Amount, inv.Amount) || amountsClose(input.Amount, interest.TotalAmount))
	canAuto := score >= 70 && (sicrediLiquidated || pixMatch) && amountOK
	if inv.Paid {
		canAuto = false
		reasons = append(reasons, "Fatura já está paga no sistema.")
	}

	resp := &models.FinancialAgentVerifyReceiptResponse{
		Matched:           score >= 45,
		Confidence:        score,
		ConfidenceLabel:   label,
		CanAutoConfirm:    canAuto,
		NeedsHumanReview:  !canAuto,
		Reasons:           reasons,
		Invoice:           &inv,
		Interest:          &interest,
		SicrediSituacao:   sicrediSituacao,
		SicrediLiquidated: sicrediLiquidated,
	}

	if inv.Paid {
		resp.Summary = fmt.Sprintf("Comprovante parece da fatura %s, que já está paga.", inv.InvoiceNumber)
		return resp, nil
	}

	if canAuto {
		if _, err := s.confirmAgentPayment(ctx, orgID, inv, input.Amount, paidAt, receiptReference(input), "Comprovante validado pelo agente financeiro", sicrediLiquidated); err != nil {
			resp.NeedsHumanReview = true
			resp.Summary = "Comprovante conferido, mas a baixa automática falhou: " + err.Error()
			return resp, nil
		}
		resp.Confirmed = true
		resp.NeedsHumanReview = false
		resp.Summary = fmt.Sprintf("Comprovante conferido e pagamento da fatura %s baixado automaticamente.", inv.InvoiceNumber)
		return resp, nil
	}

	if input.ForceConfirm && score >= 45 && amountOK {
		if _, err := s.confirmAgentPayment(ctx, orgID, inv, input.Amount, paidAt, receiptReference(input), "Baixa forçada pelo agente após análise do comprovante", false); err != nil {
			return nil, err
		}
		resp.Confirmed = true
		resp.CanAutoConfirm = true
		resp.NeedsHumanReview = false
		resp.Summary = fmt.Sprintf("Pagamento da fatura %s baixado por confirmação explícita do operador.", inv.InvoiceNumber)
		return resp, nil
	}

	resp.Summary = fmt.Sprintf("Comprovante com confiança %s (%d). Fatura candidata %s, valor original %s, total com encargos %s. Não baixar automaticamente — revisão humana.",
		label, score, inv.InvoiceNumber, formatMoneyBR(inv.Amount), formatMoneyBR(interest.TotalAmount))
	return resp, nil
}

func (s *Service) FinancialAgentConfirmPayment(ctx context.Context, input models.FinancialAgentConfirmPaymentInput) (*models.FinancialAgentConfirmPaymentResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	id := strings.TrimSpace(input.InvoiceID)
	if id == "" {
		return nil, httputil.ValidationError(notifications.N("REQUEST_VALIDATION", "invoice_id é obrigatório."))
	}
	doc, err := s.GetCustomerBillingDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	inv := toAgentInvoice(doc.ListCustomerBillingDocumentResponse, time.Now().UTC())
	if inv.Paid && !input.Force {
		return &models.FinancialAgentConfirmPaymentResponse{
			Success: true, Message: "Fatura já estava paga.", InvoiceID: inv.ID, InvoiceNumber: inv.InvoiceNumber, Amount: inv.Amount, Source: "already_paid",
		}, nil
	}

	if s.Sicredi != nil && s.Sicredi.Enabled() && derefStr(inv.SicrediNossoNumero) != "" {
		synced, err := s.SyncSicrediPaymentForDocument(ctx, inv.ID)
		if err == nil && synced != nil && synced.Paid > 0 {
			return &models.FinancialAgentConfirmPaymentResponse{
				Success: true, Message: "Pagamento confirmado no Sicredi e baixado no sistema.",
				InvoiceID: inv.ID, InvoiceNumber: inv.InvoiceNumber, Amount: inv.Amount, Source: "sicredi",
			}, nil
		}
	}

	if !input.Force {
		return nil, httputil.BusinessError(notifications.FinancialAgentConfirmDenied)
	}
	var paidAt *time.Time
	if strings.TrimSpace(input.PaymentDate) != "" {
		if parsed, err := parseAgentDate(input.PaymentDate); err == nil {
			paidAt = &parsed
		}
	}
	amount := input.Amount
	if amount <= 0 {
		amount = inv.Amount
	}
	ref := strings.TrimSpace(input.Reference)
	if ref == "" {
		ref = "Agente financeiro"
	}
	notes := strings.TrimSpace(input.Notes)
	if notes == "" {
		notes = "Baixa manual via agente financeiro (force=true)."
	}
	if _, err := s.confirmAgentPayment(ctx, orgID, inv, amount, paidAt, ref, notes, false); err != nil {
		return nil, err
	}
	return &models.FinancialAgentConfirmPaymentResponse{
		Success: true, Message: "Pagamento baixado no sistema.", InvoiceID: inv.ID, InvoiceNumber: inv.InvoiceNumber, Amount: amount, Source: "manual",
	}, nil
}

func (s *Service) confirmAgentPayment(ctx context.Context, orgID string, inv models.FinancialAgentInvoice, amount float64, paidAt *time.Time, reference, notes string, fromSicredi bool) (*models.FinancialAgentConfirmPaymentResponse, error) {
	when := time.Now().UTC()
	if paidAt != nil && !paidAt.IsZero() {
		when = paidAt.UTC()
	}
	if amount <= 0 {
		amount = inv.Amount
	}
	if fromSicredi {
		if _, err := s.SyncSicrediPaymentForDocument(ctx, inv.ID); err != nil {
			return nil, err
		}
	} else if inv.AccountsReceivableID != nil && strings.TrimSpace(*inv.AccountsReceivableID) != "" {
		receivableID := strings.TrimSpace(*inv.AccountsReceivableID)
		ref := strings.TrimSpace(reference)
		if ref == "" {
			ref = "Agente financeiro " + inv.InvoiceNumber
		}
		exists, err := s.Store.ReceivablePaymentExistsByReference(ctx, orgID, receivableID, ref)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if !exists {
			if err := s.Store.RegisterReceivablePaymentAuto(ctx, uuid.New().String(), orgID, receivableID, amount, when, ref, notes, time.Now().UTC()); err != nil {
				return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
			}
		}
		_ = s.Store.MarkSicrediBoletoPaid(ctx, orgID, inv.ID, when)
	} else {
		_ = s.Store.MarkSicrediBoletoPaid(ctx, orgID, inv.ID, when)
	}
	return &models.FinancialAgentConfirmPaymentResponse{
		Success: true, InvoiceID: inv.ID, InvoiceNumber: inv.InvoiceNumber, Amount: amount,
	}, nil
}

func (s *Service) resolveAgentInvoice(ctx context.Context, orgID, invoiceID, customerID, query, whatsapp string) (*models.GetCustomerBillingDocumentResponse, error) {
	if id := strings.TrimSpace(invoiceID); id != "" {
		return s.GetCustomerBillingDocument(ctx, id)
	}
	list, err := s.FinancialAgentListInvoices(ctx, models.FinancialAgentListInvoicesInput{
		CustomerID:     customerID,
		Query:          query,
		WhatsAppNumber: whatsapp,
	})
	if err != nil {
		return nil, err
	}
	if list.Count == 0 {
		return nil, httputil.NotFoundError(notifications.BillingDocumentNotFound)
	}
	open := make([]models.FinancialAgentInvoice, 0)
	for _, inv := range list.Items {
		if !inv.Paid && inv.Status != "cancelled" {
			open = append(open, inv)
		}
	}
	if len(open) == 0 {
		open = list.Items
	}
	return s.GetCustomerBillingDocument(ctx, open[0].ID)
}

func toAgentInvoice(doc models.ListCustomerBillingDocumentResponse, asOf time.Time) models.FinancialAgentInvoice {
	paid := doc.SicrediPaidAt != nil || strings.EqualFold(derefStr(doc.SicrediBoletoStatus), "paid")
	days := 0
	overdue := false
	if !paid && doc.Status != "cancelled" {
		due := time.Date(doc.DueDate.Year(), doc.DueDate.Month(), doc.DueDate.Day(), 0, 0, 0, 0, time.UTC)
		as := time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 0, 0, 0, 0, time.UTC)
		d := int(as.Sub(due).Hours() / 24)
		if d > 0 {
			days = d
			overdue = true
		}
	}
	return models.FinancialAgentInvoice{
		ID:                    doc.ID,
		CustomerID:            doc.CustomerID,
		CustomerName:          doc.CustomerName,
		InvoiceNumber:         doc.InvoiceNumber,
		IssueDate:             doc.IssueDate,
		DueDate:               doc.DueDate,
		Amount:                doc.Amount,
		Status:                doc.Status,
		Overdue:               overdue,
		DaysOverdue:           days,
		PhoneLineNumber:       doc.PhoneLineNumber,
		AccountsReceivableID:  doc.AccountsReceivableID,
		SicrediNossoNumero:    doc.SicrediNossoNumero,
		SicrediLinhaDigitavel: doc.SicrediLinhaDigitavel,
		SicrediPixQrCode:      doc.SicrediPixQrCode,
		SicrediPixTxID:        doc.SicrediPixTxID,
		SicrediBoletoStatus:   doc.SicrediBoletoStatus,
		SicrediPaidAt:         doc.SicrediPaidAt,
		HasBoletoPDF:          strings.TrimSpace(derefStr(doc.SicrediLinhaDigitavel)) != "",
		Paid:                  paid,
	}
}

func CalculateLateCharges(amount float64, dueDate, asOf time.Time, lateFeePct, interestMonthly float64) models.FinancialAgentInterestResponse {
	due := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, time.UTC)
	as := time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 0, 0, 0, 0, time.UTC)
	days := int(as.Sub(due).Hours() / 24)
	if days < 0 {
		days = 0
	}
	overdue := days > 0
	var multa, juros float64
	if overdue {
		multa = precision.Round2(amount * lateFeePct / 100)
		juros = precision.Round2(amount * (interestMonthly / 100) / 30 * float64(days))
	}
	total := precision.Round2(amount + multa + juros)
	explain := "Fatura em dia: sem multa nem juros."
	if overdue {
		explain = fmt.Sprintf("Atraso de %d dia(s). Multa de %.2f%% = %s. Juros de %.2f%% a.m. pro rata (÷30) = %s. Total %s.",
			days, lateFeePct, formatMoneyBR(multa), interestMonthly, formatMoneyBR(juros), formatMoneyBR(total))
	}
	return models.FinancialAgentInterestResponse{
		OriginalAmount:      precision.Round2(amount),
		DueDate:             due,
		AsOf:                as,
		DaysOverdue:         days,
		Overdue:             overdue,
		LateFeePercentage:   lateFeePct,
		InterestRateMonthly: interestMonthly,
		LateFeeAmount:       multa,
		InterestAmount:      juros,
		TotalAmount:         total,
		Explanation:         explain,
	}
}

func scoreReceiptMatch(input models.FinancialAgentVerifyReceiptInput, inv models.FinancialAgentInvoice, paidAt *time.Time) (int, []string) {
	score := 0
	var reasons []string
	receipt := strings.ToLower(input.ReceiptText + " " + input.InvoiceNumber + " " + input.NossoNumero + " " + input.LinhaDigitavel)
	if pixIDsMatch(input, inv) {
		score += 40
		reasons = append(reasons, "Identificador PIX confere com a fatura.")
	}
	if n := onlyDigitsRunes(derefStr(inv.SicrediNossoNumero)); n != "" && (strings.Contains(onlyDigitsRunes(input.NossoNumero), n) || strings.Contains(onlyDigitsRunes(receipt), n)) {
		score += 30
		reasons = append(reasons, "Nosso número Sicredi encontrado no comprovante.")
	}
	if l := onlyDigitsRunes(derefStr(inv.SicrediLinhaDigitavel)); len(l) >= 12 && (strings.Contains(onlyDigitsRunes(input.LinhaDigitavel), l) || strings.Contains(onlyDigitsRunes(receipt), l)) {
		score += 30
		reasons = append(reasons, "Linha digitável encontrada no comprovante.")
	}
	if inv.InvoiceNumber != "" && strings.Contains(strings.ToLower(receipt+" "+input.InvoiceNumber), strings.ToLower(inv.InvoiceNumber)) {
		score += 20
		reasons = append(reasons, "Número da fatura citado no comprovante.")
	}
	if input.Amount > 0 && amountsClose(input.Amount, inv.Amount) {
		score += 40
		reasons = append(reasons, "Valor do comprovante igual ao valor original da fatura.")
	}
	if paidAt != nil {
		delta := paidAt.Sub(inv.DueDate)
		if delta < 0 {
			delta = -delta
		}
		if delta <= 15*24*time.Hour {
			score += 15
			reasons = append(reasons, "Data do pagamento próxima do vencimento.")
		}
	}
	if name := strings.TrimSpace(input.PayerName); name != "" {
		if strings.Contains(strings.ToLower(inv.CustomerName), strings.ToLower(name)) ||
			strings.Contains(strings.ToLower(name), strings.ToLower(inv.CustomerName)) {
			score += 10
			reasons = append(reasons, "Nome do pagador compatível com o cliente.")
		}
	}
	if score > 100 {
		score = 100
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "Poucos sinais de correspondência entre comprovante e fatura.")
	}
	return score, reasons
}

func pixIDsMatch(input models.FinancialAgentVerifyReceiptInput, inv models.FinancialAgentInvoice) bool {
	candidates := []string{input.PixTxID, input.PixEndToEnd, onlyDigitsRunes(input.ReceiptText)}
	want := []string{derefStr(inv.SicrediPixTxID)}
	for _, c := range candidates {
		c = strings.ToLower(strings.TrimSpace(c))
		if c == "" {
			continue
		}
		for _, w := range want {
			w = strings.ToLower(strings.TrimSpace(w))
			if w != "" && (c == w || strings.Contains(c, w) || strings.Contains(w, c)) {
				return true
			}
		}
	}
	return false
}

func amountsClose(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= 0.05
}

func receiptReference(input models.FinancialAgentVerifyReceiptInput) string {
	if v := strings.TrimSpace(firstNonEmpty(input.PixEndToEnd, input.PixTxID, input.NossoNumero)); v != "" {
		return "Comprovante " + v
	}
	return "Comprovante agente financeiro"
}

func agentLateFee(settings *models.OrganizationSettingsResponse) float64 {
	if settings == nil || settings.System.LateFeePercentage <= 0 {
		return 2
	}
	return settings.System.LateFeePercentage
}

func agentInterestMonthly(settings *models.OrganizationSettingsResponse) float64 {
	if settings == nil || settings.System.InterestRateMonthly <= 0 {
		return 1
	}
	return settings.System.InterestRateMonthly
}

func parseAgentDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	layouts := []string{"2006-01-02", "02/01/2006", time.RFC3339, "2006-01-02T15:04:05", "02/01/2006 15:04", "02/01/2006 15:04:05"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.UTC); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func derefStr(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func onlyDigitsRunes(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func mustOrg(ctx context.Context) string {
	id, _ := orgFrom(ctx)
	return id
}
