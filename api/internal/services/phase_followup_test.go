package services_test

import (
	"strings"
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
)

func TestTwoLevelApproverRoles(t *testing.T) {
	if !auth.IsPrivilegedApproverRole("master") || !auth.IsPrivilegedApproverRole("financial") || !auth.IsPrivilegedApproverRole("admin") {
		t.Fatal("privileged roles should include master, admin and financial")
	}
	if auth.IsPrivilegedApproverRole("employee") {
		t.Fatal("employee must not be a two-level approver by itself")
	}
}

func TestPortalDocumentDigits(t *testing.T) {
	doc := httputil.NormalizeDigits("123.456.789-01")
	if doc != "12345678901" {
		t.Fatalf("expected normalized CPF, got %s", doc)
	}
	if strings.Contains(doc, ".") {
		t.Fatal("document must not keep formatting")
	}
}
