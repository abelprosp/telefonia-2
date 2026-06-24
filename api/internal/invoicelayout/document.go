package invoicelayout

import "strings"

// WrapHTMLDocument envolve o fragmento HTML em um documento completo com charset UTF-8.
// Necessário para e-mail e exportação em PDF reconhecerem acentuação corretamente.
func WrapHTMLDocument(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return `<!DOCTYPE html><html lang="pt-BR"><head><meta charset="UTF-8"></head><body></body></html>`
	}
	return `<!DOCTYPE html><html lang="pt-BR"><head><meta charset="UTF-8"><meta http-equiv="Content-Type" content="text/html; charset=UTF-8"></head><body style="margin:0;padding:16px;background:#ffffff;">` + body + `</body></html>`
}

// EnsureHTMLDocument garante documento HTML completo antes do envio por e-mail.
func EnsureHTMLDocument(html string) string {
	lower := strings.ToLower(html)
	if strings.Contains(lower, "<html") {
		if strings.Contains(lower, "charset") {
			return html
		}
		// Documento sem charset explícito — injeta meta após <head>.
		if idx := strings.Index(lower, "<head>"); idx >= 0 {
			insertAt := idx + len("<head>")
			return html[:insertAt] + `<meta charset="UTF-8"><meta http-equiv="Content-Type" content="text/html; charset=UTF-8">` + html[insertAt:]
		}
		return WrapHTMLDocument(html)
	}
	return WrapHTMLDocument(html)
}
