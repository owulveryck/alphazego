package tictactoe

import "github.com/owulveryck/alphazego/board"

// Define constants for the players and empty cells
const (
	gameOn    = 0
	Empty     = 0
	PlayerX   = 1
	PlayerO   = 2
	Draw      = 3
	BoardSize = 9
)

type TicTacToe struct {
	board      []uint8
	PlayerTurn uint8
}

// CurrentPlayer is the player that will play on the current board
func (tictactoe *TicTacToe) CurrentPlayer() board.Agent {
	panic("not implemented") // TODO: Implement
}

// Evaluate the state and returns gameon, or a winner
func (t *TicTacToe) Evaluate() board.Result {
	panic("not implemented") // TODO: Implement
}

func ToBoardState(t []*TicTacToe) []board.State {
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
	return ToBoardState(games)
}

func (gs *TicTacToe) IsGameOver() bool {
	return checkGameStatus(gs.board) > 0
}

func (gs *TicTacToe) MakeMove(move *TicTacToe) *TicTacToe {
	// Apply a move to the current game state and return the new state
	return &TicTacToe{
		move.board,
		move.PlayerTurn,
	}
}

func (gs *TicTacToe) GetWinner() uint8 {
	status := checkGameStatus(gs.board)
	if status == 3 {
		return 0
	}
	return status
}

var winningPositions = [][]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // Rows
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // Columns
	{0, 4, 8}, {2, 4, 6}, // Diagonals
}

// Function to check the game status of TicTacToe
func checkGameStatus(board []uint8) uint8 {
	// Define all winning positions: rows, columns, and diagonals
	// Check for a win
	for _, position := range winningPositions {
		if board[position[0]] != 0 &&
			board[position[0]] == board[position[1]] &&
			board[position[1]] == board[position[2]] {
			// Return the winner (1 for X, 2 for O)
			return board[position[0]]
		}
	}

	// Check for a draw (if there are no empty cells left)
	draw := true
	for _, cell := range board {
		if cell == 0 {
			draw = false
			break
		}
	}
	if draw {
		return Draw // Game is a draw
	}

	// Game can continue
	return gameOn
}
