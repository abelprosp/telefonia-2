package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// IntakeMoney stores BRL in cents, never in binary floating point.
type IntakeMoney int64

var intakeMoneyPattern = regexp.MustCompile(`^[0-9]{1,12}(\.[0-9]{1,2})?$`)

func (m *IntakeMoney) UnmarshalJSON(raw []byte) error {
	s := string(raw)
	if !intakeMoneyPattern.MatchString(s) {
		return fmt.Errorf("valor monetário deve ser um número não negativo com até duas casas decimais")
	}
	parts := strings.SplitN(s, ".", 2)
	whole, _ := strconv.ParseInt(parts[0], 10, 64)
	var cents int64
	if len(parts) == 2 {
		cents, _ = strconv.ParseInt((parts[1] + "00")[:2], 10, 64)
	}
	*m = IntakeMoney(whole*100 + cents)
	return nil
}

func (m IntakeMoney) String() string               { return fmt.Sprintf("%d.%02d", m/100, m%100) }
func (m IntakeMoney) MarshalJSON() ([]byte, error) { return []byte(m.String()), nil }

type CustomerIntakeInput struct {
	SourceSystem string                     `json:"sourceSystem,omitempty"`
	Sale         *CustomerIntakeSale        `json:"sale"`
	Client       *CustomerIntakeClient      `json:"client"`
	Attachments  []CustomerIntakeAttachment `json:"attachments,omitempty"`
}

type CustomerIntakeSale struct {
	ID                string       `json:"id"`
	Protocol          string       `json:"protocol"`
	TaskProtocol      string       `json:"taskProtocol"`
	SoldAt            *time.Time   `json:"soldAt"`
	ConcludedAt       *time.Time   `json:"concludedAt"`
	Value             *IntakeMoney `json:"value"`
	CommissionValue   *IntakeMoney `json:"commissionValue"`
	OperatorName      string       `json:"operatorName"`
	PlanName          string       `json:"planName"`
	LineNumber        string       `json:"lineNumber"`
	SimType           string       `json:"simType"`
	IsVirginChip      bool         `json:"isVirginChip"`
	ICCID             string       `json:"iccid"`
	DeviceIMEI        string       `json:"deviceImei"`
	DeviceEID         string       `json:"deviceEid"`
	DevicePlatform    string       `json:"devicePlatform"`
	IsPortability     bool         `json:"isPortability"`
	PortabilityNumber string       `json:"portabilityNumber"`
	DonorOperator     string       `json:"donorOperator"`
	IsUpgrade         bool         `json:"isUpgrade"`
	UpgradeOfProtocol string       `json:"upgradeOfProtocol"`
	ContractFormat    string       `json:"contractFormat"`
	PartnerName       string       `json:"partnerName"`
	BranchName        string       `json:"branchName"`
	Notes             string       `json:"notes"`
}

type CustomerIntakeClient struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	Document      string `json:"document"`
	DocumentType  string `json:"documentType"`
	LegalName     string `json:"legalName,omitempty"`
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

type CustomerIntakeAttachment struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	MimeType      string `json:"mimeType"`
	ContentBase64 string `json:"contentBase64,omitempty"`
}

type CustomerIntakeAttachmentInfo struct {
	ExternalID   string `json:"external_id"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	MimeType     string `json:"mime_type"`
	AttachmentID string `json:"attachment_id"`
	SizeBytes    int64  `json:"size_bytes"`
}

type CustomerIntakeRecord struct {
	ID           string                         `json:"id"`
	SourceSystem string                         `json:"source_system"`
	CustomerID   string                         `json:"customer_id"`
	SaleID       string                         `json:"sale_id"`
	Sale         CustomerIntakeSale             `json:"sale"`
	Client       CustomerIntakeClient           `json:"client"`
	Attachments  []CustomerIntakeAttachmentInfo `json:"attachments"`
	ImportedAt   time.Time                      `json:"imported_at"`
	PayloadHash  string                         `json:"-"`
}

type CustomerIntakeResponse struct {
	CustomerID string `json:"customer_id"`
	SaleID     string `json:"sale_id"`
	IntakeID   string `json:"intake_id"`
	Replayed   bool   `json:"replayed"`
}
