package observability

import (
	"testing"
	"time"
)

func TestRecorderSnapshot(t *testing.T) {
	r := NewRecorder(100)
	r.Observe("import", 40*time.Millisecond, true)
	r.Observe("import", 80*time.Millisecond, false)
	r.Observe("query", 10*time.Millisecond, true)
	r.Observe("db.ping", 5*time.Millisecond, true)
	r.Observe("processing.pipeline", 12*time.Millisecond, true)
	snap := r.Snapshot()
	if snap.SampleCount != 5 {
		t.Fatalf("expected 5 samples, got %d", snap.SampleCount)
	}
	if snap.DBPing == nil || snap.DBPing.Count != 1 {
		t.Fatal("expected db_ping stats")
	}
	if snap.Pipeline == nil {
		t.Fatal("expected pipeline stats")
	}
}
