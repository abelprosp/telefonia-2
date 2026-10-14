package services

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func validIntake(t *testing.T) models.CustomerIntakeInput {
	t.Helper()
	var input models.CustomerIntakeInput
	err := json.Unmarshal([]byte(`{"sale":{"id":"remote-1","protocol":"VND-ABC123","taskProtocol":"LUX-2026-00001","soldAt":"2026-10-14T14:30:00.000Z","concludedAt":"2026-10-14T18:05:00.000Z","value":79.9,"commissionValue":10,"operatorName":"Vivo","planName":"Controle","lineNumber":"51985370873","simType":"CHIP","isVirginChip":false,"iccid":"8955100000000000000","deviceImei":"490154203237518","deviceEid":"89049032000000000000000000000000","devicePlatform":"ANDROID","isPortability":true,"portabilityNumber":"51999998888","donorOperator":"TIM","isUpgrade":false,"upgradeOfProtocol":"","contractFormat":"ZAPSIGN","partnerName":"META CELL MATRIZ","branchName":"CRUZEIRO","notes":"observação"},"client":{"name":"Nome do cliente","document":"02113419084","documentType":"CPF","rg":"1234567890","email":"cliente@example.com","phone":"51999998888","address":"Rua Exemplo","addressNumber":"100","complement":"apto 2","neighborhood":"Centro","city":"Lajeado","state":"RS","zipCode":"95900000"},"attachments":[]}`), &input)
	if err != nil {
		t.Fatal(err)
	}
	return input
}

func TestCustomerIntakePayloadPreservesFieldsAndLeadingZeros(t *testing.T) {
	input := validIntake(t)
	c, files, hash, err := validateCustomerIntake(&input)
	if err != nil {
		t.Fatal(err)
	}
	if c.Document != "02113419084" || c.Type != "PF" || c.Profile.RG != "1234567890" || c.Addresses[0].ZipCode != "95900000" || len(files) != 0 || len(hash) != 64 {
		t.Fatalf("incorrect mapping: %+v", c)
	}
	if *input.Sale.Value != 7990 || *input.Sale.CommissionValue != 1000 || input.Sale.BranchName != "CRUZEIRO" || input.Sale.DeviceEID != "89049032000000000000000000000000" {
		t.Fatalf("sale fields altered: %+v", input.Sale)
	}
	_, _, again, err := validateCustomerIntake(&input)
	if err != nil || again != hash {
		t.Fatal("normalization must be stable for retries")
	}
}

func TestCustomerIntakeRejectsInvalidConditionalFields(t *testing.T) {
	for name, mutate := range map[string]func(*models.CustomerIntakeInput){
		"missing sale":       func(i *models.CustomerIntakeInput) { i.Sale = nil },
		"missing client":     func(i *models.CustomerIntakeInput) { i.Client = nil },
		"missing money":      func(i *models.CustomerIntakeInput) { i.Sale.CommissionValue = nil },
		"invalid cpf length": func(i *models.CustomerIntakeInput) { i.Client.Document = "12" },
		"invalid phone":      func(i *models.CustomerIntakeInput) { i.Client.Phone = "123" },
		"invalid email":      func(i *models.CustomerIntakeInput) { i.Client.Email = "not email" },
		"company legal name": func(i *models.CustomerIntakeInput) {
			i.Client.DocumentType = "CNPJ"
			i.Client.Document = "11222333000181"
			i.Client.RG = ""
		},
		"date order": func(i *models.CustomerIntakeInput) {
			i.Sale.SoldAt = i.Sale.ConcludedAt
			earlier := i.Sale.SoldAt.Add(-1)
			i.Sale.ConcludedAt = &earlier
		},
		"upgrade reference":  func(i *models.CustomerIntakeInput) { i.Sale.IsUpgrade = true },
		"portability number": func(i *models.CustomerIntakeInput) { i.Sale.PortabilityNumber = "abc" },
		"imei":               func(i *models.CustomerIntakeInput) { i.Sale.DeviceIMEI = "123" },
		"excess text":        func(i *models.CustomerIntakeInput) { i.Sale.Notes = strings.Repeat("x", 4001) },
	} {
		t.Run(name, func(t *testing.T) {
			i := validIntake(t)
			mutate(&i)
			if _, _, _, err := validateCustomerIntake(&i); err == nil {
				t.Fatal("invalid intake accepted")
			}
		})
	}
}

func TestCustomerIntakePJAndPDFAttachments(t *testing.T) {
	i := validIntake(t)
	i.Client.DocumentType = "CNPJ"
	i.Client.Document = "11222333000181"
	i.Client.LegalName = "Empresa Ltda"
	i.Client.RG = ""
	i.Attachments = []models.CustomerIntakeAttachment{{ID: "contract-1", Type: "CONTRACT", Name: "contrato.pdf", MimeType: "application/pdf", ContentBase64: base64.StdEncoding.EncodeToString([]byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n%%EOF"))}}
	c, files, _, err := validateCustomerIntake(&i)
	if err != nil {
		t.Fatal(err)
	}
	if c.Type != "PJ" || *c.LegalName != "Empresa Ltda" || len(files) != 1 {
		t.Fatal("company/PDF not mapped")
	}
	for name, mutate := range map[string]func(*models.CustomerIntakeInput){
		"unsafe MIME": func(i *models.CustomerIntakeInput) { i.Attachments[0].MimeType = "text/html" },
		"spoof MIME": func(i *models.CustomerIntakeInput) {
			i.Attachments[0].ContentBase64 = base64.StdEncoding.EncodeToString([]byte("<script>alert(1)</script>"))
		},
		"invalid base64": func(i *models.CustomerIntakeInput) { i.Attachments[0].ContentBase64 = "JVBERi0xLjQK..." },
		"path injection": func(i *models.CustomerIntakeInput) { i.Attachments[0].Name = "../contract.pdf" },
		"duplicate id":   func(i *models.CustomerIntakeInput) { i.Attachments = append(i.Attachments, i.Attachments[0]) },
	} {
		t.Run(name, func(t *testing.T) {
			copy := i
			copy.Attachments = append([]models.CustomerIntakeAttachment{}, i.Attachments...)
			mutate(&copy)
			if _, _, _, err := validateCustomerIntake(&copy); err == nil {
				t.Fatal("unsafe attachment accepted")
			}
		})
	}
}
