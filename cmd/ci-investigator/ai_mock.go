package main

import (
	"context"
)

// MockAIInvestigator é um provider mock para desenvolvimento e testes.
// Nunca faz chamadas externas.
type MockAIInvestigator struct{}

// NewMockAIInvestigator cria uma instância do mock provider.
func NewMockAIInvestigator() *MockAIInvestigator {
	return &MockAIInvestigator{}
}

func (m *MockAIInvestigator) Name() string {
	return "mock"
}

func (m *MockAIInvestigator) Investigate(ctx context.Context, req AIRequest) (*AIAnalysis, error) {
	// Retorna análise vazia/completable por padrão.
	// Em fase posterior será usada para testes de integração.
	analysis := &AIAnalysis{
		Status:          "completed",
		Hypotheses:      []AIHypothesis{},
		Recommendations: []string{},
	}
	return analysis, nil
}

// NoopAIInvestigator não faz nada e retorna not_requested.
type NoopAIInvestigator struct{}

func NewNoopAIInvestigator() *NoopAIInvestigator {
	return &NoopAIInvestigator{}
}

func (n *NoopAIInvestigator) Name() string {
	return "noop"
}

func (n *NoopAIInvestigator) Investigate(ctx context.Context, req AIRequest) (*AIAnalysis, error) {
	return &AIAnalysis{
		Status:          "not_requested",
		Hypotheses:      []AIHypothesis{},
		Recommendations: []string{},
	}, nil
}
