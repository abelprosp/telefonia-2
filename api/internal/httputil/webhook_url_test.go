package httputil

import "testing"

func TestValidatePublicHTTPSWebhookURL(t *testing.T) {
	if err := ValidatePublicHTTPSWebhookURL("https://1.1.1.1/hooks"); err != nil {
		t.Fatalf("expected public https ok, got %v", err)
	}
	cases := []string{
		"http://hooks.example.com/path",
		"https://localhost/hook",
		"https://127.0.0.1/hook",
		"https://10.0.0.5/hook",
		"https://192.168.1.10/hook",
		"ftp://example.com/hook",
		"",
	}
	for _, raw := range cases {
		if err := ValidatePublicHTTPSWebhookURL(raw); err == nil {
			t.Fatalf("expected rejection for %q", raw)
		}
	}
}
