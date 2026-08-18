package services

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/auth"
)

func TestPipelineSummaryIncludesSkipReason(t *testing.T) {
	raw := pipelineSummary(map[string]any{"invoices": 0}, "no_invoices")
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["skip_reason"] != "no_invoices" {
		t.Fatalf("expected skip_reason no_invoices, got %#v", parsed["skip_reason"])
	}
	if parsed["invoices"].(float64) != 0 {
		t.Fatalf("expected invoices 0")
	}
}

func TestPipelineSummaryOmitsSkipWhenEmpty(t *testing.T) {
	raw := pipelineSummary(map[string]any{"ok": true}, "")
	if strings.Contains(raw, "skip_reason") {
		t.Fatalf("empty skip_reason should be omitted: %s", raw)
	}
}

func TestSanitizeTicketFileName(t *testing.T) {
	got := sanitizeTicketFileName(`..\..\secret.pdf`)
	if got != "secret.pdf" {
		t.Fatalf("expected secret.pdf, got %q", got)
	}
	if sanitizeTicketFileName("..") != "" {
		t.Fatal("traversal-only name must be rejected")
	}
}

func TestValidTicketObjectKey(t *testing.T) {
	org, ticket := "org-1", "t-1"
	ok := "tickets/org-1/t-1/uuid_file.pdf"
	if !validTicketObjectKey(org, ticket, ok) {
		t.Fatal("expected valid key")
	}
	if validTicketObjectKey(org, ticket, "tickets/other/t-1/x.pdf") {
		t.Fatal("key from another org must be rejected")
	}
	if validTicketObjectKey(org, ticket, "tickets/org-1/t-1/../x.pdf") {
		t.Fatal("path traversal must be rejected")
	}
}

func TestMFAVerifiedFromClaims(t *testing.T) {
	if !auth.MFAVerifiedFromClaims("1", []string{"pwd", "otp"}) {
		t.Fatal("amr otp should count as MFA")
	}
	if !auth.MFAVerifiedFromClaims("2", []string{"pwd"}) {
		t.Fatal("acr 2 should count as MFA")
	}
	if auth.MFAVerifiedFromClaims("1", []string{"pwd"}) {
		t.Fatal("password-only must not count as MFA")
	}
}
