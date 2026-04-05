package mcts

import (
	"testing"

	"github.com/owulveryck/alphazego/decision"
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
			state: tictactoe.NewTicTacToe(),
			mcts:  m,
		}
		for node.expand() != nil {
		}
	}
}

// BenchmarkSelectChildUCB mesure le coût de sélection UCB1 parmi 9 enfants.
// Représente 5.8% du CPU d'après le profiling pprof.
func BenchmarkSelectChildUCB(b *testing.B) {
	m := NewMCTS()
	node := &mctsNode{
		state: tictactoe.NewTicTacToe(),
		mcts:  m,
	}
	// Expand all children and give them some visits
	for node.expand() != nil {
	}
	for i, child := range node.children {
		child.visits = float64(10 + i)
		child.wins = float64(5 + i)
	}
	node.visits = 100
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		node.selectChildUCB()
	}
}

// wrappedState enveloppe un State pour masquer l'interface RandomMover.
type wrappedState struct {
	decision.State
}

// BenchmarkSimulate_WithRandomMover mesure le rollout avec RandomMover (TicTacToe natif).
func BenchmarkSimulate_WithRandomMover(b *testing.B) {
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

// BenchmarkSimulate_WithoutRandomMover mesure le rollout sans RandomMover (fallback PossibleMoves).
func BenchmarkSimulate_WithoutRandomMover(b *testing.B) {
	m := NewMCTS()
	node := &mctsNode{
		state: &wrappedState{tictactoe.NewTicTacToe()},
		mcts:  m,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		node.simulate()
	}
}

// BenchmarkRunMCTS_1000_WithRandomMover mesure 1000 itérations avec RandomMover.
func BenchmarkRunMCTS_1000_WithRandomMover(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := NewMCTS()
		m.RunMCTS(tictactoe.NewTicTacToe(), 1000)
	}
}

// BenchmarkRunMCTS_1000_WithoutRandomMover mesure 1000 itérations sans RandomMover.
func BenchmarkRunMCTS_1000_WithoutRandomMover(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		m := NewMCTS()
		m.RunMCTS(&wrappedState{tictactoe.NewTicTacToe()}, 1000)
	}
}

// BenchmarkRunMCTS_MidGame mesure le MCTS depuis un état en milieu de partie (5 coups joués).
func BenchmarkRunMCTS_MidGame(b *testing.B) {
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0) // X
	ttt.Play(3) // O
	ttt.Play(1) // X
	ttt.Play(4) // O
	ttt.Play(6) // X
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		m := NewMCTS()
		m.RunMCTS(ttt, 1000)
	}
}
