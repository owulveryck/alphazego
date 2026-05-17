package main

import (
	"testing"
)

func TestParseCandidatesValid(t *testing.T) {
	text := `[{"description":"Evolve DB","wtg2":"App : III.5\nDB : III.5\n","confidence":0.8}]`
	candidates, err := parseCandidates(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1", len(candidates))
	}
	if candidates[0].Description != "Evolve DB" {
		t.Errorf("description = %q, want %q", candidates[0].Description, "Evolve DB")
	}
	if candidates[0].Confidence != 0.8 {
		t.Errorf("confidence = %f, want 0.8", candidates[0].Confidence)
	}
}

func TestParseCandidatesWithSurroundingText(t *testing.T) {
	text := `Voici mes propositions :
[{"description":"A","wtg2":"wtg2 content","confidence":0.5},{"description":"B","wtg2":"other","confidence":0.7}]
Bonne continuation.`
	candidates, err := parseCandidates(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("got %d candidates, want 2", len(candidates))
	}
}

func TestParseCandidatesClamp(t *testing.T) {
	text := `[{"description":"A","wtg2":"content","confidence":1.5}]`
	candidates, err := parseCandidates(text)
	if err != nil {
		t.Fatal(err)
	}
	if candidates[0].Confidence != 1.0 {
		t.Errorf("confidence = %f, want 1.0 (clamped)", candidates[0].Confidence)
	}
}

func TestParseCandidatesSkipsEmptyWTG2(t *testing.T) {
	text := `[{"description":"A","wtg2":"","confidence":0.5},{"description":"B","wtg2":"valid","confidence":0.7}]`
	candidates, err := parseCandidates(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1 (empty WTG2 filtered)", len(candidates))
	}
}

func TestParseCandidatesNoJSON(t *testing.T) {
	_, err := parseCandidates(`pas de tableau`)
	if err == nil {
		t.Error("expected error for missing JSON array")
	}
}

func TestParseCandidatesAllEmpty(t *testing.T) {
	text := `[{"description":"A","wtg2":"","confidence":0.5}]`
	_, err := parseCandidates(text)
	if err == nil {
		t.Error("expected error when all candidates have empty WTG2")
	}
}
