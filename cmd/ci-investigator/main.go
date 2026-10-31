package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type SourceLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`
}

type Report struct {
	Category    string          `json:"category"`
	Severity    string          `json:"severity"`
	FailedStep  string          `json:"failed_step"`
	Evidence    string          `json:"evidence"`
	Location    *SourceLocation `json:"location,omitempty"`
	Message     string          `json:"message,omitempty"`
	Cause       string          `json:"probable_cause"`
	Remediation string          `json:"recommended_action"`
	Confidence  string          `json:"confidence"`
}

type Investigation struct {
	Failures []Report `json:"failures"`
}

var assertionMismatch = regexp.MustCompile(
	`([\w./-]+\.go):(\d+): expected "([^"]*)", got "([^"]*)"`,
)

var vetDiagnostic = regexp.MustCompile(
	`^(.+\.go):(\d+):(\d+):\s+(.+)$`,
)

var compileFailure = regexp.MustCompile(
	`^((?:\./)?[\w./-]+\.go):(\d+):(\d+):\s+(.+)$`,
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

func unknownSection(log string) string {
	knownSections := map[string]bool{
		"FORMATTING": true,
		"GO VET":     true,
		"TESTS":      true,
		"BUILD":      true,
	}

	for _, line := range strings.Split(log, "\n") {
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "=== ") ||
			!strings.HasSuffix(line, " ===") {
			continue
		}

		section := strings.TrimSpace(
			strings.TrimSuffix(
				strings.TrimPrefix(line, "=== "),
				" ===",
			),
		)

		if section != "" && !knownSections[section] {
			return section
		}
	}

	return "unknown"
}

func analyze(log string) Investigation {
	var failures []Report

	hasSections := strings.Contains(log, "=== ")

	// Formatting
	// Formatting
	formatLog, hasFormat := sectionLog(log, "FORMATTING")
	if hasFormat &&
		formatLog != "" &&
		!strings.Contains(formatLog, "All Go files are formatted.") {

		formatFiles := strings.Split(formatLog, "\n")

		for _, file := range formatFiles {
			file = strings.TrimSpace(file)

			if file == "" {
				continue
			}

			failures = append(failures, Report{
				Category:   "format_failure",
				Severity:   "medium",
				FailedStep: "Check formatting",
				Evidence:   file,
				Location: &SourceLocation{
					File: file,
				},
				Message:     "File is not formatted with gofmt.",
				Cause:       "One or more Go files are not correctly formatted.",
				Remediation: "Run gofmt on the reported file and rerun CI.",
				Confidence:  "high",
			})
		}
	}

	// Go vet
	// Go vet
	vetLog, hasVet := sectionLog(log, "GO VET")
	if hasVet && vetLog != "" {
		for _, line := range strings.Split(vetLog, "\n") {
			diagnostic := strings.TrimSpace(line)

			match := vetDiagnostic.FindStringSubmatch(diagnostic)
			if match == nil {
				continue
			}

			lineNumber, err := strconv.Atoi(match[2])
			if err != nil {
				continue
			}

			columnNumber, err := strconv.Atoi(match[3])
			if err != nil {
				continue
			}

			failures = append(failures, Report{
				Category:   "vet_failure",
				Severity:   "medium",
				FailedStep: "Run go vet",
				Evidence:   diagnostic,
				Location: &SourceLocation{
					File:   match[1],
					Line:   lineNumber,
					Column: columnNumber,
				},
				Message:     match[4],
				Cause:       "Go vet reported a potential code issue.",
				Remediation: "Review the diagnostic and correct the underlying issue.",
				Confidence:  "high",
			})
		}
	}

	// Tests
	testLog, hasTests := sectionLog(log, "TESTS")

	if !hasTests && !hasSections {
		testLog = log
	}

	if hasTests || !hasSections {
		matches := assertionMismatch.FindAllStringSubmatch(testLog, -1)

		for _, match := range matches {
			lineNumber, err := strconv.Atoi(match[2])
			if err != nil {
				continue
			}

			message := fmt.Sprintf(
				"expected %q, got %q",
				match[3],
				match[4],
			)

			failures = append(failures, Report{
				Category:   "test_assertion_mismatch",
				Severity:   "medium",
				FailedStep: "Run tests",
				Evidence: fmt.Sprintf(
					"%s:%s: %s",
					match[1],
					match[2],
					message,
				),
				Location: &SourceLocation{
					File: match[1],
					Line: lineNumber,
				},
				Message:     message,
				Cause:       "The test expectation differs from the actual value.",
				Remediation: "Review the test expectation and application behavior before changing code.",
				Confidence:  "high",
			})
		}
	}

	// Build
	// Build
	buildLog, hasBuild := sectionLog(log, "BUILD")

	if !hasBuild && !hasSections {
		buildLog = log
	}

	if hasBuild || !hasSections {
		for _, line := range strings.Split(buildLog, "\n") {
			diagnostic := strings.TrimSpace(line)

			match := compileFailure.FindStringSubmatch(diagnostic)
			if match == nil {
				continue
			}

			lineNumber, err := strconv.Atoi(match[2])
			if err != nil {
				continue
			}

			columnNumber, err := strconv.Atoi(match[3])
			if err != nil {
				continue
			}

			failures = append(failures, Report{
				Category:   "compile_failure",
				Severity:   "high",
				FailedStep: "Build application",
				Evidence:   diagnostic,
				Location: &SourceLocation{
					File:   match[1],
					Line:   lineNumber,
					Column: columnNumber,
				},
				Message:     match[4],
				Cause:       "The Go compiler reported a source-level compilation error.",
				Remediation: "Inspect the reported file and line, correct the compilation issue, and rerun the tests.",
				Confidence:  "high",
			})
		}
	}

	// Unknown failure
	// Unknown failure
	if len(failures) == 0 {
		evidence := strings.TrimSpace(log)

		if evidence == "" {
			evidence = "No CI log output was provided."
		}

		failures = append(failures, Report{
			Category:    "unknown_failure",
			Severity:    "undetermined",
			FailedStep:  unknownSection(log),
			Evidence:    evidence,
			Cause:       "The available log does not provide enough evidence for classification.",
			Remediation: "Inspect the complete CI logs and identify the failing step.",
			Confidence:  "low",
		})
	}
	return Investigation{
		Failures: failures,
	}
}

func main() {
	log, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read input:", err)
		os.Exit(1)
	}

	investigation := analyze(string(log))

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(investigation); err != nil {
		fmt.Fprintln(os.Stderr, "encode report:", err)
		os.Exit(1)
	}
}
