package httputil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWebhookDialRejectsNonPublicTargets(t *testing.T) {
	for _, address := range []string{
		"127.0.0.1", "10.0.0.1", "169.254.169.254", "100.64.0.1", "0.1.2.3",
		"::1", "::", "fc00::1", "fe80::1", "ff02::1", "::ffff:127.0.0.1",
		"64:ff9b::a00:1", "2002:7f00:1::", "2001::1",
	} {
		t.Run(address, func(t *testing.T) {
			dials := 0
			lookup := func(context.Context, string, string) ([]net.IP, error) {
				// Even mixed public/private DNS responses must be rejected.
				return []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP(address)}, nil
			}
			dial := publicWebhookDial(lookup, func(context.Context, string, string) (net.Conn, error) {
				dials++
				return nil, nil
			})
			for _, target := range []string{"hooks.example:443", net.JoinHostPort(address, "443")} {
				if _, err := dial(context.Background(), "tcp", target); err == nil {
					t.Fatalf("accepted unsafe target %s", target)
				}
			}
			if dials != 0 {
				t.Fatal("dialed a blocked address")
			}
		})
	}
}

func TestWebhookDialPinsIPAndRechecksDNS(t *testing.T) {
	lookups, dials := 0, 0
	lookup := func(context.Context, string, string) ([]net.IP, error) {
		lookups++
		if lookups == 1 {
			return []net.IP{net.ParseIP("2606:4700:4700::1111")}, nil
		}
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}
	dial := publicWebhookDial(lookup, func(_ context.Context, network, address string) (net.Conn, error) {
		dials++
		if network != "tcp" || address != "[2606:4700:4700::1111]:443" {
			t.Fatalf("dial must use validated IP literal, got %s %s", network, address)
		}
		client, server := net.Pipe()
		server.Close()
		return client, nil
	})
	conn, err := dial(context.Background(), "tcp", "hooks.example:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	if _, err := dial(context.Background(), "tcp", "hooks.example:443"); err == nil {
		t.Fatal("accepted rebinding to loopback")
	}
	if lookups != 2 || dials != 1 {
		t.Fatalf("unexpected lookups=%d dials=%d", lookups, dials)
	}
}

func TestWebhookClientBlocksRedirectAndPreservesHTTPS(t *testing.T) {
	requests := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Location", "http://169.254.169.254/latest/meta-data/")
		w.WriteHeader(http.StatusTemporaryRedirect)
		_, _ = io.WriteString(w, "redirect rejected")
	}))
	defer server.Close()
	lookup := func(context.Context, string, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("1.1.1.1")}, nil
	}
	dialer := &net.Dialer{}
	client := newPublicHTTPSWebhookClient(time.Second, lookup, func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "1.1.1.1:443" {
			t.Fatalf("unexpected destination: %s", address)
		}
		// The only fixture substitution is the socket. Production address checks
		// and TLS certificate/hostname verification remain enabled.
		return dialer.DialContext(ctx, network, server.Listener.Addr().String())
	})
	defer client.CloseIdleConnections()
	transport := client.Transport.(webhookTransport).transport
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	transport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	if transport.Proxy != nil || client.Timeout != time.Second {
		t.Fatal("client must bypass proxies and retain its timeout")
	}
	resp, err := client.Post("https://example.com/hook", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusTemporaryRedirect || requests != 1 {
		t.Fatalf("followed redirect: status=%d requests=%d", resp.StatusCode, requests)
	}
	if _, err := client.Get("http://example.com/hook"); err == nil {
		t.Fatal("accepted unencrypted HTTP")
	}
	if _, err := client.Get("https://user:password@example.com/hook"); err == nil {
		t.Fatal("accepted embedded credentials")
	}
	// Trusting the certificate must not disable hostname validation.
	if _, err := client.Get("https://wrong-host.example/hook"); err == nil {
		t.Fatal("accepted TLS certificate for a different hostname")
	}
}
