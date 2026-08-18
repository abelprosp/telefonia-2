package importservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

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
	if err != nil || gotOrg != orgID {
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
		if pl, err := p.Store.GetPhoneLineByNumber(ctx, n); err == nil && pl != nil {
			known++
		} else {
			unknown++
		}
	}

	var warnings []string
	duplicate := false
	if existingHash, err := p.Store.FindActiveInvoiceByContentSHA256(ctx, fileHash); err == nil && existingHash != nil {
		duplicate = true
		warnings = append(warnings, "Arquivo duplicado (SHA-256 idêntico a uma fatura ativa). Marque substituição explícita para continuar.")
	}
	if getCustomer011(parsed) == nil {
		warnings = append(warnings, "Registro 011D (cliente) ausente — a importação falhará.")
	}

	valid := getCustomer011(parsed) != nil && (!duplicate || input.AllowSubstitute)

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
