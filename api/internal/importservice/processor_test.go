package importservice

import (
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/statemachine"
)

func TestValidateImportLineTransition_InvalidFails(t *testing.T) {
	eng := statemachine.NewEngine(nil)
	err := validateImportLineTransition(eng, "inactive", "active", true)
	if err == nil {
		t.Fatal("expected invalid transition inactive -> active to fail")
	}
}

func TestValidateImportLineTransition_ValidPasses(t *testing.T) {
	eng := statemachine.NewEngine(nil)
	if err := validateImportLineTransition(eng, "in_stock", "active", true); err != nil {
		t.Fatalf("expected valid transition, got %v", err)
	}
}

func TestValidateImportLineTransition_FlagOffSkips(t *testing.T) {
	eng := statemachine.NewEngine(nil)
	if err := validateImportLineTransition(eng, "inactive", "active", false); err != nil {
		t.Fatalf("flag off should skip validation, got %v", err)
	}
}

func TestValidateImportLineTransition_SameState(t *testing.T) {
	eng := statemachine.NewEngine(nil)
	if err := validateImportLineTransition(eng, "active", "active", true); err != nil {
		t.Fatalf("same state should be allowed, got %v", err)
	}
}
