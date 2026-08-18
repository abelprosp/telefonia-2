package services_test

import (
	"context"
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/privacy"
)

func TestPhase4_RBACGranularPermissions(t *testing.T) {
	// 1. Usuário Master
	masterCtx := auth.WithUser(context.Background(), &auth.User{
		ID:    "u_master",
		Roles: []string{auth.RoleMaster},
	})
	if !auth.IsMaster(masterCtx) || !auth.CanManageUsers(masterCtx) || !auth.CanApproveOperations(masterCtx) || !auth.CanAnonymizeData(masterCtx) {
		t.Errorf("expected master to have all administrative permissions")
	}

	// 2. Usuário Operador
	operatorCtx := auth.WithUser(context.Background(), &auth.User{
		ID:    "u_operator",
		Roles: []string{auth.RoleOperator},
	})
	if auth.IsMaster(operatorCtx) {
		t.Errorf("operator must not be master")
	}
	if !auth.CanAccessOperational(operatorCtx) {
		t.Errorf("operator must have operational access")
	}
	if auth.CanManageUsers(operatorCtx) || auth.CanApproveOperations(operatorCtx) || auth.CanAnonymizeData(operatorCtx) {
		t.Errorf("operator must not have approval/anonymization permissions")
	}

	// 3. Usuário Financeiro
	financialCtx := auth.WithUser(context.Background(), &auth.User{
		ID:    "u_financial",
		Roles: []string{auth.RoleFinancial},
	})
	if !auth.CanAccessFinancial(financialCtx) {
		t.Errorf("financial user must have financial access")
	}
	if auth.CanApproveOperations(financialCtx) {
		t.Errorf("financial user must not approve operational requests")
	}
}

func TestPhase4_LGPDMasking(t *testing.T) {
	cpf := "12345678901"
	maskedCPF := privacy.MaskDocument(cpf)
	if maskedCPF != "123.***.***-01" {
		t.Errorf("expected 123.***.***-01, got %s", maskedCPF)
	}

	cnpj := "12345678000190"
	maskedCNPJ := privacy.MaskDocument(cnpj)
	if maskedCNPJ != "12.***.***/0001-90" {
		t.Errorf("expected 12.***.***/0001-90, got %s", maskedCNPJ)
	}

	email := "cliente.vip@luxus.com.br"
	maskedEmail := privacy.MaskEmail(email)
	if maskedEmail != "cl*********@luxus.com.br" {
		t.Errorf("expected cl*********@luxus.com.br, got %s", maskedEmail)
	}
}
