package main

import (
	"encoding/json"
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

func TestAnalyzeReturnsAllAssertionMismatches(t *testing.T) {
	log := `first_test.go:10: expected "one", got "two"
second_test.go:20: expected "three", got "four"
third_test.go:30: expected "five", got "six"`

	investigation := analyze(log)

	if len(investigation.Failures) != 3 {
		t.Fatalf("got %d failures, want 3",
			len(investigation.Failures))
	}

	expectedFiles := []string{
		"first_test.go:10",
		"second_test.go:20",
		"third_test.go:30",
	}

	for i, expected := range expectedFiles {
		if !strings.Contains(investigation.Failures[i].Evidence, expected) {
			t.Errorf("Failures[%d] evidence = %q, want location %q",
				i, investigation.Failures[i].Evidence, expected)
		}
	}

	for i, failure := range investigation.Failures {
		if failure.Category != "test_assertion_mismatch" {
			t.Errorf("Failures[%d].Category = %q, want test_assertion_mismatch",
				i, failure.Category)
		}
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

func TestAnalyzeVetFailureIncludesStructuredLocation(t *testing.T) {
	log := `=== GO VET ===
internal/server.go:25:4: unreachable code`

	report := requireSingleFailure(t, analyze(log))

	if report.Location == nil {
		t.Fatal("Location must not be nil")
	}

	if report.Location.File != "internal/server.go" {
		t.Errorf("Location.File = %q, want internal/server.go",
			report.Location.File)
	}

	if report.Location.Line != 25 {
		t.Errorf("Location.Line = %d, want 25",
			report.Location.Line)
	}

	if report.Location.Column != 4 {
		t.Errorf("Location.Column = %d, want 4",
			report.Location.Column)
	}

	if report.Message != "unreachable code" {
		t.Errorf("Message = %q, want unreachable code",
			report.Message)
	}
}

func TestAnalyzeDoesNotMisclassifyGenericError(t *testing.T) {
	report := requireSingleFailure(t, analyze("ERROR: connection timeout"))

	if report.Category != "unknown_failure" {
		t.Errorf("Category = %q, want unknown_failure",
			report.Category)
	}
}

func TestAnalyzeReturnsAllVetDiagnostics(t *testing.T) {
	log := `=== GO VET ===
main.go:10:2: printf: invalid format
internal/server.go:25:4: unreachable code`

	investigation := analyze(log)

	if len(investigation.Failures) != 2 {
		t.Fatalf("got %d failures, want 2",
			len(investigation.Failures))
	}

	expectedEvidence := []string{
		"main.go:10:2",
		"internal/server.go:25:4",
	}

	for i, expected := range expectedEvidence {
		failure := investigation.Failures[i]

		if failure.Category != "vet_failure" {
			t.Errorf("Failures[%d].Category = %q, want vet_failure",
				i, failure.Category)
		}

		if !strings.Contains(failure.Evidence, expected) {
			t.Errorf("Failures[%d].Evidence = %q, want %q",
				i, failure.Evidence, expected)
		}
	}
}

func TestAnalyzeReturnsAllFormattingFailures(t *testing.T) {
	log := `=== FORMATTING ===
main.go
internal/server.go
cmd/worker.go`

	investigation := analyze(log)

	if len(investigation.Failures) != 3 {
		t.Fatalf("got %d failures, want 3",
			len(investigation.Failures))
	}

	expectedFiles := []string{
		"main.go",
		"internal/server.go",
		"cmd/worker.go",
	}

	for i, expected := range expectedFiles {
		failure := investigation.Failures[i]

		if failure.Category != "format_failure" {
			t.Errorf("Failures[%d].Category = %q, want format_failure",
				i, failure.Category)
		}

		if !strings.Contains(failure.Evidence, expected) {
			t.Errorf("Failures[%d].Evidence = %q, want %q",
				i, failure.Evidence, expected)
		}
	}
}
func TestAnalyzeReturnsAllCompileFailures(t *testing.T) {
	log := `=== BUILD ===
# example.com/app
./main.go:12:2: undefined: missingFunction
./internal/server.go:25:4: undefined: missingHandler
./cmd/worker.go:40:3: undefined: missingWorker`

	investigation := analyze(log)

	if len(investigation.Failures) != 3 {
		t.Fatalf("got %d failures, want 3",
			len(investigation.Failures))
	}

	expectedEvidence := []string{
		"./main.go:12:2",
		"./internal/server.go:25:4",
		"./cmd/worker.go:40:3",
	}

	for i, expected := range expectedEvidence {
		failure := investigation.Failures[i]

		if failure.Category != "compile_failure" {
			t.Errorf("Failures[%d].Category = %q, want compile_failure",
				i, failure.Category)
		}

		if !strings.Contains(failure.Evidence, expected) {
			t.Errorf("Failures[%d].Evidence = %q, want %q",
				i, failure.Evidence, expected)
		}

		if failure.Severity != "high" {
			t.Errorf("Failures[%d].Severity = %q, want high",
				i, failure.Severity)
		}
	}
}
func TestAnalyzeVetIgnoresPackageHeaders(t *testing.T) {
	log := `=== GO VET ===
# example.com/app
./main.go:10:2: printf: invalid format
# example.com/app/internal
./internal/server.go:25:4: unreachable code`

	investigation := analyze(log)

	if len(investigation.Failures) != 2 {
		t.Fatalf("got %d failures, want 2",
			len(investigation.Failures))
	}

	expectedEvidence := []string{
		"./main.go:10:2",
		"./internal/server.go:25:4",
	}

	for i, expected := range expectedEvidence {
		if !strings.Contains(investigation.Failures[i].Evidence, expected) {
			t.Errorf("Failures[%d].Evidence = %q, want %q",
				i, investigation.Failures[i].Evidence, expected)
		}
	}
}

func TestAnalyzeCompileFailureIncludesStructuredLocation(t *testing.T) {
	log := `=== BUILD ===
internal/database/connection.go:42:5: undefined: missingFunction`

	report := requireSingleFailure(t, analyze(log))

	if report.Location == nil {
		t.Fatal("Location must not be nil")
	}

	if report.Location.File != "internal/database/connection.go" {
		t.Errorf("Location.File = %q, want internal/database/connection.go",
			report.Location.File)
	}

	if report.Location.Line != 42 {
		t.Errorf("Location.Line = %d, want 42", report.Location.Line)
	}

	if report.Location.Column != 5 {
		t.Errorf("Location.Column = %d, want 5", report.Location.Column)
	}

	if report.Message != "undefined: missingFunction" {
		t.Errorf("Message = %q, want undefined: missingFunction",
			report.Message)
	}
}
func TestAnalyzeTestFailureIncludesStructuredLocation(t *testing.T) {
	log := `=== TESTS ===
cmd/api_test.go:42: expected "success", got "failed"`

	report := requireSingleFailure(t, analyze(log))

	if report.Location == nil {
		t.Fatal("Location must not be nil")
	}

	if report.Location.File != "cmd/api_test.go" {
		t.Errorf("Location.File = %q, want cmd/api_test.go",
			report.Location.File)
	}

	if report.Location.Line != 42 {
		t.Errorf("Location.Line = %d, want 42", report.Location.Line)
	}

	if report.Message != `expected "success", got "failed"` {
		t.Errorf("Message = %q, want expected \"success\", got \"failed\"",
			report.Message)
	}
}
func TestAnalyzeFormattingFailureIncludesStructuredLocation(t *testing.T) {
	log := `=== FORMATTING ===
internal/server.go`

	report := requireSingleFailure(t, analyze(log))

	if report.Location == nil {
		t.Fatal("Location must not be nil")
	}

	if report.Location.File != "internal/server.go" {
		t.Errorf("Location.File = %q, want internal/server.go",
			report.Location.File)
	}

	if report.Message != "File is not formatted with gofmt." {
		t.Errorf("Message = %q, want File is not formatted with gofmt.",
			report.Message)
	}
}
func TestAnalyzeJSONPreservesStructuredDiagnostics(t *testing.T) {
	log := `=== FORMATTING ===
internal/server.go
=== GO VET ===
internal/server.go:10:2: printf: invalid format
=== TESTS ===
cmd/api_test.go:42: expected "success", got "failed"
=== BUILD ===
./cmd/worker.go:20:3: undefined: missingWorker`

	investigation := analyze(log)

	data, err := json.Marshal(investigation)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var result struct {
		Failures []struct {
			Category string `json:"category"`
			Location *struct {
				File   string `json:"file"`
				Line   int    `json:"line"`
				Column int    `json:"column"`
			} `json:"location"`
			Message string `json:"message"`
		} `json:"failures"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(result.Failures) != 4 {
		t.Fatalf("got %d failures, want 4", len(result.Failures))
	}

	expectedCategories := []string{
		"format_failure",
		"vet_failure",
		"test_assertion_mismatch",
		"compile_failure",
	}

	for i, expected := range expectedCategories {
		if result.Failures[i].Category != expected {
			t.Errorf(
				"Failures[%d].Category = %q, want %q",
				i,
				result.Failures[i].Category,
				expected,
			)
		}

		if result.Failures[i].Location == nil {
			t.Fatalf("Failures[%d].Location is nil", i)
		}
	}

	if result.Failures[0].Location.File != "internal/server.go" {
		t.Errorf("Formatting location incorrect")
	}

	if result.Failures[1].Location.Column != 2 {
		t.Errorf("Vet column = %d, want 2",
			result.Failures[1].Location.Column)
	}

	if result.Failures[2].Location.Line != 42 {
		t.Errorf("Test line = %d, want 42",
			result.Failures[2].Location.Line)
	}

	if result.Failures[3].Location.Column != 3 {
		t.Errorf("Compile column = %d, want 3",
			result.Failures[3].Location.Column)
	}

	jsonOutput := string(data)

	if !strings.Contains(jsonOutput,
		`"location":{"file":"internal/server.go"}`) {
		t.Error("Formatting JSON should omit unavailable line and column")
	}

	if !strings.Contains(jsonOutput,
		`"message":"File is not formatted with gofmt."`) {
		t.Error("Formatting message missing from JSON")
	}
}
