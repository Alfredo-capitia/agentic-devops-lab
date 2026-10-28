package main

import "testing"

func TestLabBaseline(t *testing.T) {
	got := "CI baseline ready"
	want := "CI baseline ready"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
