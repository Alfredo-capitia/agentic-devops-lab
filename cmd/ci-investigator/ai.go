package main

import (
	"context"
	"errors"
	"time"
)

// AIRequest contém apenas o contexto mínimo e necessário para análise pela IA.
// Nunca deve conter logs brutos não filtrados ou dados sensíveis.
type AIRequest struct {
	Failures    []AIRequestFailure `json:"failures"`
	CIContext   map[string]string  `json:"ci_context,omitempty"`
	MaxFindings int                `json:"max_findings,omitempty"`
}

// AIRequestFailure representa uma falha determinística simplificada para envio à IA.
type AIRequestFailure struct {
	Category   string          `json:"category"`
	FailedStep string          `json:"failed_step"`
	Evidence   string          `json:"evidence"`
	Message    string          `json:"message,omitempty"`
	Location   *SourceLocation `json:"location,omitempty"`
	Confidence string          `json:"confidence,omitempty"`
}

// AIInvestigator define a interface independente de provider para análise de IA.
type AIInvestigator interface {
	Name() string
	Investigate(ctx context.Context, req AIRequest) (*AIAnalysis, error)
}

// AIConfig define configurações da camada de IA.
type AIConfig struct {
	Enabled        bool
	Provider       string
	Timeout        time.Duration
	MaxContextSize int
}

// ErrAINotAvailable indica que o provider de IA não está disponível.
var ErrAINotAvailable = errors.New("ai provider not available")

// ErrAIInvalidResponse indica que a resposta da IA não passou na validação.
var ErrAIInvalidResponse = errors.New("ai provider returned invalid response")
