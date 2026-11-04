package main

import (
	"regexp"
)

// redactionPatterns contém padrões para mascarar dados sensíveis em logs/contexto.
var redactionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(ghp_[A-Za-z0-9]{36,})`),
	regexp.MustCompile(`(?i)(github_pat_[A-Za-z0-9_]{22,})`),
	regexp.MustCompile(`(?i)(xox[baprs]-[A-Za-z0-9-]+)`),
	regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16})`),
	regexp.MustCompile(`(?i)(ASIA[0-9A-Z]{16})`),
	regexp.MustCompile(`(?i)(aws_secret_access_key|secret_access_key)\s*[:=]\s*[^\s]+`),
	regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*[^\s]+`),
	regexp.MustCompile(`(?i)(token|api[_-]?key|access[_-]?key)\s*[:=]\s*[^\s]+`),
	regexp.MustCompile(`(?i)(bearer\s+[A-Za-z0-9\-\._~\+\/]+=*)`),
	regexp.MustCompile(`(?i)(-----BEGIN (RSA|EC|OPENSSH|PGP) PRIVATE KEY-----[\s\S]+?-----END (RSA|EC|OPENSSH|PGP) PRIVATE KEY-----)`),
	regexp.MustCompile(`(?i)([A-Za-z0-9+/]{40,}={0,2})`), // base64 longo (heurístico)
}

// RedactSecrets remove/mascara dados sensíveis de texto.
func RedactSecrets(s string) string {
	if s == "" {
		return s
	}
	out := s
	for _, re := range redactionPatterns {
		out = re.ReplaceAllStringFunc(out, func(m string) string {
			if len(m) <= 8 {
				return "[REDACTED]"
			}
			// Preserva prefixo pequeno para debugging sem expor segredo
			return m[:4] + "..." + "[REDACTED]"
		})
	}
	return out
}

// RedactReport aplica redação a um Report antes de enviar à IA.
func RedactReport(r Report) Report {
	r.Evidence = RedactSecrets(r.Evidence)
	r.Message = RedactSecrets(r.Message)
	r.Cause = RedactSecrets(r.Cause)
	r.Remediation = RedactSecrets(r.Remediation)
	if r.Location != nil {
		// Location não contém segredos normalmente, mas por segurança não alteramos estrutura
	}
	return r
}

// RedactFindings aplica redação a todos os findings.
func RedactFindings(findings []Report) []Report {
	if len(findings) == 0 {
		return findings
	}
	out := make([]Report, 0, len(findings))
	for _, f := range findings {
		out = append(out, RedactReport(f))
	}
	return out
}
