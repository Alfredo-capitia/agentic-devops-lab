package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadAIConfig carrega configuração da IA a partir de variáveis de ambiente.
// Por defeito, IA está desativada (report-only / sem envio externo não solicitado).
func LoadAIConfig() AIConfig {
	cfg := AIConfig{
		Enabled:        false,
		Provider:       strings.TrimSpace(os.Getenv("CI_INVESTIGATOR_AI_PROVIDER")),
		Timeout:        30 * time.Second,
		MaxContextSize: DefaultMaxContextSize,
	}

	if v := strings.TrimSpace(os.Getenv("CI_INVESTIGATOR_AI_ENABLED")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Enabled = b
		}
	}
	if v := strings.TrimSpace(os.Getenv("CI_INVESTIGATOR_AI_TIMEOUT")); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Timeout = d
		}
	}
	if v := strings.TrimSpace(os.Getenv("CI_INVESTIGATOR_AI_MAX_CONTEXT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxContextSize = n
		}
	}

	return cfg
}

// NewAIInvestigator cria o provider com base na configuração.
// Apenas retorna provider se explicitamente habilitado e provider reconhecido.
// Por segurança, default é noop.
func NewAIInvestigator(cfg AIConfig) AIInvestigator {
	if !cfg.Enabled {
		return NewNoopAIInvestigator()
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "mock":
		return NewMockAIInvestigator()
	case "http":
		p, err := NewHTTPAIProvider()
		if err != nil {
			return NewNoopAIInvestigator()
		}
		return p
	case "noop", "":
		return NewNoopAIInvestigator()
	default:
		// Provider desconhecido -> não habilita IA automaticamente
		return NewNoopAIInvestigator()
	}
}
