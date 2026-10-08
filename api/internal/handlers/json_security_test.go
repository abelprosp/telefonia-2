package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsTrailingValues(t *testing.T) {
	for _, input := range []string{`{"name":"safe"} {"name":"other"}`, `{"name":"safe"} garbage`} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(input))
		var body struct {
			Name string `json:"name"`
		}
		if decodeJSON(r, &body) == nil {
			t.Errorf("accepted trailing content: %s", input)
		}
	}
}
