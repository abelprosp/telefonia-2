package zapsign

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Config struct {
	APIToken string
	BaseURL  string // Padrão: "https://api.zapsign.com.br" ou "https://sandbox.api.zapsign.com.br"
	Sandbox  bool
}

type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		if cfg.Sandbox {
			baseURL = "https://sandbox.api.zapsign.com.br"
		} else {
			baseURL = "https://api.zapsign.com.br"
		}
	}
	return &Client{
		token:   cfg.APIToken,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.token != ""
}

type SignerPosition struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Page    int     `json:"page"`
	Zindex  int     `json:"zindex,omitempty"`
}

type CreateSignerRequest struct {
	Name                   string           `json:"name"`
	Email                  string           `json:"email,omitempty"`
	PhoneCountry           string           `json:"phone_country,omitempty"`
	PhoneNumber            string           `json:"phone_number,omitempty"`
	AuthMode               string           `json:"auth_mode,omitempty"` // "email", "whatsapp", "sms", "assinaturaTela"
	SendAutomaticEmail     bool             `json:"send_automatic_email,omitempty"`
	SendAutomaticWhatsApp  bool             `json:"send_automatic_whatsapp,omitempty"`
	Positions              []SignerPosition `json:"positions,omitempty"`
	RequireCPF             bool             `json:"require_cpf,omitempty"`
	RequireSelfiePhoto     bool             `json:"require_selfie_photo,omitempty"`
	RequireDocumentPhoto   bool             `json:"require_document_photo,omitempty"`
}

type CreateDocRequest struct {
	Name            string                `json:"name"`
	URLPdf          string                `json:"url_pdf,omitempty"`
	Base64Pdf       string                `json:"base64_pdf,omitempty"`
	Signers         []CreateSignerRequest `json:"signers"`
	Lang            string                `json:"lang,omitempty"`
	BrandName       string                `json:"brand_name,omitempty"`
	BrandLogo       string                `json:"brand_logo,omitempty"`
	ExternalID      string                `json:"external_id,omitempty"`
	FolderPath      string                `json:"folder_path,omitempty"`
	SignatureType   string                `json:"signature_type,omitempty"` // "electronic" ou "digital"
}

type SignerResponse struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	SignURL  string `json:"sign_url"`
	Status   string `json:"status"` // "pending", "signed", etc.
}

type DocResponse struct {
	OpenID     int64            `json:"open_id"`
	Token      string           `json:"token"`
	Name       string           `json:"name"`
	Status     string           `json:"status"` // "pending", "signed", "rejected"
	SignedFile *string          `json:"signed_file"`
	OriginalFile string         `json:"original_file"`
	Signers    []SignerResponse `json:"signers"`
	CreatedAt  string           `json:"created_at"`
	LastUpdateAt string         `json:"last_update_at"`
}

func (c *Client) CreateDocument(ctx context.Context, req CreateDocRequest) (*DocResponse, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("zapsign integration is not configured (missing API token)")
	}

	if req.Lang == "" {
		req.Lang = "pt-br"
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/docs/", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send zapsign request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("zapsign error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var docResp DocResponse
	if err := json.Unmarshal(bodyBytes, &docResp); err != nil {
		return nil, fmt.Errorf("unmarshal zapsign response: %w", err)
	}

	return &docResp, nil
}

func (c *Client) GetDocument(ctx context.Context, docToken string) (*DocResponse, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("zapsign integration is not configured")
	}

	url := fmt.Sprintf("%s/api/v1/docs/%s/", c.baseURL, docToken)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send zapsign get request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("zapsign get error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var docResp DocResponse
	if err := json.Unmarshal(bodyBytes, &docResp); err != nil {
		return nil, fmt.Errorf("unmarshal zapsign get response: %w", err)
	}

	return &docResp, nil
}
