package importservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/vivo"
)

func (p *Processor) PreviewImport(ctx context.Context, orgID string, input models.ProviderInvoiceImportRequestInput) (*models.ImportPreviewResponse, error) {
	if p.Storage == nil {
		return nil, httputil.BusinessError(notifications.ObjectStorageUnavailable)
	}
	gotOrg, _, err := p.Store.GetProviderByID(ctx, input.ProviderID)
	if err != nil {
		return nil, err
	}
	if gotOrg != orgID {
		return nil, httputil.NotFoundError(notifications.ProviderNotFound)
	}
	raw, err := p.Storage.GetObject(ctx, input.StorageBucket, input.StorageObjectKey)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError("storage get: " + err.Error()))
	}
	if isPDFBytes(raw) {
		return &models.ImportPreviewResponse{
			Warnings: []string{notifications.ImportPDFNotParsed.Message},
			IsValid:  false,
		}, nil
	}

	sum := sha256.Sum256(raw)
	fileHash := hex.EncodeToString(sum[:])
	parsed, err := vivo.ParseLatin1(raw)
	if err != nil {
		return nil, httputil.BusinessError(notifications.N("IMPORT_PARSE_FAILED", "Não foi possível interpretar o TXT VIVO."))
	}

	header := getHeader(parsed)
	if header == nil {
		return &models.ImportPreviewResponse{
			Warnings: []string{"Registro 010D (cabeçalho da fatura) não encontrado."},
			IsValid:  false,
		}, nil
	}

	numbersMap := buildNumbersFrom110D(parsed)
	lineItems := make([]string, 0, len(numbersMap))
	known, unknown := 0, 0
	for n := range numbersMap {
		lineItems = append(lineItems, n)
		pl, err := p.Store.GetPhoneLineByNumber(ctx, n)
		if err != nil {
			return nil, err
		}
		if pl != nil {
			known++
		} else {
			unknown++
		}
	}

	var warnings []string
	duplicate := false
	existingHash, err := p.Store.FindActiveInvoiceByContentSHA256(ctx, fileHash)
	if err != nil {
		return nil, err
	}
	if existingHash != nil {
		duplicate = true
		warnings = append(warnings, "Este arquivo já foi importado. Para substituir uma fatura, envie um arquivo corrigido.")
	}
	if getCustomer011(parsed) == nil {
		warnings = append(warnings, "Registro 011D (cliente) ausente — a importação falhará.")
	}

	valid := getCustomer011(parsed) != nil && !duplicate
	if customer := getCustomer011(parsed); customer != nil {
		doc := httputil.NormalizeDigits(customer.Document)
		if len(doc) != 11 && len(doc) != 14 {
			valid = false
			warnings = append(warnings, notifications.ImportCustomerDocumentInvalid.Message)
		}
	}
	sort.Strings(lineItems)

	return &models.ImportPreviewResponse{
		Summary: models.ImportPreviewInvoiceSummary{
			InvoiceNumber: header.ReferenceMonth,
			AccountNumber: header.AccountNumber,
			IssueDate:     header.IssueDate,
			DueDate:       header.DueDate,
			TotalAmount:   header.TotalAmount,
			LinesCount:    len(lineItems),
			KnownLines:    known,
			UnknownLines:  unknown,
		},
		LineItems:  lineItems,
		Warnings:   warnings,
		IsValid:    valid,
		FileSHA256: fileHash,
	}, nil
}
