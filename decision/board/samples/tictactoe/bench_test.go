package tictactoe

import "testing"

// BenchmarkID mesure le coût de génération d'un identifiant d'état.
// ID() alloue une string à chaque appel (string([]byte)).
func BenchmarkID(b *testing.B) {
	t := NewTicTacToe()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.ID()
	}
}

// BenchmarkEvaluate mesure le coût d'évaluation d'un état (détection victoire/match nul).
func BenchmarkEvaluate(b *testing.B) {
	t := NewTicTacToe()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.Evaluate()
	}
}

// BenchmarkEvaluate_MidGame mesure Evaluate en milieu de partie (5 coups joués).
func BenchmarkEvaluate_MidGame(b *testing.B) {
	t := NewTicTacToe()
	t.Play(0) // X
	t.Play(3) // O
	t.Play(1) // X
	t.Play(4) // O
	t.Play(6) // X
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.Evaluate()
	}
}

// BenchmarkPossibleMoves mesure le coût de génération des coups possibles.
// C'est le hotspot principal dans simulate() (rollout MCTS).
func BenchmarkPossibleMoves(b *testing.B) {
	t := NewTicTacToe()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.PossibleMoves()
	}
}

// BenchmarkPossibleMoves_MidGame mesure PossibleMoves avec moins de cases libres.
func BenchmarkPossibleMoves_MidGame(b *testing.B) {
	t := NewTicTacToe()
	t.Play(0)
	t.Play(3)
	t.Play(1)
	t.Play(4)
	t.Play(6)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.PossibleMoves()
	}
}
