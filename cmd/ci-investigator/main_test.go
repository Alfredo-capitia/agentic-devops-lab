package main

import (
	"strings"
	"testing"
)

func TestAnalyzeAssertionMismatch(t *testing.T) {
	log := `main_test.go:10: expected "CI should fail", got "CI baseline ready"`

	report := analyze(log)

	if report.Category != "test_assertion_mismatch" {
		t.Fatalf("category = %q, want %q",
			report.Category, "test_assertion_mismatch")
	}

	if report.Confidence != "high" {
		t.Errorf("confidence = %q, want %q",
			report.Confidence, "high")
	}

	if !strings.Contains(report.Evidence, "main_test.go:10") {
		t.Errorf("evidence does not contain the source location: %q",
			report.Evidence)
	}
}

func TestAnalyzeUnknownFailure(t *testing.T) {
	report := analyze("ERROR: connection timeout")

	if report.Category != "unknown_failure" {
		t.Fatalf("category = %q, want %q",
			report.Category, "unknown_failure")
	}

	if report.Confidence != "low" {
		t.Errorf("confidence = %q, want %q",
			report.Confidence, "low")
	}
}

func TestAnalyzeEmptyLog(t *testing.T) {
	report := analyze("")

	if report.Category != "unknown_failure" {
		t.Errorf("category = %q, want %q",
			report.Category, "unknown_failure")
	}
}
