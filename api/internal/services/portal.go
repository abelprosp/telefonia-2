package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/precision"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) PortalGetMe(ctx context.Context) (*models.PortalCustomerMeResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}

	customerID, err := s.resolvePortalCustomerID(ctx, orgID, user)
	if err != nil {
		return nil, err
	}
	cust, err := s.Store.GetCustomerInOrg(ctx, orgID, customerID, nil)
	if err != nil || cust == nil {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}

	lines, _, _ := s.Store.ListCustomerPhoneLines(ctx, orgID, cust.ID, httputil.PageSearch{PageSize: 500})
	activeCount := 0
	var totalMonthly float64
	for _, l := range lines {
		if l.IsActive {
			activeCount++
			if l.MonthlyAmount != nil {
				totalMonthly = precision.SumCents(totalMonthly, *l.MonthlyAmount)
			}
		}
	}

	return &models.PortalCustomerMeResponse{
		Customer:           *cust,
		ActiveLinesCount:   activeCount,
		TotalMonthlyAmount: totalMonthly,
	}, nil
}

func (s *Service) resolvePortalCustomerID(ctx context.Context, orgID string, user *auth.User) (string, error) {
	if link, err := s.Store.GetPortalLinkByUser(ctx, orgID, user.ID); err != nil {
		return "", httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	} else if link != nil {
		return link.CustomerID, nil
	}
	doc := httputil.NormalizeDigits(user.Document)
	if len(doc) != 11 && len(doc) != 14 {
		return "", httputil.BusinessError(notifications.N("PORTAL_DOCUMENT_REQUIRED", "Vínculo do portal exige CPF/CNPJ no usuário. Não há fallback por e-mail."))
	}
	ids, err := s.Store.ListCustomersByDocument(ctx, orgID, doc)
	if err != nil {
		return "", httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if len(ids) == 0 {
		return "", httputil.NotFoundError(notifications.CustomerNotFound)
	}
	if len(ids) != 1 {
		return "", httputil.BusinessError(notifications.N("PORTAL_DOCUMENT_AMBIGUOUS", "Há mais de um cliente com este CPF/CNPJ. Vínculo recusado para não vazar dados."))
	}
	now := time.Now().UTC()
	_ = s.Store.UpsertPortalCustomerLink(ctx, store.PortalCustomerLinkRow{
		ID: uuid.New().String(), OrganizationID: orgID, UserID: user.ID, CustomerID: ids[0], Document: doc, CreatedAt: now,
	})
	return ids[0], nil
}

func (s *Service) PortalListContracts(ctx context.Context) ([]models.GeneratedContractResponse, error) {
	me, err := s.PortalGetMe(ctx)
	if err != nil {
		return nil, err
	}
	return s.ListGeneratedContractsForCustomer(ctx, me.Customer.ID)
}

func (s *Service) PortalUpdateProfile(ctx context.Context, input models.PortalUpdateProfileInput) (*models.PortalCustomerMeResponse, error) {
	me, err := s.PortalGetMe(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if input.BillingEmail != nil {
		email := strings.TrimSpace(*input.BillingEmail)
		if err := s.Store.UpdateCustomerBillingEmail(ctx, orgID, me.Customer.ID, email); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}
	return s.PortalGetMe(ctx)
}

func (s *Service) PortalListTickets(ctx context.Context, page httputil.PageSearch) ([]models.SupportTicketResponse, int64, error) {
	me, err := s.PortalGetMe(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.ListSupportTickets(ctx, me.Customer.ID, "", page)
}

func (s *Service) PortalCreateTicket(ctx context.Context, input models.CreateSupportTicketInput) (*models.SupportTicketResponse, error) {
	me, err := s.PortalGetMe(ctx)
	if err != nil {
		return nil, err
	}
	input.CustomerID = me.Customer.ID
	return s.CreateSupportTicket(ctx, input)
}

func (s *Service) PortalDownloadInvoice(ctx context.Context, documentID string) ([]byte, string, error) {
	me, err := s.PortalGetMe(ctx)
	if err != nil {
		return nil, "", err
	}
	doc, err := s.GetCustomerBillingDocument(ctx, documentID)
	if err != nil {
		return nil, "", err
	}
	if doc.CustomerID != me.Customer.ID {
		return nil, "", httputil.NotFoundError(notifications.BillingDocumentNotFound)
	}
	return s.GetCustomerBillingDocumentDownload(ctx, documentID)
}

func (s *Service) PortalListLines(ctx context.Context) ([]models.PortalLineItemResponse, error) {
	me, err := s.PortalGetMe(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	links, _, err := s.Store.ListCustomerPhoneLines(ctx, orgID, me.Customer.ID, httputil.PageSearch{PageSize: 500})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	var items []models.PortalLineItemResponse
	for _, cl := range links {
		l, err := s.Store.GetPhoneLine(ctx, orgID, cl.PhoneLineID)
		if err != nil || l == nil {
			continue
		}
		amt := 0.0
		if cl.MonthlyAmount != nil {
			amt = *cl.MonthlyAmount
		}
		fid, _ := s.GetLineFidelity(ctx, l.ID)
		hasFid := fid != nil && fid.Status == "active"

		items = append(items, models.PortalLineItemResponse{
			ID:                 l.ID,
			Number:             l.Number,
			PlanName:           l.ProviderPlanName,
			LineClassification: l.LineClassification,
			MonthlyAmount:      amt,
			Status:             l.Status,
			HasActiveFidelity:  hasFid,
		})
	}

	return items, nil
}

func (s *Service) PortalListInvoices(ctx context.Context) ([]models.PortalInvoiceItemResponse, error) {
	me, err := s.PortalGetMe(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	docs, _, err := s.Store.ListCustomerBillingDocuments(ctx, orgID, nil, &me.Customer.ID, httputil.PageSearch{PageSize: 100})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	var items []models.PortalInvoiceItemResponse
	for _, d := range docs {
		hasBoleto := d.SicrediNossoNumero != nil && *d.SicrediNossoNumero != ""
		items = append(items, models.PortalInvoiceItemResponse{
			ID:            d.ID,
			InvoiceNumber: d.InvoiceNumber,
			DueDate:       d.DueDate,
			TotalAmount:   d.Amount,
			Status:        d.Status,
			HasBoleto:     hasBoleto,
			PixCode:       d.SicrediLinhaDigitavel,
		})
	}

	return items, nil
}

func (s *Service) GenerateSicrediPix(ctx context.Context, documentID string) (*models.SicrediPixGenerateResponse, error) {
	doc, err := s.GetCustomerBillingDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}

	nossoNum := "0000000000"
	if doc.SicrediNossoNumero != nil && *doc.SicrediNossoNumero != "" {
		nossoNum = *doc.SicrediNossoNumero
	}

	pixCopiaCola := ""
	qrCodeURL := ""
	if doc.SicrediPixQrCode != nil && strings.TrimSpace(*doc.SicrediPixQrCode) != "" {
		pixCopiaCola = strings.TrimSpace(*doc.SicrediPixQrCode)
		qrCodeURL = pixCopiaCola
	} else {
		pixCopiaCola = fmt.Sprintf("00020126580014br.gov.bcb.pix0136%s520400005303986540%0.2f5802BR5915LUXUS CONECTA6009SAO PAULO62070503***6304", nossoNum, doc.Amount)
		qrCodeURL = fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=%s", pixCopiaCola)
		orgID, _ := orgFrom(ctx)
		linha := ""
		if doc.SicrediLinhaDigitavel != nil {
			linha = *doc.SicrediLinhaDigitavel
		}
		barras := ""
		if doc.SicrediCodigoBarras != nil {
			barras = *doc.SicrediCodigoBarras
		}
		status := ""
		if doc.SicrediBoletoStatus != nil {
			status = *doc.SicrediBoletoStatus
		}
		_ = s.Store.UpdateCustomerBillingDocumentSicredi(ctx, orgID, documentID, nossoNum, linha, barras, pixCopiaCola, "", status, doc.SicrediBoletoError, "", time.Now().UTC())
	}

	s.auditLog(ctx, "GeneratePix", "CustomerBillingDocument", documentID, nil, map[string]any{
		"nosso_numero": nossoNum, "amount": doc.Amount,
	})

	return &models.SicrediPixGenerateResponse{
		DocumentID:   documentID,
		NossoNumero:  nossoNum,
		PixQrCode:    qrCodeURL,
		PixCopiaCola: pixCopiaCola,
		ExpiresAt:    doc.DueDate.Add(24 * time.Hour),
	}, nil
}

func (s *Service) GetFinancialSummaryReport(ctx context.Context) (*models.FinancialSummaryReportResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	custs, _, err := s.Store.ListCustomers(ctx, orgID, nil, nil, httputil.PageSearch{PageSize: 1000})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	var items []models.FinancialSummaryReportItem
	var totalGross, totalCost float64
	totalLines := 0

	for _, c := range custs {
		lines, _, err := s.Store.ListCustomerPhoneLines(ctx, orgID, c.ID, httputil.PageSearch{PageSize: 500})
		if err != nil {
			continue
		}
		var custGross, custCost float64
		activeLines := 0
		for _, cl := range lines {
			if cl.IsActive {
				activeLines++
				if cl.MonthlyAmount != nil {
					custGross = precision.SumCents(custGross, *cl.MonthlyAmount)
				}
				pl, err := s.Store.GetPhoneLine(ctx, orgID, cl.PhoneLineID)
				if err == nil && pl != nil && pl.BaseCost != nil {
					custCost = precision.SumCents(custCost, *pl.BaseCost)
				}
			}
		}
		if activeLines > 0 {
			totalLines += activeLines
			totalGross = precision.SumCents(totalGross, custGross)
			totalCost = precision.SumCents(totalCost, custCost)
			custMargin := precision.Round2(custGross - custCost)
			custMarginPct := 0.0
			if custGross > 0 {
				custMarginPct = precision.Round2((custMargin / custGross) * 100.0)
			}
			items = append(items, models.FinancialSummaryReportItem{
				CustomerID:          c.ID,
				CustomerName:        c.Name,
				ContractedLuxusCnpj: c.ContractedLuxusCnpj,
				LinesCount:          activeLines,
				TotalGrossRevenue:   custGross,
				TotalOperatorCost:   custCost,
				GrossMargin:         custMargin,
				MarginPercentage:    custMarginPct,
			})
		}
	}

	totalMargin := precision.Round2(totalGross - totalCost)

	return &models.FinancialSummaryReportResponse{
		GeneratedAt: time.Now().UTC(),
		TotalLines:  totalLines,
		TotalGross:  totalGross,
		TotalCost:   totalCost,
		TotalMargin: totalMargin,
		Items:       items,
	}, nil
}

func (s *Service) GetCustomerProfitabilityReport(ctx context.Context) (*models.CustomerProfitabilityReportResponse, error) {
	summary, err := s.GetFinancialSummaryReport(ctx)
	if err != nil {
		return nil, err
	}

	var items []models.CustomerProfitabilityItem
	for _, item := range summary.Items {
		avgTicket := 0.0
		if item.LinesCount > 0 {
			avgTicket = precision.Round2(item.TotalGrossRevenue / float64(item.LinesCount))
		}
		items = append(items, models.CustomerProfitabilityItem{
			CustomerID:       item.CustomerID,
			CustomerName:     item.CustomerName,
			ActiveLines:      item.LinesCount,
			AverageTicket:    avgTicket,
			GrossRevenue:     item.TotalGrossRevenue,
			Cost:             item.TotalOperatorCost,
			Margin:           item.GrossMargin,
			MarginPercentage: item.MarginPercentage,
		})
	}

	return &models.CustomerProfitabilityReportResponse{
		GeneratedAt: time.Now().UTC(),
		Items:       items,
	}, nil
}
