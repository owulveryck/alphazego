package taquin

import (
	"math/rand"
	"testing"

	"github.com/owulveryck/alphazego/decision"
)

// FuzzShuffle vérifie que Shuffle ne panique jamais quelle que soit la graine
// et que l'état reste cohérent après le mélange.
func FuzzShuffle(f *testing.F) {
	f.Add(int64(0), 5)
	f.Add(int64(42), 10)
	f.Add(int64(-1), 100)
	f.Add(int64(999), 0)

	f.Fuzz(func(t *testing.T, seed int64, moves int) {
		if moves < 0 {
			moves = -moves
		}
		if moves > 1000 {
			moves = 1000
		}
		taq := NewTaquin(3, 3, 50)
		rng := rand.New(rand.NewSource(seed))
		taq.Shuffle(moves, rng)

		// Vérifier les invariants après shuffle.
		if taq.Evaluate() == decision.Undecided || moves == 0 {
			// steps doit être 0 après shuffle.
			if taq.steps != 0 {
				t.Fatalf("steps should be 0 after Shuffle, got %d", taq.steps)
			}
		}
		// ID ne doit pas paniquer.
		_ = taq.ID()
		// PossibleMoves ne doit pas paniquer.
		_ = taq.PossibleMoves()
	})
}

// FuzzPlaySequence vérifie qu'une séquence de directions ne panique jamais.
func FuzzPlaySequence(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3})
	f.Add([]byte{0, 0, 0, 0})
	f.Add([]byte{255, 100, 4, 5})

	f.Fuzz(func(t *testing.T, dirs []byte) {
		if len(dirs) > 100 {
			dirs = dirs[:100]
		}
		taq := NewTaquin(3, 3, 50)
		rng := rand.New(rand.NewSource(42))
		taq.Shuffle(10, rng)

		for _, d := range dirs {
			if taq.Evaluate() != decision.Undecided {
				break
			}
			moves := taq.PossibleMoves()
			if len(moves) == 0 {
				break
			}
			// Choisir un coup valide parmi les possibles.
			idx := int(d) % len(moves)
			taq = moves[idx].(*Taquin)
		}
		// L'état final doit être cohérent.
		result := taq.Evaluate()
		if result != decision.Undecided && result != decision.Stalemate && result != Player {
			t.Fatalf("unexpected Evaluate result: %d", result)
		}
	})
}
