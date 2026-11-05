package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// HTTPAIProvider chama um endpoint HTTP compatível com o contrato de análise.
// Requer configuração explícita via variáveis de ambiente:
//
//	CI_INVESTIGATOR_AI_ENDPOINT (obrigatório quando provider=http)
//	CI_INVESTIGATOR_AI_API_KEY (opcional, enviado como Bearer token)
//	CI_INVESTIGATOR_AI_MODEL (opcional, incluído no payload)
type HTTPAIProvider struct {
	Endpoint string
	APIKey   string
	Model    string
	Client   *http.Client
}

func NewHTTPAIProvider() (*HTTPAIProvider, error) {
	endpoint := strings.TrimSpace(os.Getenv("CI_INVESTIGATOR_AI_ENDPOINT"))
	if endpoint == "" {
		return nil, ErrAINotAvailable
	}
	return &HTTPAIProvider{
		Endpoint: endpoint,
		APIKey:   strings.TrimSpace(os.Getenv("CI_INVESTIGATOR_AI_API_KEY")),
		Model:    strings.TrimSpace(os.Getenv("CI_INVESTIGATOR_AI_MODEL")),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (h *HTTPAIProvider) Name() string { return "http" }

type httpAIRequest struct {
	Model  string                 `json:"model,omitempty"`
	Input  AIRequest              `json:"input"`
	Schema map[string]interface{} `json:"schema_hint,omitempty"`
}

func (h *HTTPAIProvider) Investigate(ctx context.Context, req AIRequest) (*AIAnalysis, error) {
	if h.Endpoint == "" {
		return nil, ErrAINotAvailable
	}

	payload := httpAIRequest{
		Model: h.Model,
		Input: req,
		Schema: map[string]interface{}{
			"ai_analysis": map[string]string{
				"status": "completed|unavailable|timed_out|rejected",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, h.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if h.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+h.APIKey)
	}

	resp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: status %d", ErrAINotAvailable, resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, err
	}

	// Aceita resposta direta como AIAnalysis ou embrulhada em {"ai_analysis":...}
	var direct AIAnalysis
	if err := json.Unmarshal(data, &direct); err == nil && direct.Status != "" {
		if err := ValidateAIAnalysis(&direct); err != nil {
			return nil, err
		}
		return &direct, nil
	}

	var wrapped struct {
		AIAnalysis AIAnalysis `json:"ai_analysis"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.AIAnalysis.Status != "" {
		if err := ValidateAIAnalysis(&wrapped.AIAnalysis); err != nil {
			return nil, err
		}
		return &wrapped.AIAnalysis, nil
	}

	return nil, ErrAIInvalidResponse
}
