// Package tictactoe implements a tic-tac-toe game compatible with the
// [board.State] interface, allowing it to be used with the MCTS engine.
//
// The board is represented as a flat slice of 9 cells (positions 0-8):
//
//	0 | 1 | 2
//	──┼───┼──
//	3 | 4 | 5
//	──┼───┼──
//	6 | 7 | 8
//
// Each cell contains 0 (empty), 1 ([board.Player1] / X), or 2 ([board.Player2] / O).
package tictactoe

import (
	"fmt"

	"github.com/owulveryck/alphazego/board"
)

// BoardSize is the number of cells on a tic-tac-toe board (3x3 = 9).
const (
	BoardSize = 9
)

// TicTacToe represents the state of a tic-tac-toe game.
// It implements [board.State].
type TicTacToe struct {
	board      []uint8
	PlayerTurn board.PlayerID
	lastMove   uint8
}

// ID returns a unique identifier for this board state.
// The ID is the board cells concatenated with the current player byte,
// producing a 10-character string.
func (tictactoe *TicTacToe) ID() string {
	id := make([]byte, BoardSize+1)
	copy(id, tictactoe.board)
	id[BoardSize] = byte(tictactoe.PlayerTurn)
	return string(id)
}

// LastMove retourne la position (0-8) du dernier coup joue.
// Pour l'etat initial, retourne 0 (non significatif).
func (tictactoe *TicTacToe) LastMove() uint8 {
	return tictactoe.lastMove
}

// NewTicTacToe creates a new tic-tac-toe game with an empty board.
// Player1 goes first.
func NewTicTacToe() *TicTacToe {
	return &TicTacToe{
		board:      make([]uint8, BoardSize),
		PlayerTurn: board.Player1,
	}
}

// Play places the current player's mark at position p (0-8)
// and switches the turn to the other player.
// It returns an error if the position is out of bounds, already occupied,
// or the game is already over.
func (t *TicTacToe) Play(p uint8) error {
	if p >= BoardSize {
		return fmt.Errorf("position %d hors limites (0-%d)", p, BoardSize-1)
	}
	if t.board[p] != 0 {
		return fmt.Errorf("position %d deja occupee", p)
	}
	if t.Evaluate() != board.NoPlayer {
		return fmt.Errorf("la partie est terminee")
	}
	t.board[p] = uint8(t.PlayerTurn)
	t.lastMove = p
	t.PlayerTurn = 3 - t.PlayerTurn
	return nil
}

// CurrentPlayer returns the player whose turn it is to play.
func (t *TicTacToe) CurrentPlayer() board.PlayerID {
	return t.PlayerTurn
}

// PreviousPlayer retourne le joueur qui a joue le dernier coup.
// Au morpion, c'est l'adversaire du joueur courant (alternance stricte a deux joueurs).
// Pour l'etat initial, retourne Player2 (le "dernier" dans l'ordre de jeu).
func (t *TicTacToe) PreviousPlayer() board.PlayerID {
	return 3 - t.PlayerTurn
}

// Evaluate checks the board for a winner or draw.
// It returns [board.NoPlayer] if the game is still in progress,
// the winning [board.PlayerID] if a player has three in a row,
// or [board.DrawResult] if all cells are filled with no winner.
func (t *TicTacToe) Evaluate() board.PlayerID {
	// Check all winning positions: rows, columns, and diagonals
	for _, position := range winningPositions {
		if t.board[position[0]] != 0 &&
			t.board[position[0]] == t.board[position[1]] &&
			t.board[position[1]] == t.board[position[2]] {
			// Return the winner's PlayerID
			return board.PlayerID(t.board[position[0]])
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
		return board.DrawResult
	}

	// Game can continue
	return board.NoPlayer
}

func toBoardState(t []*TicTacToe) []board.State {
	output := make([]board.State, len(t))
	for i := range t {
		output[i] = t[i]
	}
	return output
}

// PossibleMoves returns a slice of all reachable game states from the current
// position. Each returned state has one additional move played (at an empty cell)
// and the turn switched to the other player.
func (t *TicTacToe) PossibleMoves() []board.State {
	games := make([]*TicTacToe, 0)
	for i := 0; i < BoardSize; i++ {
		if t.board[i] == 0 {
			game := make([]uint8, BoardSize)
			copy(game, t.board)
			game[i] = uint8(t.PlayerTurn)
			games = append(games, &TicTacToe{
				board:      game,
				PlayerTurn: 3 - t.PlayerTurn,
				lastMove:   uint8(i),
			})
		}
	}
	// Return a slice of possible next states
	return toBoardState(games)
}

var winningPositions = [][]uint8{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // Rows
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // Columns
	{0, 4, 8}, {2, 4, 6}, // Diagonals
}

// Features retourne l'etat du morpion sous forme de tenseur aplati [3 * 3 * 3] = 27 float32.
//
//   - Plan 0 (indices 0-8) : positions du joueur courant (1.0 si occupee, 0.0 sinon)
//   - Plan 1 (indices 9-17) : positions de l'adversaire
//   - Plan 2 (indices 18-26) : indicateur du joueur courant (1.0 si Player1, 0.0 si Player2)
func (t *TicTacToe) Features() []float32 {
	features := make([]float32, 3*3*3) // [3][3][3]
	current := uint8(t.CurrentPlayer())
	opponent := uint8(3 - t.PlayerTurn)

	for i := 0; i < BoardSize; i++ {
		if t.board[i] == current {
			features[i] = 1.0 // Plan 0 : joueur courant
		}
		if t.board[i] == opponent {
			features[9+i] = 1.0 // Plan 1 : adversaire
		}
	}

	// Plan 2 : indicateur du joueur courant
	val := float32(0.0)
	if t.PlayerTurn == board.Player1 {
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
