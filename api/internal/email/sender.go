package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/invoicelayout"
)

type Message struct {
	To      string
	Subject string
	HTML    string
}

type Sender struct {
	cfg config.Config
}

func NewSender(cfg config.Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Enabled() bool {
	return strings.TrimSpace(s.cfg.SMTPHost) != ""
}

func (s *Sender) Send(ctx context.Context, msg Message) error {
	if !s.Enabled() {
		return fmt.Errorf("smtp is not configured")
	}
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return fmt.Errorf("recipient email is required")
	}
	from := strings.TrimSpace(s.cfg.SMTPFrom)
	if from == "" {
		from = strings.TrimSpace(s.cfg.SMTPUser)
	}
	if from == "" {
		return fmt.Errorf("smtp from address is required")
	}

	body := buildMIMEBody(from, to, msg.Subject, msg.HTML)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	var auth smtp.Auth
	if strings.TrimSpace(s.cfg.SMTPUser) != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	}

	if s.cfg.SMTPTLS {
		return s.sendTLS(ctx, addr, auth, from, to, body)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(body))
}

func (s *Sender) sendTLS(_ context.Context, addr string, auth smtp.Auth, from, to, body string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(body)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildMIMEBody(from, to, subject, html string) string {
	wrapped := invoicelayout.EnsureHTMLDocument(html)

	var encodedBody bytes.Buffer
	qp := quotedprintable.NewWriter(&encodedBody)
	_, _ = qp.Write([]byte(wrapped))
	_ = qp.Close()

	var buf bytes.Buffer
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + to + "\r\n")
	buf.WriteString("Subject: " + encodeSubject(subject) + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buf.WriteString("\r\n")
	buf.Write(encodedBody.Bytes())
	return buf.String()
}

func encodeSubject(subject string) string {
	if subject == "" {
		return subject
	}
	for _, r := range subject {
		if r > 127 {
			return mime.QEncoding.Encode("utf-8", subject)
		}
	}
	return subject
}
