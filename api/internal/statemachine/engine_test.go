package statemachine

import (
	"context"
	"testing"
	"time"
)

type mockRecorder struct {
	recorded []string
}

func (m *mockRecorder) InsertStateTransitionLog(ctx context.Context, id, orgID, entityType, entityID, fromState, toState, triggerEvent string, justification, actorUserID *string, metadataJSON *string, ts time.Time) error {
	m.recorded = append(m.recorded, entityType+":"+fromState+"->"+toState)
	return nil
}

func TestStateMachine_CanTransition_PhoneLine(t *testing.T) {
	engine := NewEngine(nil)

	tests := []struct {
		name      string
		from      string
		to        string
		roles     []string
		wantAllow bool
	}{
		// Valid transitions
		{"in_stock to in_transition (operational)", "in_stock", "in_transition", []string{"operational"}, true},
		{"in_stock to active (operational)", "in_stock", "active", []string{"operational"}, true},
		{"in_stock to inactive (operational)", "in_stock", "inactive", []string{"operational"}, true},
		{"in_transition to active (operational)", "in_transition", "active", []string{"operational"}, true},
		{"active to awaiting_invoice (operational)", "active", "awaiting_invoice", []string{"operational"}, true},
		{"active to suspended (operational)", "active", "suspended", []string{"operational"}, true},
		{"active to in_stock (operational)", "active", "in_stock", []string{"operational"}, true},
		{"active to cancelled (operational)", "active", "cancelled", []string{"operational"}, true},
		{"awaiting_invoice to active (operational)", "awaiting_invoice", "active", []string{"operational"}, true},
		{"suspended to active (operational)", "suspended", "active", []string{"operational"}, true},
		{"inactive to in_stock (operational)", "inactive", "in_stock", []string{"operational"}, true},

		// Cancelled to in_stock requires master role
		{"cancelled to in_stock (master)", "cancelled", "in_stock", []string{"master"}, true},
		{"cancelled to in_stock (operational denied)", "cancelled", "in_stock", []string{"operational"}, false},

		// Same state is always allowed (no-op)
		{"active to active", "active", "active", []string{"operational"}, true},

		// Invalid transitions
		{"active to in_existent_state", "active", "pre_cadastrada", []string{"operational"}, false},
		{"inactive to active directly (must go to in_stock)", "inactive", "active", []string{"operational"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, _, err := engine.CanTransition(EntityPhoneLine, tt.from, tt.to, tt.roles)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if allowed != tt.wantAllow {
				t.Errorf("CanTransition(%s, %s, %s, %v) = %v; want %v", EntityPhoneLine, tt.from, tt.to, tt.roles, allowed, tt.wantAllow)
			}
		})
	}
}

func TestStateMachine_CanTransition_ProcessingMonth(t *testing.T) {
	engine := NewEngine(nil)

	tests := []struct {
		name      string
		from      string
		to        string
		roles     []string
		wantAllow bool
	}{
		{"open to closed (operational)", "open", "closed", []string{"operational"}, true},
		{"open to closed (master)", "open", "closed", []string{"master"}, true},
		{"closed to open (operational denied)", "closed", "open", []string{"operational"}, false},
		{"closed to open (master allowed)", "closed", "open", []string{"master"}, true},
		{"open to cancelled (invalid)", "open", "cancelled", []string{"master"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, _, err := engine.CanTransition(EntityProcessingMonth, tt.from, tt.to, tt.roles)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if allowed != tt.wantAllow {
				t.Errorf("CanTransition(%s, %s, %s, %v) = %v; want %v", EntityProcessingMonth, tt.from, tt.to, tt.roles, allowed, tt.wantAllow)
			}
		})
	}
}

func TestStateMachine_RecordTransition(t *testing.T) {
	mock := &mockRecorder{}
	engine := NewEngine(mock)

	ctx := context.Background()
	err := engine.RecordTransition(ctx, "org1", EntityPhoneLine, "line1", "in_stock", "active", "import_invoice", nil, nil, nil)
	if err != nil {
		t.Fatalf("RecordTransition failed: %v", err)
	}

	if len(mock.recorded) != 1 || mock.recorded[0] != "phone_line:in_stock->active" {
		t.Errorf("expected recorded transition 'phone_line:in_stock->active', got %v", mock.recorded)
	}
}
