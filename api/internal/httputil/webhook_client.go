package httputil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"time"
)

// These non-public and translation ranges can otherwise bypass IsPrivate,
// including IPv6 tunnels/translation that can route back into IPv4 networks.
var webhookDeniedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("3fff::/20"),
}

func isSpecialWebhookIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return true
	}
	addr = addr.Unmap()
	for _, prefix := range webhookDeniedPrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

type webhookLookup func(context.Context, string, string) ([]net.IP, error)
type webhookDial func(context.Context, string, string) (net.Conn, error)

// NewPublicHTTPSWebhookClient checks every new connection, pins the dial to a
// validated IP, preserves TLS hostname verification, and never follows redirects
// or environment proxies. The timeout includes DNS, TLS and response reads.
func NewPublicHTTPSWebhookClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}
	return newPublicHTTPSWebhookClient(timeout, net.DefaultResolver.LookupIP, dialer.DialContext)
}

func newPublicHTTPSWebhookClient(timeout time.Duration, lookup webhookLookup, dial webhookDial) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = publicWebhookDial(lookup, dial)
	return &http.Client{
		Timeout:       timeout,
		Transport:     webhookTransport{transport},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

type webhookTransport struct{ transport *http.Transport }

func (t webhookTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if _, err := parseWebhookURL(req.URL.String()); err != nil {
		return nil, err
	}
	return t.transport.RoundTrip(req)
}

func (t webhookTransport) CloseIdleConnections() { t.transport.CloseIdleConnections() }

func publicWebhookDial(lookup webhookLookup, dial webhookDial) webhookDial {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips := []net.IP{net.ParseIP(host)}
		if ips[0] == nil {
			ips, err = lookup(ctx, "ip", host)
			if err != nil {
				return nil, fmt.Errorf("não foi possível resolver o host do webhook: %w", err)
			}
		}
		if err := validateWebhookIPs(ips); err != nil {
			return nil, err
		}
		for _, ip := range ips {
			// Dial the validated literal, never the hostname (DNS rebinding).
			conn, dialErr := dial(ctx, network, net.JoinHostPort(ip.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			err = dialErr
		}
		return nil, err
	}
}
