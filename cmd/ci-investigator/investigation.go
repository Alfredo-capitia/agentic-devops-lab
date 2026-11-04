package main

type InvestigationResult struct {
	DeterministicFindings []Report   `json:"deterministic_findings"`
	AIAnalysis            AIAnalysis `json:"ai_analysis"`
}

type AIAnalysis struct {
	Status          string         `json:"status"`
	Hypotheses      []AIHypothesis `json:"hypotheses"`
	Recommendations []string       `json:"recommendations"`
}

type AIHypothesis struct {
	Description string   `json:"description"`
	Evidence    []string `json:"evidence"`
	Confidence  string   `json:"confidence"`
}
