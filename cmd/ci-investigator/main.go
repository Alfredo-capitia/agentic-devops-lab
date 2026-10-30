package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
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

var compileFailure = regexp.MustCompile(
	`(?m)^(\./?[\w./-]+\.go:\d+:\d+:\s+.+)`,
)

func sectionLog(log, section string) (string, bool) {
	marker := "=== " + section + " ==="
	start := strings.Index(log, marker)

	if start == -1 {
		return "", false
	}

	content := log[start+len(marker):]

	if end := strings.Index(content, "=== "); end >= 0 {
		content = content[:end]
	}

	return strings.TrimSpace(content), true
}

func analyze(log string) Report {
	formatLog, hasFormat := sectionLog(log, "FORMATTING")
	if hasFormat &&
		formatLog != "" &&
		!strings.Contains(formatLog, "All Go files are formatted.") {
		return Report{
			Category:    "format_failure",
			Severity:    "medium",
			FailedStep:  "Check formatting",
			Evidence:    formatLog,
			Cause:       "One or more Go files are not correctly formatted.",
			Remediation: "Run gofmt on the reported files and rerun CI.",
			Confidence:  "high",
		}
	}

	vetLog, hasVet := sectionLog(log, "GO VET")
	if hasVet && vetLog != "" {
		return Report{
			Category:    "vet_failure",
			Severity:    "medium",
			FailedStep:  "Run go vet",
			Evidence:    vetLog,
			Cause:       "Go vet reported a potential code issue.",
			Remediation: "Review the diagnostic and correct the underlying issue.",
			Confidence:  "high",
		}
	}

	testLog, hasTests := sectionLog(log, "TESTS")
	if !hasTests {
		testLog = log
	}

	match := assertionMismatch.FindStringSubmatch(testLog)
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

	buildLog, hasBuild := sectionLog(log, "BUILD")
	if !hasBuild {
		buildLog = log
	}

	compileMatch := compileFailure.FindString(buildLog)
	if compileMatch != "" {
		return Report{
			Category:    "compile_failure",
			Severity:    "high",
			FailedStep:  "Build application",
			Evidence:    compileMatch,
			Cause:       "The Go compiler reported a source-level compilation error.",
			Remediation: "Inspect the reported file and line, correct the compilation issue, and rerun the tests.",
			Confidence:  "high",
		}
	}

	return Report{
		Category:    "unknown_failure",
		Severity:    "undetermined",
		FailedStep:  "unknown",
		Evidence:    "No recognized failure pattern was found.",
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
