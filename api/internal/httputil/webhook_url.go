package httputil

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidatePublicHTTPSWebhookURL rejects non-HTTPS URLs and hosts that resolve
// to loopback/private/link-local addresses (SSRF protection).
func ValidatePublicHTTPSWebhookURL(raw string) error {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return fmt.Errorf("URL do webhook inválida")
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return fmt.Errorf("URL do webhook deve usar HTTPS")
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return fmt.Errorf("URL do webhook sem host")
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "0.0.0.0" {
		return fmt.Errorf("URL do webhook não pode apontar para localhost")
	}
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("URL do webhook não pode apontar para rede privada")
		}
		return nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("não foi possível resolver o host do webhook")
	}
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
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified()
}
