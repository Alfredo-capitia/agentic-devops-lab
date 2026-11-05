package main

import (
	"fmt"
	"io"
	"strings"
)

// WriteMarkdown escreve um relatório legível em Markdown.
func WriteMarkdown(w io.Writer, r InvestigationResult) error {
	fmt.Fprintln(w, "# CI Investigation Report")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- Deterministic findings: **%d**\n", len(r.DeterministicFindings))
	fmt.Fprintf(w, "- AI status: **%s**\n\n", r.AIAnalysis.Status)

	if len(r.DeterministicFindings) == 0 {
		fmt.Fprintln(w, "No deterministic failures found.")
	} else {
		fmt.Fprintln(w, "## Deterministic findings")
		fmt.Fprintln(w)
		for i, f := range r.DeterministicFindings {
			fmt.Fprintf(w, "### %d. %s (%s)\n\n", i+1, f.Category, f.Severity)
			fmt.Fprintf(w, "- **Failed step:** %s\n", f.FailedStep)
			if f.Location != nil && f.Location.File != "" {
				loc := f.Location.File
				if f.Location.Line > 0 {
					loc = fmt.Sprintf("%s:%d", loc, f.Location.Line)
					if f.Location.Column > 0 {
						loc = fmt.Sprintf("%s:%d", loc, f.Location.Column)
					}
				}
				fmt.Fprintf(w, "- **Location:** `%s`\n", loc)
			}
			if f.Message != "" {
				fmt.Fprintf(w, "- **Message:** %s\n", f.Message)
			}
			if f.Cause != "" {
				fmt.Fprintf(w, "- **Probable cause:** %s\n", f.Cause)
			}
			if f.Remediation != "" {
				fmt.Fprintf(w, "- **Recommended action:** %s\n", f.Remediation)
			}
			fmt.Fprintf(w, "- **Confidence:** %s\n", f.Confidence)
			if f.Evidence != "" {
				fmt.Fprintln(w, "\n**Evidence:**")
				fmt.Fprintln(w, "```")
				fmt.Fprintln(w, strings.TrimSpace(f.Evidence))
				fmt.Fprintln(w, "```")
			}
			fmt.Fprintln(w)
		}
	}

	if r.AIAnalysis.Status != "" && r.AIAnalysis.Status != "not_requested" {
		fmt.Fprintln(w, "## AI analysis")
		fmt.Fprintln(w)
		fmt.Fprintf(w, "- **Status:** %s\n\n", r.AIAnalysis.Status)
		if len(r.AIAnalysis.Hypotheses) > 0 {
			fmt.Fprintln(w, "### Hypotheses")
			for _, h := range r.AIAnalysis.Hypotheses {
				fmt.Fprintf(w, "- %s", h.Description)
				if h.Confidence != "" {
					fmt.Fprintf(w, " *(confidence: %s)*", h.Confidence)
				}
				fmt.Fprintln(w)
				for _, e := range h.Evidence {
					fmt.Fprintf(w, "  - evidence: %s\n", e)
				}
			}
			fmt.Fprintln(w)
		}
		if len(r.AIAnalysis.Recommendations) > 0 {
			fmt.Fprintln(w, "### Recommendations")
			for _, rec := range r.AIAnalysis.Recommendations {
				fmt.Fprintf(w, "- %s\n", rec)
			}
		}
	}

	return nil
}
