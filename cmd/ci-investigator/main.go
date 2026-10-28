package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
)

type Report struct {
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	FailedStep  string `json:"failed_step"`
	Evidence    string `json:"evidence"`
	Cause       string `json:"probable_cause"`
	Remediation string `json:"recommended_action"`
	Confidence  string `json:"confidence"`
}

var assertionMismatch = regexp.MustCompile(
	`([\w./-]+\.go):(\d+): expected "([^"]*)", got "([^"]*)"`,
)

func analyze(log string) Report {
	match := assertionMismatch.FindStringSubmatch(log)

	if match != nil {
		return Report{
			Category:   "test_assertion_mismatch",
			Severity:   "medium",
			FailedStep: "Run tests",
			Evidence: fmt.Sprintf("%s:%s: expected %q, got %q",
				match[1], match[2], match[3], match[4]),
			Cause:       "The test expectation differs from the actual value.",
			Remediation: "Review the test expectation and application behavior before changing code.",
			Confidence:  "high",
		}
	}

	return Report{
		Category:    "unknown_failure",
		Severity:    "undetermined",
		FailedStep:  "unknown",
		Evidence:    "No recognized test assertion mismatch was found.",
		Cause:       "The available log does not provide enough evidence for classification.",
		Remediation: "Inspect the complete CI logs and identify the failing step.",
		Confidence:  "low",
	}
}

func main() {
	log, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read input:", err)
		os.Exit(1)
	}

	report := analyze(string(log))

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, "encode report:", err)
		os.Exit(1)
	}
}
