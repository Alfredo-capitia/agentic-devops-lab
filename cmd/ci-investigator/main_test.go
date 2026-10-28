package main

import (
	"strings"
	"testing"
)

func TestAnalyzeAssertionMismatch(t *testing.T) {
	log := `main_test.go:10: expected "CI should fail", got "CI baseline ready"`

	report := analyze(log)

	if report.Category != "test_assertion_mismatch" {
		t.Errorf("Category = %q, want %q",
			report.Category, "test_assertion_mismatch")
	}

	if report.Severity != "medium" {
		t.Errorf("Severity = %q, want %q",
			report.Severity, "medium")
	}

	if report.FailedStep != "Run tests" {
		t.Errorf("FailedStep = %q, want %q",
			report.FailedStep, "Run tests")
	}

	if report.Confidence != "high" {
		t.Errorf("Confidence = %q, want %q",
			report.Confidence, "high")
	}

	if !strings.Contains(report.Evidence, "main_test.go:10") {
		t.Errorf("Evidence lacks source location: %q", report.Evidence)
	}

	if !strings.Contains(report.Evidence, "CI should fail") {
		t.Errorf("Evidence lacks expected value: %q", report.Evidence)
	}

	if !strings.Contains(report.Evidence, "CI baseline ready") {
		t.Errorf("Evidence lacks actual value: %q", report.Evidence)
	}

	if report.Cause == "" {
		t.Error("Cause must not be empty")
	}

	if report.Remediation == "" {
		t.Error("Remediation must not be empty")
	}
}

func TestAnalyzeAssertionMismatchWithNestedPath(t *testing.T) {
	log := `internal/checks/example_test.go:42: expected "healthy", got "unhealthy"`

	report := analyze(log)

	if report.Category != "test_assertion_mismatch" {
		t.Fatalf("Category = %q, want test_assertion_mismatch",
			report.Category)
	}

	if !strings.Contains(report.Evidence, "internal/checks/example_test.go:42") {
		t.Errorf("Evidence lost source location: %q", report.Evidence)
	}
}

func TestAnalyzeUnknownFailure(t *testing.T) {
	cases := []struct {
		name string
		log  string
	}{
		{name: "connection timeout", log: "ERROR: connection timeout"},
		{name: "empty log", log: ""},
		{name: "whitespace only", log: " \n\t "},
		{name: "unrecognized compiler output", log: "unexpected build output"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := analyze(tc.log)

			if report.Category != "unknown_failure" {
				t.Errorf("Category = %q, want %q",
					report.Category, "unknown_failure")
			}

			if report.Confidence != "low" {
				t.Errorf("Confidence = %q, want %q",
					report.Confidence, "low")
			}

			if report.Cause == "" {
				t.Error("Cause must not be empty")
			}

			if report.Remediation == "" {
				t.Error("Remediation must not be empty")
			}
		})
	}
}

func TestAnalyzeUsesFirstRecognizedMismatch(t *testing.T) {
	log := `first_test.go:10: expected "one", got "two"
second_test.go:20: expected "three", got "four"`

	report := analyze(log)

	if !strings.Contains(report.Evidence, "first_test.go:10") {
		t.Errorf("Expected first mismatch, got evidence: %q", report.Evidence)
	}

	if strings.Contains(report.Evidence, "second_test.go:20") {
		t.Errorf("Expected only the first mismatch, got evidence: %q",
			report.Evidence)
	}
}

func TestAnalyzeCompileFailure(t *testing.T) {
	log := `# example.com/app
./main.go:12:2: undefined: missingFunction`

	report := analyze(log)

	if report.Category != "compile_failure" {
		t.Fatalf("Category = %q, want %q",
			report.Category, "compile_failure")
	}

	if report.Confidence != "high" {
		t.Errorf("Confidence = %q, want %q",
			report.Confidence, "high")
	}

	if !strings.Contains(report.Evidence, "undefined: missingFunction") {
		t.Errorf("Evidence lacks compiler error: %q", report.Evidence)
	}

	if report.Cause == "" {
		t.Error("Cause must not be empty")
	}

	if report.Remediation == "" {
		t.Error("Remediation must not be empty")
	}
}

func TestAnalyzeDoesNotMisclassifyGenericError(t *testing.T) {
	report := analyze("ERROR: connection timeout")

	if report.Category != "unknown_failure" {
		t.Errorf("Category = %q, want %q",
			report.Category, "unknown_failure")
	}
}
