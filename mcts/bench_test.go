package mcts

import (
	"testing"

	"github.com/owulveryck/alphazego/decision/board/samples/tictactoe"
)

// BenchmarkRunMCTS_100 mesure le coût d'une recherche MCTS complète (100 itérations).
// Lancer avec : go test -bench=BenchmarkRunMCTS_100 -benchmem ./mcts/
func BenchmarkRunMCTS_100(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := NewMCTS()
		m.RunMCTS(tictactoe.NewTicTacToe(), 100)
	}
}

// BenchmarkRunMCTS_1000 mesure le scaling à 1000 itérations.
// Profiling CPU/mémoire :
//
//	go test -bench=BenchmarkRunMCTS_1000 -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./mcts/
//	go tool pprof cpu.prof
func BenchmarkRunMCTS_1000(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := NewMCTS()
		m.RunMCTS(tictactoe.NewTicTacToe(), 1000)
	}
}

// BenchmarkRunMCTS_10000 mesure le scaling à 10000 itérations.
func BenchmarkRunMCTS_10000(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := NewMCTS()
		m.RunMCTS(tictactoe.NewTicTacToe(), 10000)
	}
}

// BenchmarkSimulate mesure le coût d'un rollout aléatoire depuis un état initial.
func BenchmarkSimulate(b *testing.B) {
	m := NewMCTS()
	node := &mctsNode{
		state: tictactoe.NewTicTacToe(),
		mcts:  m,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		node.simulate()
	}
}

// BenchmarkExpand mesure le coût de l'expansion d'un nœud (allocation map + ID).
func BenchmarkExpand(b *testing.B) {
	m := NewMCTS()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		node := &mctsNode{
			state:    tictactoe.NewTicTacToe(),
			children: []*mctsNode{},
			mcts:     m,
		}
		node.expand()
	}
}

// BenchmarkExpandFull mesure le coût d'expansion complète (9 enfants pour le morpion).
func BenchmarkExpandFull(b *testing.B) {
	m := NewMCTS()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		node := &mctsNode{
			state:    tictactoe.NewTicTacToe(),
			children: []*mctsNode{},
			mcts:     m,
		}
		for node.expand() != nil {
		}
	}
}
