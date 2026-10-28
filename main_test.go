package main

import "testing"

func TestLabBaseline(t *testing.T) {
	got := "CI baseline ready"
	want := "CI should fail"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
