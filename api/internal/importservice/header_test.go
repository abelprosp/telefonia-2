package importservice

import (
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/vivo"
)

func TestGetHeaderMissingReturnsNil(t *testing.T) {
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("getHeader panicked: %v", rec)
		}
	}()
	if h := getHeader(nil); h != nil {
		t.Fatalf("expected nil header, got %+v", h)
	}
	if h := getHeader([]any{"not a header"}); h != nil {
		t.Fatalf("expected nil header for unrelated records, got %+v", h)
	}
}

func TestGetHeaderFinds010D(t *testing.T) {
	want := &vivo.Line010DHeader{}
	got := getHeader([]any{"x", want, "y"})
	if got != want {
		t.Fatalf("expected header pointer, got %+v", got)
	}
}
