package importservice

import (
	"os"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/statemachine"
)

func importEnforceStateMachine() bool {
	v := strings.TrimSpace(os.Getenv("IMPORT_ENFORCE_STATE_MACHINE"))
	if v == "" {
		return true
	}
	switch strings.ToLower(v) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func importActorRoles() []string {
	return []string{"operational", "employee", "master"}
}

// validateImportLineTransition aplica a máquina de estado da linha.
// enforce=false desliga a recusa (rollback da feature) mas a transição continua sendo a mesma.
func validateImportLineTransition(engine *statemachine.Engine, from, to string, enforce bool) error {
	if engine == nil || from == to {
		return nil
	}
	if !enforce {
		return nil
	}
	return engine.ValidateTransition(statemachine.EntityPhoneLine, from, to, importActorRoles())
}
