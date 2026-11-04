package main

import (
	"strings"
	"unicode/utf8"
)

const (
	DefaultMaxContextSize = 64 * 1024 // 64KB
	DefaultMaxFindings    = 20
)

// TruncateString trunca string para tamanho máximo em bytes (aproximado).
func TruncateString(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	// Trunca por bytes, mas tenta manter válido UTF-8
	truncated := s[:maxBytes]
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated + "\n... [TRUNCATED]"
}

// SanitizeForAI aplica sanitização básica contra prompt injection.
// Remove instruções suspeitas mantendo o conteúdo factual.
func SanitizeForAI(s string) string {
	if s == "" {
		return s
	}
	out := s
	// Neutraliza tentativas comuns de injeção
	suspicious := []string{
		"ignore previous instructions",
		"ignore all previous",
		"system:",
		"assistant:",
		"jailbreak",
		"act as",
		"you must",
	}
	lower := strings.ToLower(out)
	for _, p := range suspicious {
		if strings.Contains(lower, p) {
			// Substitui de forma conservadora - mantém contexto mas marca
			out = strings.ReplaceAll(out, p, "[FILTERED]")
			out = strings.ReplaceAll(out, strings.ToUpper(p), "[FILTERED]")
			// título
			if len(p) > 0 {
				out = strings.ReplaceAll(out, strings.Title(p), "[FILTERED]")
			}
			lower = strings.ToLower(out)
		}
	}
	return out
}

// PrepareFindingsForAI prepara findings para envio: redige, sanitiza e trunca.
func PrepareFindingsForAI(findings []Report, maxBytes int) []AIRequestFailure {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxContextSize
	}
	if len(findings) == 0 {
		return []AIRequestFailure{}
	}

	// Limita número de findings
	if len(findings) > DefaultMaxFindings {
		findings = findings[:DefaultMaxFindings]
	}

	out := make([]AIRequestFailure, 0, len(findings))
	total := 0
	for _, f := range findings {
		fr := RedactReport(f)
		fr.Evidence = SanitizeForAI(fr.Evidence)
		fr.Message = SanitizeForAI(fr.Message)
		fr.Cause = SanitizeForAI(fr.Cause)
		fr.Remediation = SanitizeForAI(fr.Remediation)

		// Estima tamanho
		est := len(fr.Evidence) + len(fr.Message) + len(fr.Cause)
		if total+est > maxBytes && len(out) > 0 {
			break // não adiciona mais para respeitar limite
		}
		total += est

		out = append(out, AIRequestFailure{
			Category:   fr.Category,
			FailedStep: fr.FailedStep,
			Evidence:   TruncateString(fr.Evidence, 4096),
			Message:    TruncateString(fr.Message, 1024),
			Location:   fr.Location,
			Confidence: fr.Confidence,
		})
	}
	return out
}
