package main

import (
	"strings"
	"testing"
)

func requireSingleFailure(t *testing.T, investigation Investigation) Report {
	t.Helper()

	if len(investigation.Failures) != 1 {
		t.Fatalf("got %d failures, want 1", len(investigation.Failures))
	}

	return investigation.Failures[0]
}

func findFailure(t *testing.T, investigation Investigation, category string) Report {
	t.Helper()

	for _, failure := range investigation.Failures {
		if failure.Category == category {
			return failure
		}
	}

	t.Fatalf("failure category %q not found", category)
	return Report{}
}

func TestAnalyzeAssertionMismatch(t *testing.T) {
	log := `main_test.go:10: expected "CI should fail", got "CI baseline ready"`

	report := requireSingleFailure(t, analyze(log))

	if report.Category != "test_assertion_mismatch" {
		t.Errorf("Category = %q, want %q",
			report.Category, "test_assertion_mismatch")
	}

	if report.Severity != "medium" {
		t.Errorf("Severity = %q, want medium", report.Severity)
	}

	if report.FailedStep != "Run tests" {
		t.Errorf("FailedStep = %q, want Run tests", report.FailedStep)
	}

	if report.Confidence != "high" {
		t.Errorf("Confidence = %q, want high", report.Confidence)
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

	report := requireSingleFailure(t, analyze(log))

	if report.Category != "test_assertion_mismatch" {
		t.Fatalf("Category = %q, want test_assertion_mismatch",
			report.Category)
	}

	if !strings.Contains(report.Evidence, "internal/checks/example_test.go:42") {
		t.Errorf("Evidence lost source location: %q", report.Evidence)
	}
}

func TestAnalyzeUsesFirstRecognizedMismatch(t *testing.T) {
	log := `first_test.go:10: expected "one", got "two"
second_test.go:20: expected "three", got "four"`

	report := requireSingleFailure(t, analyze(log))

	if !strings.Contains(report.Evidence, "first_test.go:10") {
		t.Errorf("Expected first mismatch, got evidence: %q", report.Evidence)
	}

	if strings.Contains(report.Evidence, "second_test.go:20") {
		t.Errorf("Expected only the first mismatch, got evidence: %q",
			report.Evidence)
	}
}

func TestAnalyzeFormattingFailure(t *testing.T) {
	log := `=== FORMATTING ===
main.go`

	report := requireSingleFailure(t, analyze(log))

	if report.Category != "format_failure" {
		t.Errorf("Category = %q, want format_failure", report.Category)
	}

	if report.FailedStep != "Check formatting" {
		t.Errorf("FailedStep = %q, want Check formatting", report.FailedStep)
	}
}

func TestAnalyzeVetFailure(t *testing.T) {
	log := `=== GO VET ===
main.go:10:2: printf: invalid format`

	report := requireSingleFailure(t, analyze(log))

	if report.Category != "vet_failure" {
		t.Errorf("Category = %q, want vet_failure", report.Category)
	}

	if !strings.Contains(report.Evidence, "invalid format") {
		t.Errorf("Evidence lacks vet diagnostic: %q", report.Evidence)
	}
}

func TestAnalyzeCompileFailure(t *testing.T) {
	log := `# example.com/app
./main.go:12:2: undefined: missingFunction`

	report := requireSingleFailure(t, analyze(log))

	if report.Category != "compile_failure" {
		t.Fatalf("Category = %q, want compile_failure",
			report.Category)
	}

	if report.Confidence != "high" {
		t.Errorf("Confidence = %q, want high", report.Confidence)
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

func TestAnalyzeReturnsAllFailures(t *testing.T) {
	log := `=== FORMATTING ===
main.go

=== GO VET ===
main.go:10:2: printf: invalid format

=== TESTS ===
main_test.go:10: expected "A", got "B"

=== BUILD ===
./main.go:12:2: undefined: missingFunction`

	investigation := analyze(log)

	expected := []string{
		"format_failure",
		"vet_failure",
		"test_assertion_mismatch",
		"compile_failure",
	}

	if len(investigation.Failures) != len(expected) {
		t.Fatalf("got %d failures, want %d",
			len(investigation.Failures), len(expected))
	}

	for i, category := range expected {
		if investigation.Failures[i].Category != category {
			t.Errorf("Failures[%d].Category = %q, want %q",
				i, investigation.Failures[i].Category, category)
		}
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
			report := requireSingleFailure(t, analyze(tc.log))

			if report.Category != "unknown_failure" {
				t.Errorf("Category = %q, want unknown_failure",
					report.Category)
			}

			if report.Confidence != "low" {
				t.Errorf("Confidence = %q, want low", report.Confidence)
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

func TestAnalyzeDoesNotMisclassifyGenericError(t *testing.T) {
	report := requireSingleFailure(t, analyze("ERROR: connection timeout"))

	if report.Category != "unknown_failure" {
		t.Errorf("Category = %q, want unknown_failure",
			report.Category)
	}
}
