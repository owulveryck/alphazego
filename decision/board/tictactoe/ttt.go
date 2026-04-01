// Package tictactoe implements a tic-tac-toe game compatible with the
// [decision.State] interface, allowing it to be used with the MCTS engine.
//
// The board is represented as a flat slice of 9 cells (positions 0-8):
//
//	0 | 1 | 2
//	──┼───┼──
//	3 | 4 | 5
//	──┼───┼──
//	6 | 7 | 8
//
// Each cell contains 0 (empty), 1 (Cross / X), or 2 (Circle / O).
package tictactoe

import (
	"fmt"

	"github.com/owulveryck/alphazego/decision"
)

// BoardSize is the number of cells on a tic-tac-toe board (3x3 = 9).
const BoardSize = 9

// Cross et Circle identifient les deux acteurs du morpion.
const (
	Cross  decision.ActorID = 1 // X, premier acteur
	Circle decision.ActorID = 2 // O, second acteur
)

// TicTacToe represents the state of a tic-tac-toe game.
// It implements [decision.State] and [board.ActionRecorder].
type TicTacToe struct {
	board      [BoardSize]uint8
	actorTurn  decision.ActorID
	lastAction int
}

// ID returns a unique identifier for this board state.
// The ID is the board cells concatenated with the current actor byte,
// producing a 10-character string.
func (tictactoe *TicTacToe) ID() string {
	var id [BoardSize + 1]byte
	copy(id[:], tictactoe.board[:])
	id[BoardSize] = byte(tictactoe.actorTurn)
	return string(id[:])
}

// LastAction retourne la position (0-8) du dernier coup joué.
// Pour l'état initial, retourne 0 (non significatif).
// Implémente [board.ActionRecorder].
func (tictactoe *TicTacToe) LastAction() int {
	return tictactoe.lastAction
}

// NewTicTacToe creates a new tic-tac-toe game with an empty board.
// Actor1 goes first.
func NewTicTacToe() *TicTacToe {
	return &TicTacToe{
		actorTurn: Cross,
	}
}

// Play places the current actor's mark at position p (0-8)
// and switches the turn to the other actor.
// It returns an error if the position is out of bounds, already occupied,
// or the game is already over.
func (t *TicTacToe) Play(p uint8) error {
	if p >= BoardSize {
		return fmt.Errorf("position %d hors limites (0-%d)", p, BoardSize-1)
	}
	if t.board[p] != 0 {
		return fmt.Errorf("position %d déjà occupée", p)
	}
	if t.Evaluate() != decision.Undecided {
		return fmt.Errorf("la partie est terminée")
	}
	t.board[p] = uint8(t.actorTurn)
	t.lastAction = int(p)
	t.actorTurn = 3 - t.actorTurn
	return nil
}

// CurrentActor returns the actor whose turn it is to play.
func (t *TicTacToe) CurrentActor() decision.ActorID {
	return t.actorTurn
}

// PreviousActor retourne l'acteur qui a joué le dernier coup.
// Au morpion, c'est l'adversaire de l'acteur courant (alternance stricte à deux acteurs).
// Pour l'état initial, retourne Actor2 (le "dernier" dans l'ordre de jeu).
func (t *TicTacToe) PreviousActor() decision.ActorID {
	return 3 - t.actorTurn
}

// Evaluate checks the board for a winner or draw.
// It returns [decision.Undecided] if the game is still in progress,
// the winning [decision.ActorID] if an actor has three in a row,
// or [decision.Stalemate] if all cells are filled with no winner.
func (t *TicTacToe) Evaluate() decision.ActorID {
	// Check all winning positions: rows, columns, and diagonals
	for _, position := range winningPositions {
		if t.board[position[0]] != 0 &&
			t.board[position[0]] == t.board[position[1]] &&
			t.board[position[1]] == t.board[position[2]] {
			// Return the winner's ActorID
			return decision.ActorID(t.board[position[0]])
		}
	}

	// Check for a draw (if there are no empty cells left)
	draw := true
	for _, cell := range t.board {
		if cell == 0 {
			draw = false
			break
		}
	}
	if draw {
		return decision.Stalemate
	}

	// Game can continue
	return decision.Undecided
}

func toDecisionState(t []*TicTacToe) []decision.State {
	output := make([]decision.State, len(t))
	for i := range t {
		output[i] = t[i]
	}
	return output
}

// PossibleMoves returns a slice of all reachable game states from the current
// position. Each returned state has one additional move played (at an empty cell)
// and the turn switched to the other actor.
func (t *TicTacToe) PossibleMoves() []decision.State {
	games := make([]*TicTacToe, 0)
	for i := 0; i < BoardSize; i++ {
		if t.board[i] == 0 {
			game := t.board // copie par valeur (tableau fixe)
			game[i] = uint8(t.actorTurn)
			games = append(games, &TicTacToe{
				board:      game,
				actorTurn:  3 - t.actorTurn,
				lastAction: i,
			})
		}
	}
	// Return a slice of possible next states
	return toDecisionState(games)
}

var winningPositions = [8][3]uint8{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // Rows
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // Columns
	{0, 4, 8}, {2, 4, 6}, // Diagonals
}

// Features retourne l'état du morpion sous forme de tenseur aplati [3 * 3 * 3] = 27 float32.
//
//   - Plan 0 (indices 0-8) : positions de l'acteur courant (1.0 si occupée, 0.0 sinon)
//   - Plan 1 (indices 9-17) : positions de l'adversaire
//   - Plan 2 (indices 18-26) : indicateur de l'acteur courant (1.0 si Actor1, 0.0 si Actor2)
func (t *TicTacToe) Features() []float32 {
	features := make([]float32, 3*3*3) // [3][3][3]
	current := uint8(t.CurrentActor())
	opponent := uint8(3 - t.actorTurn)

	for i := 0; i < BoardSize; i++ {
		if t.board[i] == current {
			features[i] = 1.0 // Plan 0 : acteur courant
		}
		if t.board[i] == opponent {
			features[9+i] = 1.0 // Plan 1 : adversaire
		}
	}

	// Plan 2 : indicateur de l'acteur courant
	val := float32(0.0)
	if t.actorTurn == Cross {
		val = 1.0
	}
	for i := 18; i < 27; i++ {
		features[i] = val
	}

	return features
}

// FeatureShape retourne les dimensions du tenseur : 3 canaux, plateau 3x3.
func (t *TicTacToe) FeatureShape() [3]int {
	return [3]int{3, 3, 3}
}

// ActionSize retourne le nombre total d'actions possibles au morpion (9 cases).
func (t *TicTacToe) ActionSize() int {
	return BoardSize
}
