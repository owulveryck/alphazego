package tictactoe

import (
	"testing"

	"github.com/owulveryck/alphazego/decision"
)

// FuzzPlay vérifie que Play ne panique jamais quelle que soit l'entrée.
func FuzzPlay(f *testing.F) {
	// Corpus initial : positions valides et invalides.
	for i := 0; i < 12; i++ {
		f.Add(uint8(i))
	}
	f.Add(uint8(255))

	f.Fuzz(func(t *testing.T, pos uint8) {
		ttt := NewTicTacToe()
		// Jouer la position ; une erreur est attendue si hors limites.
		_ = ttt.Play(pos)
	})
}

// FuzzSequence vérifie qu'une séquence de coups aléatoires ne panique jamais
// et respecte les invariants du morpion.
func FuzzSequence(f *testing.F) {
	// Corpus : séquences de 9 octets (positions 0-8).
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{4, 0, 8, 2, 6, 1, 3, 5, 7})
	f.Add([]byte{0, 0, 0, 0, 0}) // coups répétés
	f.Add([]byte{255, 100, 9})   // hors limites

	f.Fuzz(func(t *testing.T, moves []byte) {
		ttt := NewTicTacToe()
		for _, m := range moves {
			if ttt.Evaluate() != decision.Undecided {
				// La partie est terminée : PossibleMoves doit être vide.
				if pm := ttt.PossibleMoves(); len(pm) != 0 {
					t.Fatalf("terminal state has %d possible moves", len(pm))
				}
				break
			}
			_ = ttt.Play(m)
		}
		// Vérifier que l'état final est cohérent.
		result := ttt.Evaluate()
		if result != decision.Undecided && result != decision.Stalemate &&
			result != Cross && result != Circle {
			t.Fatalf("unexpected Evaluate result: %d", result)
		}
		// ID ne doit pas paniquer.
		_ = ttt.ID()
	})
}
