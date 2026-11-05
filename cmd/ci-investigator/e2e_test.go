package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestE2EInvestigator(t *testing.T) {
	// build binary
	dir := t.TempDir()
	bin := filepath.Join(dir, "ci-investigator")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = filepath.Join(".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	input, err := os.ReadFile("testdata/ci.log")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	run := exec.Command(bin)
	run.Stdin = bytes.NewReader(input)
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}

	var result InvestigationResult
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}

	if len(result.DeterministicFindings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.DeterministicFindings))
	}
	f := result.DeterministicFindings[0]
	if f.Category != "test_assertion_mismatch" {
		t.Fatalf("category = %q", f.Category)
	}
	if f.Location == nil || f.Location.File != "example_test.go" || f.Location.Line != 12 {
		t.Fatalf("unexpected location: %+v", f.Location)
	}
	if result.AIAnalysis.Status != "not_requested" {
		t.Fatalf("ai status = %q", result.AIAnalysis.Status)
	}
}
