# Agentic DevOps Lab

A hands-on lab for learning GitHub Agentic Workflows,
AI-powered automation, CI troubleshooting, and DevSecOps.

## Learning objectives

- Understand agentic workflows in GitHub Actions.
- Investigate CI failures with AI-assisted analysis.
- Apply least privilege and secure automation practices.
- Test agent behavior before allowing write access.

## Initial safety policy

The first agent will be report-only.
It must not modify application code, merge pull requests,
deploy software, or access production credentials.

## Status

- [x] GitHub CLI authenticated
- [x] SSH configured for Git operations
- [x] Remote repository created
- [ ] AI engine selected
- [ ] First agentic workflow configured

## CI Investigator (deterministic + AI-ready)

The `cmd/ci-investigator` tool reads CI logs from stdin and produces a structured report (JSON by default, or `--format markdown`).

### Output contract (v2)
```json
{
  "deterministic_findings": [
    {
      "category": "test_failure|format_failure|vet_failure|compile_failure|test_assertion_mismatch|unknown_failure",
      "severity": "medium|high|undetermined",
      "failed_step": "Check formatting|Run go vet|Run tests|Build application|...",
      "evidence": "...",
      "location": {"file": "...", "line": 123, "column": 4},
      "message": "...",
      "probable_cause": "...",
      "recommended_action": "...",
      "confidence": "high|low"
    }
  ],
  "ai_analysis": {
    "status": "not_requested|completed|unavailable|timed_out|rejected",
    "hypotheses": [
      {"description": "...", "evidence": ["..."], "confidence": "high|medium|low"}
    ],
    "recommendations": ["..."]
  }
}
```

### AI layer (optional, disabled by default)
The AI layer is **complementary** and never replaces deterministic findings. It is disabled by default.

Environment variables:
- `CI_INVESTIGATOR_AI_ENABLED` (true/false)
- `CI_INVESTIGATOR_AI_PROVIDER` (noop|mock) 
- `CI_INVESTIGATOR_AI_TIMEOUT` (e.g. `30s`)
- `CI_INVESTIGATOR_AI_MAX_CONTEXT` (bytes)

Safety guarantees:
- Deterministic findings are never modified by AI
- Secrets are redacted before sending context to AI
- Context is size-limited and sanitized
- Timeouts and errors fall back gracefully
- No autonomous command execution or code changes
