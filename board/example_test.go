package board_test

import (
	"fmt"

	"github.com/owulveryck/alphazego/board"
	"github.com/owulveryck/alphazego/mcts"
)

// ttt est une implementation minimale de [board.State] pour le morpion.
// Elle illustre les trois champs necessaires : le plateau, le joueur
// courant, et le dernier coup joue.
//
// Le plateau utilise un tableau fixe [9]uint8 (pas un slice) :
// l'affectation d'un tableau copie les donnees automatiquement,
// ce qui simplifie [board.State.PossibleMoves].
type ttt struct {
	cells [9]uint8       // 0=vide, 1=Player1, 2=Player2
	turn  board.PlayerID // joueur dont c'est le tour
	last  uint8          // coup qui a produit cet etat
}

func (t *ttt) CurrentPlayer() board.PlayerID  { return t.turn }
func (t *ttt) PreviousPlayer() board.PlayerID { return 3 - t.turn }
func (t *ttt) LastMove() uint8                { return t.last }

// ID inclut le joueur courant : meme plateau + joueur different = ID different.
func (t *ttt) ID() string {
	var id [10]byte
	copy(id[:], t.cells[:])
	id[9] = byte(t.turn)
	return string(id[:])
}

// Evaluate verifie les 8 combinaisons gagnantes, puis le match nul.
func (t *ttt) Evaluate() board.PlayerID {
	for _, l := range [][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // lignes
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // colonnes
		{0, 4, 8}, {2, 4, 6}, // diagonales
	} {
		if t.cells[l[0]] != 0 &&
			t.cells[l[0]] == t.cells[l[1]] &&
			t.cells[l[1]] == t.cells[l[2]] {
			return board.PlayerID(t.cells[l[0]])
		}
	}
	for _, c := range t.cells {
		if c == 0 {
			return board.NoPlayer
		}
	}
	return board.DrawResult
}

// PossibleMoves retourne un etat par case vide.
// Chaque enfant est une copie independante grace au tableau [9]uint8 :
// l'affectation child.cells = t.cells copie les 9 octets.
func (t *ttt) PossibleMoves() []board.State {
	if t.Evaluate() != board.NoPlayer {
		return nil // etat terminal : aucun coup
	}
	var moves []board.State
	for i, c := range t.cells {
		if c == 0 {
			child := &ttt{
				cells: t.cells,    // copie automatique (tableau, pas slice)
				turn:  3 - t.turn, // joueur suivant
				last:  uint8(i),   // coup qui a produit cet etat
			}
			child.cells[i] = uint8(t.turn) // placer le pion du joueur courant
			moves = append(moves, child)
		}
	}
	return moves
}

// Cet exemple montre une implementation complete de [board.State] pour le
// morpion (tic-tac-toe), connectee au moteur MCTS.
//
// L'implementation tient en ~50 lignes grace a deux choix :
//   - un tableau fixe [9]uint8 pour le plateau (copie automatique)
//   - le joueur suivant calcule par 3 - joueur (alternance 1↔2)
//
// Voir le type ttt dans le code source pour l'implementation complete.
func Example() {
	game := &ttt{turn: board.Player1}

	// Connecter au MCTS : une seule ligne
	m := mcts.NewMCTS()
	bestState := m.RunMCTS(game, 1000)

	// LastMove() donne le coup choisi par le MCTS
	move := bestState.LastMove()
	fmt.Println("MCTS chose a valid position:", move <= 8)
	fmt.Println("Next player after MCTS move:", bestState.CurrentPlayer())
	// Output:
	// MCTS chose a valid position: true
	// Next player after MCTS move: 2
}

// Cet exemple montre une partie complete MCTS vs MCTS.
func Example_fullGame() {
	game := &ttt{turn: board.Player1}
	m := mcts.NewMCTS()

	for game.Evaluate() == board.NoPlayer {
		bestState := m.RunMCTS(game, 500)
		move := bestState.LastMove()
		// Appliquer le coup
		game.cells[move] = uint8(game.turn)
		game.turn = 3 - game.turn
		game.last = move
	}

	result := game.Evaluate()
	fmt.Println("Game finished:", result != board.NoPlayer)
	// Output:
	// Game finished: true
}
