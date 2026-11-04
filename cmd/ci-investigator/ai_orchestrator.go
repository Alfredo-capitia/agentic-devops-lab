package main

import (
	"context"
	"errors"
	"time"
)

// EnrichWithAI enriquece o resultado determinístico com análise de IA (opcional).
// Se o provider falhar, expirar ou devolver resposta inválida, mantém o resultado determinístico
// e define ai_analysis com o status apropriado. Nunca altera deterministic_findings.
func EnrichWithAI(result InvestigationResult, investigator AIInvestigator, cfg AIConfig) InvestigationResult {
	if !cfg.Enabled || investigator == nil {
		if result.AIAnalysis.Status == "" {
			result.AIAnalysis = AIAnalysis{
				Status:          "not_requested",
				Hypotheses:      []AIHypothesis{},
				Recommendations: []string{},
			}
		}
		return result
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := AIRequest{
		Failures:    PrepareFindingsForAI(result.DeterministicFindings, cfg.MaxContextSize),
		MaxFindings: len(result.DeterministicFindings),
	}

	analysis, err := investigator.Investigate(ctx, req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			result.AIAnalysis = AIAnalysis{
				Status:          "timed_out",
				Hypotheses:      []AIHypothesis{},
				Recommendations: []string{},
			}
			return result
		}
		if errors.Is(err, ErrAINotAvailable) {
			result.AIAnalysis = AIAnalysis{
				Status:          "unavailable",
				Hypotheses:      []AIHypothesis{},
				Recommendations: []string{},
			}
			return result
		}
		result.AIAnalysis = AIAnalysis{
			Status:          "unavailable",
			Hypotheses:      []AIHypothesis{},
			Recommendations: []string{},
		}
		return result
	}

	if err := ValidateAIAnalysis(analysis); err != nil {
		result.AIAnalysis = AIAnalysis{
			Status:          "rejected",
			Hypotheses:      []AIHypothesis{},
			Recommendations: []string{},
		}
		return result
	}

	// Garante arrays não-nulos
	if analysis.Hypotheses == nil {
		analysis.Hypotheses = []AIHypothesis{}
	}
	if analysis.Recommendations == nil {
		analysis.Recommendations = []string{}
	}
	if analysis.Status == "" {
		analysis.Status = "completed"
	}

	result.AIAnalysis = *analysis
	return result
}
