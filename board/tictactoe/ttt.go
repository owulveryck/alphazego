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
	"github.com/owulveryck/alphazego/board"
)

// BoardSize is the number of cells on a tic-tac-toe board (3x3 = 9).
const (
	BoardSize = 9
)

// TicTacToe represents the state of a tic-tac-toe game.
// It implements [board.State] and [board.Playable].
type TicTacToe struct {
	board      []uint8
	PlayerTurn uint8
}

// ID returns a unique identifier for this board state.
// The ID is the board cells concatenated with the current player byte,
// producing a 10-byte slice.
func (tictactoe *TicTacToe) ID() []byte {
	return append(tictactoe.board, tictactoe.PlayerTurn)
}

// GetMoveFromState compares the current board with another state and returns
// the position (0-8) where the boards differ. This identifies which move was
// played to transition from the current state to s.
func (tictactoe *TicTacToe) GetMoveFromState(s board.State) board.Move {
	next := s.(*TicTacToe)
	for i := 0; i < len(next.board); i++ {
		if next.board[i] != tictactoe.board[i] {
			return uint8(i)
		}
	}
	return 0
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
func (t *TicTacToe) Play(p board.Move) {
	t.board[p] = t.PlayerTurn
	t.PlayerTurn = 3 - t.PlayerTurn
}

// CurrentPlayer returns the player whose turn it is to play.
func (t *TicTacToe) CurrentPlayer() board.Agent {
	return t.PlayerTurn
}

// PreviousPlayer retourne l'agent qui a joue le dernier coup.
// Au morpion, c'est l'adversaire du joueur courant (alternance stricte a deux joueurs).
// Pour l'etat initial, retourne Player2 (le "dernier" dans l'ordre de jeu).
func (t *TicTacToe) PreviousPlayer() board.Agent {
	return 3 - t.PlayerTurn
}

// Evaluate checks the board for a winner or draw.
// It returns [board.GameOn] if the game is still in progress,
// [board.Player1Wins] or [board.Player2Wins] if a player has three in a row,
// or [board.Draw] if all cells are filled with no winner.
func (t *TicTacToe) Evaluate() board.Result {
	// Define all winning positions: rows, columns, and diagonals
	// Check for a win
	for _, position := range winningPositions {
		if t.board[position[0]] != 0 &&
			t.board[position[0]] == t.board[position[1]] &&
			t.board[position[1]] == t.board[position[2]] {
			// Return the winner (1 for X, 2 for O)
			return t.board[position[0]]
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
		return board.Draw // Game is a draw
	}

	// Game can continue
	return board.GameOn
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
			game[i] = t.PlayerTurn
			games = append(games, &TicTacToe{
				board:      game,
				PlayerTurn: 3 - t.PlayerTurn,
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
	current := t.CurrentPlayer()
	opponent := 3 - current

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
	if current == board.Player1 {
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
