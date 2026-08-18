package statemachine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

// StateLogRecorder define a interface para persistência dos logs de transição.
type StateLogRecorder interface {
	InsertStateTransitionLog(ctx context.Context, id, orgID, entityType, entityID, fromState, toState, triggerEvent string, justification, actorUserID *string, metadataJSON *string, ts time.Time) error
}

// Engine centraliza as validações e execuções de máquinas de estado.
type Engine struct {
	recorder StateLogRecorder
}

func NewEngine(recorder StateLogRecorder) *Engine {
	return &Engine{recorder: recorder}
}

// CanTransition verifica se a transição entre estados é permitida para o conjunto de roles do usuário.
func (e *Engine) CanTransition(entityType, fromState, toState string, userRoles []string) (bool, string, error) {
	fromNorm := strings.ToLower(strings.TrimSpace(fromState))
	toNorm := strings.ToLower(strings.TrimSpace(toState))

	if fromNorm == toNorm {
		return true, "Estado inalterado", nil
	}

	rules, ok := CanonicalRules[entityType]
	if !ok {
		return false, "", fmt.Errorf("tipo de entidade '%s' não possui máquina de estado configurada", entityType)
	}

	var foundTransition *TransitionRule
	for _, rule := range rules {
		if rule.FromState == fromNorm && rule.ToState == toNorm {
			foundTransition = &rule
			break
		}
	}

	if foundTransition == nil {
		return false, fmt.Sprintf("Transição de '%s' para '%s' não é permitida para a entidade '%s'", fromState, toState, entityType), nil
	}

	if !hasAllowedRole(foundTransition.AllowedRoles, userRoles) {
		return false, fmt.Sprintf("Usuário não possui perfil autorizado para transicionar de '%s' para '%s' (requer: %s)", fromState, toState, strings.Join(foundTransition.AllowedRoles, ", ")), nil
	}

	return true, foundTransition.Description, nil
}

// ValidateTransition valida se a transição pode ocorrer, retornando um erro HTTP amigável em caso de recusa.
func (e *Engine) ValidateTransition(entityType, fromState, toState string, userRoles []string) error {
	allowed, reason, err := e.CanTransition(entityType, fromState, toState, userRoles)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !allowed {
		return httputil.BusinessError(notifications.N("INVALID_STATE_TRANSITION", reason))
	}
	return nil
}

// RecordTransition registra a transição efetuada no histórico de auditoria do estado.
func (e *Engine) RecordTransition(ctx context.Context, orgID, entityType, entityID, fromState, toState, triggerEvent string, justification, actorUserID *string, metadataJSON *string) error {
	if e.recorder == nil {
		return nil
	}
	id := uuid.New().String()
	return e.recorder.InsertStateTransitionLog(ctx, id, orgID, entityType, entityID, fromState, toState, triggerEvent, justification, actorUserID, metadataJSON, time.Now().UTC())
}
