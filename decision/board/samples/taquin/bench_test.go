package taquin

import (
	"math/rand"
	"testing"
)

// BenchmarkID mesure le coût de génération d'un identifiant d'état.
// ID() fait un make([]byte) + string() — double allocation.
func BenchmarkID(b *testing.B) {
	t := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	t.Shuffle(10, rng)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.ID()
	}
}

// BenchmarkEvaluate mesure le coût d'évaluation (isSolved scanne tout le plateau).
func BenchmarkEvaluate(b *testing.B) {
	t := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	t.Shuffle(10, rng)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.Evaluate()
	}
}

// BenchmarkEvaluate_Large mesure Evaluate sur un grand plateau (5x5).
func BenchmarkEvaluate_Large(b *testing.B) {
	t := NewTaquin(5, 5, 200)
	rng := rand.New(rand.NewSource(42))
	t.Shuffle(10, rng)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.Evaluate()
	}
}

// BenchmarkPossibleMoves mesure le coût de génération des coups possibles.
func BenchmarkPossibleMoves(b *testing.B) {
	t := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	t.Shuffle(10, rng)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.PossibleMoves()
	}
}

// BenchmarkRandomMove mesure le coût de RandomMove (1 seule allocation).
func BenchmarkRandomMove(b *testing.B) {
	t := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	t.Shuffle(10, rng)
	rngFn := func(n int) int { return n / 2 }
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		t.RandomMove(rngFn)
	}
}
