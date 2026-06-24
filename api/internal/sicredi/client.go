package sicredi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	cfg    Config
	http   *http.Client
	mu     sync.Mutex
	tokens *tokenState
}

type tokenState struct {
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	refreshUntil time.Time
}

type Pagador struct {
	TipoPessoa string
	Documento  string
	Nome       string
	Endereco   string
	Cidade     string
	UF         string
	CEP        string
	Email      string
}

type CreateBoletoInput struct {
	SeuNumero       string
	IdTituloEmpresa string
	DataVencimento  time.Time
	Valor           float64
	Mensagens       []string
	Pagador         Pagador
}

type BoletoResult struct {
	NossoNumero    string
	LinhaDigitavel string
	CodigoBarras   string
	PixQrCode      string
	PixTxID        string
}

type BoletoDetail struct {
	NossoNumero    string
	Situacao       string
	DataPagamento  *time.Time
	ValorLiquidado float64
}

type LiquidadoItem struct {
	NossoNumero     string
	SeuNumero       string
	DataPagamento   time.Time
	ValorLiquidado  float64
	TipoLiquidacao  string
}

type LiquidadosPage struct {
	Items   []LiquidadoItem
	HasNext bool
}

func IsSituacaoLiquidada(situacao string) bool {
	s := strings.ToUpper(strings.TrimSpace(situacao))
	return strings.Contains(s, "LIQUIDADO")
}

type apiError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c.cfg.EnabledAndConfigured()
}

// Ping validates OAuth credentials against the configured environment (sandbox or production).
func (c *Client) Ping(ctx context.Context) error {
	if !c.Enabled() {
		return fmt.Errorf("sicredi não configurado")
	}
	_, err := c.accessToken(ctx)
	return err
}

type WebhookContract struct {
	URL           string   `json:"url"`
	Eventos       []string `json:"eventos"`
	ContratoStatus string  `json:"contratoStatus"`
}

func (c *Client) ListWebhookContracts(ctx context.Context) ([]WebhookContract, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("sicredi não configurado")
	}
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("cooperativa", c.cfg.Cooperativa)
	q.Set("posto", c.cfg.Posto)
	q.Set("beneficiario", c.cfg.CodigoBeneficiario)
	reqURL := c.cfg.WebhookContratosURL() + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.setBoletoHeaders(req, token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseHTTPError(resp.StatusCode, body)
	}
	var out struct {
		Items []WebhookContract `json:"items"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		// Some responses return a single contract object.
		var single WebhookContract
		if err2 := json.Unmarshal(body, &single); err2 == nil && single.URL != "" {
			return []WebhookContract{single}, nil
		}
		return nil, fmt.Errorf("resposta inválida do Sicredi: %w", err)
	}
	return out.Items, nil
}

func (c *Client) Config() Config {
	return c.cfg
}

func (c *Client) CreateHybridBoleto(ctx context.Context, input CreateBoletoInput) (*BoletoResult, error) {
	return c.createBoleto(ctx, input, "HIBRIDO")
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if c.tokens != nil && c.tokens.accessToken != "" && now.Before(c.tokens.expiresAt.Add(-15*time.Second)) {
		return c.tokens.accessToken, nil
	}
	if c.tokens != nil && c.tokens.refreshToken != "" && now.Before(c.tokens.refreshUntil) {
		if err := c.refresh(ctx); err == nil {
			return c.tokens.accessToken, nil
		}
	}
	if err := c.authenticate(ctx); err != nil {
		return "", err
	}
	return c.tokens.accessToken, nil
}

func (c *Client) authenticate(ctx context.Context) error {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", c.cfg.Username)
	form.Set("password", c.cfg.Password)
	form.Set("scope", "cobranca")
	return c.tokenRequest(ctx, form)
}

func (c *Client) refresh(ctx context.Context) error {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", c.tokens.refreshToken)
	return c.tokenRequest(ctx, form)
}

func (c *Client) tokenRequest(ctx context.Context, form url.Values) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.AuthURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("x-api-key", c.cfg.APIKey)
	req.Header.Set("context", "COBRANCA")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseHTTPError(resp.StatusCode, body)
	}

	var out struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("token Sicredi inválido: %w", err)
	}
	now := time.Now()
	c.tokens = &tokenState{
		accessToken:  out.AccessToken,
		refreshToken: out.RefreshToken,
		expiresAt:    now.Add(time.Duration(out.ExpiresIn) * time.Second),
		refreshUntil: now.Add(time.Duration(out.RefreshExpiresIn) * time.Second),
	}
	return nil
}

func parseHTTPError(status int, body []byte) error {
	var errBody apiError
	_ = json.Unmarshal(body, &errBody)
	msg := strings.TrimSpace(errBody.Message)
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}
	if msg == "" {
		msg = http.StatusText(status)
	}
	return fmt.Errorf("sicredi HTTP %d: %s", status, msg)
}

func (c *Client) GetBoleto(ctx context.Context, nossoNumero string) (*BoletoDetail, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("sicredi não configurado")
	}
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("codigoBeneficiario", c.cfg.CodigoBeneficiario)
	q.Set("nossoNumero", strings.TrimSpace(nossoNumero))
	reqURL := c.cfg.BoletoURL() + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.setBoletoHeaders(req, token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseHTTPError(resp.StatusCode, body)
	}

	var out struct {
		NossoNumero    string  `json:"nossoNumero"`
		Situacao       string  `json:"situacao"`
		DataPagamento  *string `json:"dataPagamento"`
		ValorLiquidado float64 `json:"valorLiquidado"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("resposta inválida do Sicredi: %w", err)
	}
	detail := &BoletoDetail{
		NossoNumero:    out.NossoNumero,
		Situacao:       out.Situacao,
		ValorLiquidado: out.ValorLiquidado,
	}
	if out.DataPagamento != nil && strings.TrimSpace(*out.DataPagamento) != "" {
		if t, err := time.Parse("2006-01-02", strings.TrimSpace(*out.DataPagamento)); err == nil {
			detail.DataPagamento = &t
		}
	}
	return detail, nil
}

func (c *Client) ListLiquidadosDia(ctx context.Context, day time.Time, page int) (*LiquidadosPage, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("sicredi não configurado")
	}
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("codigoBeneficiario", c.cfg.CodigoBeneficiario)
	q.Set("dia", day.Format("02/01/2006"))
	if page > 0 {
		q.Set("pagina", fmt.Sprintf("%d", page))
	}
	reqURL := c.cfg.LiquidadosDiaURL() + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.setBoletoHeaders(req, token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseHTTPError(resp.StatusCode, body)
	}

	var out struct {
		Items []struct {
			NossoNumero    string  `json:"nossoNumero"`
			SeuNumero      string  `json:"seuNumero"`
			DataPagamento  string  `json:"dataPagamento"`
			ValorLiquidado float64 `json:"valorLiquidado"`
			TipoLiquidacao string  `json:"tipoLiquidacao"`
		} `json:"items"`
		HasNext bool `json:"hasNext"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("resposta inválida do Sicredi: %w", err)
	}
	pageResult := &LiquidadosPage{HasNext: out.HasNext}
	for _, item := range out.Items {
		paidAt, _ := time.Parse("2006-01-02", strings.TrimSpace(item.DataPagamento))
		pageResult.Items = append(pageResult.Items, LiquidadoItem{
			NossoNumero:    strings.TrimSpace(item.NossoNumero),
			SeuNumero:      item.SeuNumero,
			DataPagamento:  paidAt,
			ValorLiquidado: item.ValorLiquidado,
			TipoLiquidacao: item.TipoLiquidacao,
		})
	}
	return pageResult, nil
}

func (c *Client) setBoletoHeaders(req *http.Request, token string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.cfg.APIKey)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("context", "COBRANCA")
	req.Header.Set("cooperativa", c.cfg.Cooperativa)
	req.Header.Set("posto", c.cfg.Posto)
}

func (c *Client) invalidateTokens() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens = nil
}

func (c *Client) do(ctx context.Context, method, reqURL string, body []byte) (*http.Response, error) {
	return c.doWithRetry(ctx, method, reqURL, body, false)
}

func (c *Client) doWithRetry(ctx context.Context, method, reqURL string, body []byte, retried bool) (*http.Response, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, err
	}
	c.setBoletoHeaders(req, token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && !retried {
		resp.Body.Close()
		c.invalidateTokens()
		return c.doWithRetry(ctx, method, reqURL, body, true)
	}
	return resp, nil
}

func (c *Client) CreateTraditionalBoleto(ctx context.Context, input CreateBoletoInput) (*BoletoResult, error) {
	return c.createBoleto(ctx, input, "NORMAL")
}

func (c *Client) createBoleto(ctx context.Context, input CreateBoletoInput, tipoCobranca string) (*BoletoResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("sicredi não configurado")
	}
	pagador := map[string]any{
		"tipoPessoa": input.Pagador.TipoPessoa,
		"documento":  input.Pagador.Documento,
		"nome":       truncate(input.Pagador.Nome, 40),
	}
	if addr := strings.TrimSpace(input.Pagador.Endereco); addr != "" {
		pagador["endereco"] = truncate(addr, 40)
	}
	if city := strings.TrimSpace(input.Pagador.Cidade); city != "" {
		pagador["cidade"] = truncate(city, 25)
	}
	if uf := strings.TrimSpace(input.Pagador.UF); uf != "" {
		pagador["uf"] = truncate(uf, 2)
	}
	if cep := normalizeDigits(input.Pagador.CEP); cep != "" {
		pagador["cep"] = cep
	}
	if email := strings.TrimSpace(input.Pagador.Email); email != "" {
		pagador["email"] = truncate(email, 40)
	}

	bodyMap := map[string]any{
		"tipoCobranca":       tipoCobranca,
		"codigoBeneficiario": c.cfg.CodigoBeneficiario,
		"dataVencimento":     input.DataVencimento.Format("2006-01-02"),
		"especieDocumento":   "DUPLICATA_MERCANTIL_INDICACAO",
		"seuNumero":          truncate(input.SeuNumero, 10),
		"idTituloEmpresa":    truncate(input.IdTituloEmpresa, 25),
		"valor":              input.Valor,
		"pagador":            pagador,
	}
	if len(input.Mensagens) > 0 {
		bodyMap["mensagens"] = input.Mensagens
	}
	raw, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodPost, c.cfg.BoletoURL(), raw)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseHTTPError(resp.StatusCode, respBody)
	}
	var out struct {
		NossoNumero    string `json:"nossoNumero"`
		LinhaDigitavel string `json:"linhaDigitavel"`
		CodigoBarras   string `json:"codigoBarras"`
		QrCode         string `json:"qrCode"`
		TxID           string `json:"txid"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("resposta inválida do Sicredi: %w", err)
	}
	return &BoletoResult{
		NossoNumero:    out.NossoNumero,
		LinhaDigitavel: out.LinhaDigitavel,
		CodigoBarras:   out.CodigoBarras,
		PixQrCode:      out.QrCode,
		PixTxID:        out.TxID,
	}, nil
}

func (c *Client) GetBoletoPDF(ctx context.Context, linhaDigitavel string) ([]byte, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("sicredi não configurado")
	}
	q := url.Values{}
	q.Set("linhaDigitavel", strings.TrimSpace(linhaDigitavel))
	resp, err := c.do(ctx, http.MethodGet, c.cfg.PdfURL()+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseHTTPError(resp.StatusCode, data)
	}
	return data, nil
}

func (c *Client) CancelBoleto(ctx context.Context, nossoNumero string) error {
	if !c.Enabled() {
		return fmt.Errorf("sicredi não configurado")
	}
	raw, err := json.Marshal(map[string]any{"codigoBeneficiario": c.cfg.CodigoBeneficiario})
	if err != nil {
		return err
	}
	url := c.cfg.BoletoURL() + "/" + strings.TrimSpace(nossoNumero) + "/baixa"
	resp, err := c.do(ctx, http.MethodPatch, url, raw)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseHTTPError(resp.StatusCode, body)
	}
	return nil
}

func (c *Client) AlterBoletoDueDate(ctx context.Context, nossoNumero string, dueDate time.Time) error {
	if !c.Enabled() {
		return fmt.Errorf("sicredi não configurado")
	}
	raw, err := json.Marshal(map[string]any{
		"codigoBeneficiario": c.cfg.CodigoBeneficiario,
		"dataVencimento":     dueDate.Format("2006-01-02"),
	})
	if err != nil {
		return err
	}
	resp, err := c.do(ctx, http.MethodPatch, c.cfg.BoletoByNossoNumeroURL(nossoNumero), raw)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseHTTPError(resp.StatusCode, body)
	}
	return nil
}

type WebhookContractInput struct {
	URL     string
	Eventos []string
	Token   string
}

func (c *Client) RegisterWebhookContract(ctx context.Context, input WebhookContractInput) error {
	if !c.Enabled() {
		return fmt.Errorf("sicredi não configurado")
	}
	eventos := input.Eventos
	if len(eventos) == 0 {
		eventos = []string{"LIQUIDACAO", "BAIXA", "REGISTRO"}
	}
	payload := map[string]any{
		"cooperativa":      c.cfg.Cooperativa,
		"posto":            c.cfg.Posto,
		"codBeneficiario":  c.cfg.CodigoBeneficiario,
		"url":              input.URL,
		"eventos":          eventos,
		"contratoStatus":   "ATIVO",
		"enviarIdTituloEmpresa": true,
	}
	if token := strings.TrimSpace(input.Token); token != "" {
		payload["token"] = token
		payload["header"] = "Authorization"
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := c.do(ctx, http.MethodPost, c.cfg.WebhookContratoURL(), raw)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseHTTPError(resp.StatusCode, body)
	}
	return nil
}

func (c *Client) GetBoletoBySeuNumero(ctx context.Context, seuNumero string) (*BoletoDetail, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("sicredi não configurado")
	}
	q := url.Values{}
	q.Set("codigoBeneficiario", c.cfg.CodigoBeneficiario)
	q.Set("seuNumero", strings.TrimSpace(seuNumero))
	resp, err := c.do(ctx, http.MethodGet, c.cfg.BoletoURL()+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseHTTPError(resp.StatusCode, body)
	}
	var out struct {
		NossoNumero    string  `json:"nossoNumero"`
		Situacao       string  `json:"situacao"`
		DataPagamento  *string `json:"dataPagamento"`
		ValorLiquidado float64 `json:"valorLiquidado"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("resposta inválida do Sicredi: %w", err)
	}
	detail := &BoletoDetail{
		NossoNumero:    out.NossoNumero,
		Situacao:       out.Situacao,
		ValorLiquidado: out.ValorLiquidado,
	}
	if out.DataPagamento != nil && strings.TrimSpace(*out.DataPagamento) != "" {
		if t, err := time.Parse("2006-01-02", strings.TrimSpace(*out.DataPagamento)); err == nil {
			detail.DataPagamento = &t
		}
	}
	return detail, nil
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func normalizeDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
