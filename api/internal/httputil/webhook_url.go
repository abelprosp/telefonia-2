package httputil

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// ValidatePublicHTTPSWebhookURL rejects non-HTTPS URLs and hosts that resolve
// to loopback/private/link-local addresses (SSRF protection).
func ValidatePublicHTTPSWebhookURL(raw string) error {
	raw = strings.TrimSpace(raw)
	u, err := parseWebhookURL(raw)
	if err != nil {
		return err
	}
	host := strings.ToLower(u.Hostname())
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("URL do webhook não pode apontar para rede privada")
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return fmt.Errorf("não foi possível resolver o host do webhook")
	}
	return validateWebhookIPs(ips)
}

func parseWebhookURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("URL do webhook inválida")
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return nil, fmt.Errorf("URL do webhook deve usar HTTPS")
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return nil, fmt.Errorf("URL do webhook sem host")
	}
	if u.User != nil || strings.Contains(host, "%") || strings.TrimSuffix(host, ".") == "localhost" {
		return nil, fmt.Errorf("URL do webhook não pode conter credenciais ou host local")
	}
	return u, nil
}

func validateWebhookIPs(ips []net.IP) error {
	if len(ips) == 0 {
		return fmt.Errorf("host do webhook sem endereço")
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("URL do webhook não pode apontar para rede privada")
		}
	}
	return nil
}

func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() ||
		!ip.IsGlobalUnicast() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() ||
		isSpecialWebhookIP(ip)
}
