# Architecture

```
┌──────────────────────┐
│   CI Pipeline        │  (GitHub Actions, etc.)
│  - logs, exit codes  │
└─────────┬────────────┘
          │
          v
┌──────────────────────┐
│ Evidence Collector   │  Aggregates sectioned logs (ci.log)
│  - preserve evidence │  Applies size limits/redaction before external use
└─────────┬────────────┘
          │
          v
┌──────────────────────────────────────────────┐
│ Deterministic Investigation Engine            │
│  - known patterns (format/vet/test/build)    │
│  - structured locations + confidence         │
│  - fallback to unknown_failure               │
└─────────┬────────────────────────────────────┘
          │ deterministic_findings
          v
┌──────────────────────────────────────────────┐
│ AI Investigation Layer (optional)             │
│  - disabled by default                       │
│  - validates responses, links evidence       │
│  - timeout/unavailable/rejected handling     │
└─────────┬────────────────────────────────────┘
          │ ai_analysis
          v
┌──────────────────────────────────────────────┐
│ Structured Investigation Report (JSON)        │
│  deterministic_findings + ai_analysis         │
└──────────────────────────────────────────────┘
```
