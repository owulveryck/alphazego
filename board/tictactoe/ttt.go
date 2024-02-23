package tictactoe

import (
	"github.com/owulveryck/alphazego/board"
)

// Define constants for the players and empty cells
const (
	BoardSize = 9
)

type TicTacToe struct {
	board      []uint8
	PlayerTurn uint8
}

func (tictactoe *TicTacToe) GetMoveFromState(s board.State) board.Move {
	next := s.(*TicTacToe)
	for i := 0; i < len(next.board); i++ {
		if next.board[i] != tictactoe.board[i] {
			return uint8(i)
		}
	}
	return 0
}

func NewTicTacToe() *TicTacToe {
	return &TicTacToe{
		board:      make([]uint8, BoardSize),
		PlayerTurn: board.Player1,
	}
}

func (t *TicTacToe) Play(p board.Move) {
	t.board[p] = t.PlayerTurn
	t.PlayerTurn = 3 - t.PlayerTurn
}

// CurrentPlayer is the player that will play on the current board
func (t *TicTacToe) CurrentPlayer() board.Agent {
	return t.PlayerTurn
}

// Evaluate the state and returns gameon, or a winner
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

// Placeholder for GameState methods
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
