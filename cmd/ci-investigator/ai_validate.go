package main

import (
	"fmt"
	"strings"
)

// ValidateAIAnalysis valida rigorosamente a resposta da IA.
// Garante que hipóteses estão ligadas a evidências e estrutura é válida.
func ValidateAIAnalysis(analysis *AIAnalysis) error {
	if analysis == nil {
		return fmt.Errorf("%w: nil analysis", ErrAIInvalidResponse)
	}

	switch analysis.Status {
	case "completed", "not_requested", "unavailable", "timed_out", "rejected":
		// status válido
	default:
		return fmt.Errorf("%w: invalid status %q", ErrAIInvalidResponse, analysis.Status)
	}

	if analysis.Hypotheses == nil {
		analysis.Hypotheses = []AIHypothesis{}
	}
	if analysis.Recommendations == nil {
		analysis.Recommendations = []string{}
	}

	for i, h := range analysis.Hypotheses {
		if strings.TrimSpace(h.Description) == "" {
			return fmt.Errorf("%w: hypothesis %d has empty description", ErrAIInvalidResponse, i)
		}
		if len(h.Evidence) == 0 {
			return fmt.Errorf("%w: hypothesis %d must be linked to at least one evidence", ErrAIInvalidResponse, i)
		}
		for _, e := range h.Evidence {
			if strings.TrimSpace(e) == "" {
				return fmt.Errorf("%w: hypothesis %d has empty evidence", ErrAIInvalidResponse, i)
			}
		}
		if h.Confidence != "" {
			switch h.Confidence {
			case "low", "medium", "high":
				// ok
			default:
				return fmt.Errorf("%w: hypothesis %d has invalid confidence %q", ErrAIInvalidResponse, i, h.Confidence)
			}
		}
	}

	return nil
}

// BuildAIRequest constrói um pedido mínimo à IA a partir dos achados determinísticos.
func BuildAIRequest(findings []Report) AIRequest {
	req := AIRequest{
		Failures:    make([]AIRequestFailure, 0, len(findings)),
		MaxFindings: len(findings),
	}

	for _, f := range findings {
		req.Failures = append(req.Failures, AIRequestFailure{
			Category:   f.Category,
			FailedStep: f.FailedStep,
			Evidence:   f.Evidence,
			Message:    f.Message,
			Location:   f.Location,
			Confidence: f.Confidence,
		})
	}

	return req
}
