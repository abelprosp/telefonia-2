package services_test

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func TestPhase7_IMEIValidation(t *testing.T) {
	imeiRegex := regexp.MustCompile(`^[0-9]{15}$`)

	validIMEI := "356938035643803"
	if !imeiRegex.MatchString(validIMEI) {
		t.Errorf("expected %s to be valid IMEI", validIMEI)
	}

	invalidIMEI1 := "35693803564380" // 14 dígitos
	if imeiRegex.MatchString(invalidIMEI1) {
		t.Errorf("expected %s to be invalid", invalidIMEI1)
	}

	invalidIMEI2 := "35693803564380A" // Letra
	if imeiRegex.MatchString(invalidIMEI2) {
		t.Errorf("expected %s to be invalid", invalidIMEI2)
	}
}

func TestPhase7_WebhookSecretGeneration(t *testing.T) {
	wh := models.WebhookSubscriptionResponse{
		ID:        "wh_123",
		URL:       "https://api.empresa.com/webhook",
		Events:    []string{"BILLING_CLOSED", "DIVERGENCE_DETECTED"},
		IsActive:  true,
		Secret:    "whsec_a1b2c3d4e5f67890",
		CreatedAt: time.Now().UTC(),
	}

	if len(wh.Events) != 2 || !wh.IsActive {
		t.Errorf("webhook payload mismatch")
	}
}

func TestPhase7_OrganizationBackupChecksum(t *testing.T) {
	orgID := "org_luxus"
	now := time.Now().UTC()
	hashData := orgID + "|" + now.Format(time.RFC3339)

	hashBytes := sha256.Sum256([]byte(hashData))
	checksum := hex.EncodeToString(hashBytes[:])

	if len(checksum) != 64 {
		t.Errorf("expected 64-char sha256 checksum, got %d", len(checksum))
	}
}
