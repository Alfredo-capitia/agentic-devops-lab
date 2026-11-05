# Enabling AI analysis in GitHub Actions (example)

AI is **disabled by default**. Enable only with explicit configuration and ensure secrets/permissions are handled safely.

```yaml
- name: Investigate CI failure (with AI)
  if: always() && failure()
  env:
    CI_INVESTIGATOR_AI_ENABLED: "true"
    CI_INVESTIGATOR_AI_PROVIDER: "mock" # replace with a real provider when implemented
    CI_INVESTIGATOR_AI_TIMEOUT: "30s"
  run: |
    go run ./cmd/ci-investigator < ci.log > ci-investigation.json
```

Notes:
- AI must never change the CI result. Keep `Enforce CI result` based on original step outcomes.
- Do not send sensitive logs externally without explicit approval and redaction.
- If AI is unavailable, timed out, or rejected, the report still contains `deterministic_findings`.
