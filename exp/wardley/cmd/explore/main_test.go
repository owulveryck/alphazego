package main

import (
	"testing"
)

func TestParseBatchScoresValid(t *testing.T) {
	scores, err := parseBatchScores(`[0.7, 0.3, 0.8]`, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(scores) != 3 {
		t.Fatalf("got %d scores, want 3", len(scores))
	}
	want := []float64{0.7, 0.3, 0.8}
	for i, s := range scores {
		if diff := s - want[i]; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("score[%d] = %f, want %f", i, s, want[i])
		}
	}
}

func TestParseBatchScoresWithSurroundingText(t *testing.T) {
	text := "Voici les scores :\n[0.5, 0.9, 0.1]\nBonne continuation."
	scores, err := parseBatchScores(text, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(scores) != 3 {
		t.Fatalf("got %d scores, want 3", len(scores))
	}
}

func TestParseBatchScoresClamp(t *testing.T) {
	scores, err := parseBatchScores(`[-0.5, 1.5, 0.5]`, 3)
	if err != nil {
		t.Fatal(err)
	}
	if scores[0] != 0.0 {
		t.Errorf("score[0] = %f, want 0.0 (clamped)", scores[0])
	}
	if scores[1] != 1.0 {
		t.Errorf("score[1] = %f, want 1.0 (clamped)", scores[1])
	}
}

func TestParseBatchScoresCountMismatch(t *testing.T) {
	_, err := parseBatchScores(`[0.5, 0.9]`, 3)
	if err == nil {
		t.Error("expected error for count mismatch")
	}
}

func TestParseBatchScoresNoJSON(t *testing.T) {
	_, err := parseBatchScores(`pas de tableau`, 1)
	if err == nil {
		t.Error("expected error for missing JSON array")
	}
}
