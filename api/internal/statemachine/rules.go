package statemachine

import "strings"

const (
	EntityPhoneLine               = "phone_line"
	EntityProcessingMonth         = "processing_month"
	EntityProviderInvoice         = "provider_invoice"
	EntityCustomerBillingDocument = "customer_billing_document"
)

type TransitionRule struct {
	FromState    string
	ToState      string
	AllowedRoles []string
	Description  string
}

// CanonicalRules define o mapa de regras em memória para validação determinística.
var CanonicalRules = map[string][]TransitionRule{
	EntityPhoneLine: {
		// in_stock
		{FromState: "in_stock", ToState: "in_transition", AllowedRoles: []string{"operational", "master"}, Description: "Vinculação antecipada com trâmite"},
		{FromState: "in_stock", ToState: "active", AllowedRoles: []string{"operational", "master"}, Description: "Ativação/vinculação direta ou importação"},
		{FromState: "in_stock", ToState: "inactive", AllowedRoles: []string{"operational", "master"}, Description: "Linha de estoque ausente em fatura"},
		{FromState: "in_stock", ToState: "cancelled", AllowedRoles: []string{"operational", "master"}, Description: "Cancelamento ou descarte do número"},

		// in_transition
		{FromState: "in_transition", ToState: "active", AllowedRoles: []string{"operational", "master"}, Description: "Conciliação automática via fatura ou ativação"},
		{FromState: "in_transition", ToState: "in_stock", AllowedRoles: []string{"operational", "master"}, Description: "Desvinculação/cancelamento do trâmite de migração"},
		{FromState: "in_transition", ToState: "cancelled", AllowedRoles: []string{"operational", "master"}, Description: "Cancelamento do trâmite de ativação"},

		// active
		{FromState: "active", ToState: "awaiting_invoice", AllowedRoles: []string{"operational", "master"}, Description: "Linha sumiu da fatura atual"},
		{FromState: "active", ToState: "suspended", AllowedRoles: []string{"operational", "master"}, Description: "Suspensão temporária a pedido ou inadimplência"},
		{FromState: "active", ToState: "in_stock", AllowedRoles: []string{"operational", "master"}, Description: "Desvinculação do cliente e retorno ao estoque"},
		{FromState: "active", ToState: "cancelled", AllowedRoles: []string{"operational", "master"}, Description: "Cancelamento definitivo da linha"},
		{FromState: "active", ToState: "in_transition", AllowedRoles: []string{"operational", "master"}, Description: "Início de processo de troca/migração de titularidade"},

		// awaiting_invoice
		{FromState: "awaiting_invoice", ToState: "active", AllowedRoles: []string{"operational", "master"}, Description: "Linha reapareceu em fatura recente"},
		{FromState: "awaiting_invoice", ToState: "in_stock", AllowedRoles: []string{"operational", "master"}, Description: "Desvinculação do cliente"},
		{FromState: "awaiting_invoice", ToState: "cancelled", AllowedRoles: []string{"operational", "master"}, Description: "Cancelamento definitivo confirmado"},

		// suspended
		{FromState: "suspended", ToState: "active", AllowedRoles: []string{"operational", "master"}, Description: "Reativação da linha após suspensão"},
		{FromState: "suspended", ToState: "cancelled", AllowedRoles: []string{"operational", "master"}, Description: "Cancelamento definitivo da linha suspensa"},
		{FromState: "suspended", ToState: "in_stock", AllowedRoles: []string{"operational", "master"}, Description: "Desvinculação e devolução ao estoque"},

		// inactive
		{FromState: "inactive", ToState: "in_stock", AllowedRoles: []string{"operational", "master"}, Description: "Linha inativa reapareceu em fatura"},
		{FromState: "inactive", ToState: "cancelled", AllowedRoles: []string{"operational", "master"}, Description: "Baixa permanente do número"},

		// cancelled
		{FromState: "cancelled", ToState: "in_stock", AllowedRoles: []string{"master"}, Description: "Reativação excepcional de linha cancelada"},
	},

	EntityProcessingMonth: {
		{FromState: "open", ToState: "closed", AllowedRoles: []string{"operational", "master"}, Description: "Fechamento regular ou contingencial do mês"},
		{FromState: "closed", ToState: "open", AllowedRoles: []string{"master"}, Description: "Reabertura excepcional do mês para ajustes controlados"},
	},

	EntityProviderInvoice: {
		{FromState: "draft", ToState: "pending", AllowedRoles: []string{"operational", "master"}, Description: "Fatura importada e aguardando conciliação"},
		{FromState: "pending", ToState: "paid", AllowedRoles: []string{"financial", "master"}, Description: "Liquidação da fatura da operadora"},
		{FromState: "pending", ToState: "overdue", AllowedRoles: []string{"financial", "master"}, Description: "Fatura vencida sem baixa"},
		{FromState: "pending", ToState: "substituted", AllowedRoles: []string{"operational", "master"}, Description: "Fatura substituída; original permanece"},
		{FromState: "overdue", ToState: "substituted", AllowedRoles: []string{"operational", "master"}, Description: "Fatura vencida substituída"},
		{FromState: "overdue", ToState: "paid", AllowedRoles: []string{"financial", "master"}, Description: "Liquidação de fatura em atraso"},
		{FromState: "overdue", ToState: "cancelled", AllowedRoles: []string{"operational", "master"}, Description: "Cancelamento de fatura vencida"},
	},

	EntityCustomerBillingDocument: {
		{FromState: "draft", ToState: "ready", AllowedRoles: []string{"financial", "master"}, Description: "Fatura do cliente validada e pronta para envio"},
		{FromState: "draft", ToState: "cancelled", AllowedRoles: []string{"financial", "master"}, Description: "Fatura rascunho cancelada"},
		{FromState: "ready", ToState: "sent", AllowedRoles: []string{"financial", "master"}, Description: "Fatura e boleto despachados ao cliente"},
		{FromState: "ready", ToState: "cancelled", AllowedRoles: []string{"financial", "master"}, Description: "Fatura pronta cancelada antes do envio"},
		{FromState: "sent", ToState: "cancelled", AllowedRoles: []string{"financial", "master"}, Description: "Fatura enviada estornada/cancelada"},
	},
}

func hasAllowedRole(requiredRoles, userRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true
	}
	for _, req := range requiredRoles {
		reqClean := strings.ToLower(strings.TrimSpace(req))
		for _, user := range userRoles {
			userClean := strings.ToLower(strings.TrimSpace(user))
			if userClean == "master" || userClean == reqClean {
				return true
			}
		}
	}
	return false
}
