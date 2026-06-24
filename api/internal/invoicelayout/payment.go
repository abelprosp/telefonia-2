package invoicelayout

import (
	"encoding/base64"
	"fmt"
	"html"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

type PaymentData struct {
	LinhaDigitavel      string
	CodigoBarras        string
	PixCopyPaste        string
	PixQrCodeDataURL    string
	NossoNumero         string
}

func FormatLinhaDigitavel(raw string) string {
	digits := onlyDigits(raw)
	if len(digits) != 47 {
		return raw
	}
	return fmt.Sprintf("%s.%s %s.%s %s.%s %s %s",
		digits[0:5], digits[5:10],
		digits[10:15], digits[15:21],
		digits[21:26], digits[26:32],
		digits[32:33], digits[33:47])
}

func PixQRCodeDataURL(pixCopyPaste string) string {
	pixCopyPaste = strings.TrimSpace(pixCopyPaste)
	if pixCopyPaste == "" {
		return ""
	}
	png, err := qrcode.Encode(pixCopyPaste, qrcode.Medium, 220)
	if err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
}

func RenderPaymentSection(data PaymentData, theme Theme) string {
	if strings.TrimSpace(data.LinhaDigitavel) == "" && strings.TrimSpace(data.PixCopyPaste) == "" {
		return ""
	}
	border := theme.BorderColor
	if border == "" {
		border = "#222222"
	}
	accent := theme.AccentColor
	if accent == "" {
		accent = "#00a0c6"
	}

	var b strings.Builder
	b.WriteString(`<div style="border:1px solid `)
	b.WriteString(border)
	b.WriteString(`;border-radius:12px;padding:16px;margin-top:12px;background:#fafafa;">`)
	b.WriteString(`<div style="font-size:16px;font-weight:700;margin-bottom:12px;color:`)
	b.WriteString(accent)
	b.WriteString(`;">Pagamento — Boleto / PIX</div>`)

	if data.NossoNumero != "" {
		b.WriteString(`<div style="margin-bottom:8px;font-size:12px;"><strong>Nosso número:</strong> `)
		b.WriteString(html.EscapeString(data.NossoNumero))
		b.WriteString(`</div>`)
	}
	if data.LinhaDigitavel != "" {
		formatted := FormatLinhaDigitavel(data.LinhaDigitavel)
		b.WriteString(`<div style="margin-bottom:10px;"><div style="font-weight:700;font-size:12px;margin-bottom:4px;">Linha digitável</div>`)
		b.WriteString(`<div style="font-family:monospace;font-size:13px;letter-spacing:0.5px;word-break:break-all;">`)
		b.WriteString(html.EscapeString(formatted))
		b.WriteString(`</div></div>`)
	}
	if data.CodigoBarras != "" {
		b.WriteString(`<div style="margin-bottom:10px;"><div style="font-weight:700;font-size:12px;margin-bottom:4px;">Código de barras</div>`)
		b.WriteString(`<div style="font-family:monospace;font-size:12px;word-break:break-all;">`)
		b.WriteString(html.EscapeString(data.CodigoBarras))
		b.WriteString(`</div></div>`)
	}
	if data.PixCopyPaste != "" {
		b.WriteString(`<div style="margin-top:12px;padding-top:12px;border-top:1px dashed `)
		b.WriteString(border)
		b.WriteString(`;"><div style="font-weight:700;font-size:12px;margin-bottom:8px;">PIX — copia e cola</div>`)
		if data.PixQrCodeDataURL != "" {
			b.WriteString(`<div style="text-align:center;margin-bottom:10px;"><img src="`)
			b.WriteString(html.EscapeString(data.PixQrCodeDataURL))
			b.WriteString(`" alt="QR Code PIX" width="220" height="220" style="max-width:220px;height:auto;" /></div>`)
		}
		b.WriteString(`<div style="font-family:monospace;font-size:11px;word-break:break-all;background:#fff;border:1px solid `)
		b.WriteString(border)
		b.WriteString(`;padding:8px;border-radius:6px;">`)
		b.WriteString(html.EscapeString(data.PixCopyPaste))
		b.WriteString(`</div></div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func AppendPaymentSection(htmlBody string, payment PaymentData, theme Theme) string {
	section := RenderPaymentSection(payment, theme)
	if section == "" {
		return htmlBody
	}
	if strings.Contains(htmlBody, "</div>") {
		return strings.TrimSpace(htmlBody) + section
	}
	return htmlBody + section
}
