package evolution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type ConnectionState struct {
	State         string
	QRCode        string
	PairingCode   string
	InstanceName  string
	OwnerJID      string
	ProfileName   string
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:  strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func (c *Client) Configured() bool {
	return c != nil && c.baseURL != "" && c.apiKey != ""
}

func (c *Client) ConnectionState(ctx context.Context, instance string) (*ConnectionState, error) {
	instance = strings.TrimSpace(instance)
	if instance == "" {
		return nil, fmt.Errorf("instância WhatsApp não informada")
	}
	var raw map[string]any
	if err := c.do(ctx, http.MethodGet, "/instance/connectionState/"+instance, nil, &raw); err != nil {
		return nil, err
	}
	out := &ConnectionState{InstanceName: instance, State: "close"}
	if inst, ok := raw["instance"].(map[string]any); ok {
		out.State = strVal(inst["state"], inst["connectionStatus"], inst["status"])
		out.OwnerJID = strVal(inst["owner"], inst["wuid"], inst["ownerJid"])
		out.ProfileName = strVal(inst["profileName"], inst["name"])
	}
	if out.State == "" {
		out.State = strVal(raw["state"], raw["status"])
	}
	if out.State == "" {
		out.State = "close"
	}
	return out, nil
}

func (c *Client) Connect(ctx context.Context, instance string) (*ConnectionState, error) {
	instance = strings.TrimSpace(instance)
	if instance == "" {
		return nil, fmt.Errorf("instância WhatsApp não informada")
	}
	_ = c.ensureInstance(ctx, instance)
	var raw map[string]any
	if err := c.do(ctx, http.MethodGet, "/instance/connect/"+instance, nil, &raw); err != nil {
		return nil, err
	}
	out := &ConnectionState{InstanceName: instance, State: "connecting"}
	out.QRCode = firstQR(raw)
	out.PairingCode = strVal(raw["pairingCode"], raw["pairing_code"])
	if inst, ok := raw["instance"].(map[string]any); ok {
		if s := strVal(inst["state"], inst["status"]); s != "" {
			out.State = s
		}
	}
	if s := strVal(raw["state"], raw["status"]); s != "" {
		out.State = s
	}
	return out, nil
}

func (c *Client) Disconnect(ctx context.Context, instance string) error {
	instance = strings.TrimSpace(instance)
	if instance == "" {
		return fmt.Errorf("instância WhatsApp não informada")
	}
	if err := c.do(ctx, http.MethodDelete, "/instance/logout/"+instance, nil, nil); err != nil {
		_ = c.do(ctx, http.MethodDelete, "/instance/logoutInstance/"+instance, nil, nil)
	}
	return nil
}

func (c *Client) SetWebhook(ctx context.Context, instance, webhookURL string) error {
	instance = strings.TrimSpace(instance)
	webhookURL = strings.TrimSpace(webhookURL)
	if instance == "" || webhookURL == "" {
		return nil
	}
	body := map[string]any{
		"enabled":  true,
		"url":      webhookURL,
		"webhookByEvents": false,
		"webhookBase64":   true,
		"events":   []string{"MESSAGES_UPSERT", "CONNECTION_UPDATE"},
		"webhook": map[string]any{
			"enabled": true,
			"url":     webhookURL,
			"byEvents": false,
			"base64":  true,
			"events":  []string{"MESSAGES_UPSERT", "CONNECTION_UPDATE"},
		},
	}
	if err := c.do(ctx, http.MethodPost, "/webhook/set/"+instance, body, nil); err != nil {
		return c.do(ctx, http.MethodPost, "/webhook/instance/"+instance, map[string]any{
			"url": webhookURL, "enabled": true, "events": []string{"MESSAGES_UPSERT"},
		}, nil)
	}
	return nil
}

func (c *Client) ensureInstance(ctx context.Context, instance string) error {
	body := map[string]any{
		"instanceName": instance,
		"qrcode":       true,
		"integration":  "WHATSAPP-BAILEYS",
	}
	err := c.do(ctx, http.MethodPost, "/instance/create", body, nil)
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "already") || strings.Contains(msg, "exist") || strings.Contains(msg, "403") {
		return nil
	}
	return err
}

func (c *Client) do(ctx context.Context, method, path string, body any, dest any) error {
	if !c.Configured() {
		return fmt.Errorf("Evolution API não configurada")
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Evolution API: %w", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = res.Status
		}
		return fmt.Errorf("Evolution API %d: %s", res.StatusCode, truncate(msg, 400))
	}
	if dest == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return fmt.Errorf("Evolution API: resposta inválida")
	}
	return nil
}

func firstQR(raw map[string]any) string {
	candidates := []any{
		raw["base64"], raw["qrcode"], raw["code"], raw["qr"],
	}
	if q, ok := raw["qrcode"].(map[string]any); ok {
		candidates = append(candidates, q["base64"], q["code"])
	}
	if q, ok := raw["instance"].(map[string]any); ok {
		candidates = append(candidates, q["qrcode"], q["base64"])
	}
	for _, c := range candidates {
		s := strings.TrimSpace(fmt.Sprint(c))
		if s == "" || s == "<nil>" {
			continue
		}
		if strings.HasPrefix(s, "data:image") {
			return s
		}
		if len(s) > 80 {
			return "data:image/png;base64," + strings.TrimPrefix(s, "data:image/png;base64,")
		}
	}
	return ""
}

func strVal(vals ...any) string {
	for _, v := range vals {
		s := strings.TrimSpace(fmt.Sprint(v))
		if s != "" && s != "<nil>" {
			return s
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
