package gamestate

// Define constants for the players and empty cells
const (
	gameOn    = 0
	Empty     = 0
	PlayerX   = 1
	PlayerO   = 2
	Draw      = 3
	BoardSize = 9
)

type GameState struct {
	board      []uint8
	playerTurn uint8
}

// Placeholder for GameState methods
func (gs *GameState) PossibleMoves() []GameState {
	games := make([]GameState, 0)
	for i := 0; i < BoardSize; i++ {
		if gs.board[i] == 0 {
			game := make([]uint8, BoardSize)
			copy(game, gs.board)
			game[i] = gs.playerTurn
			games = append(games, GameState{
				board:      game,
				playerTurn: 3 - gs.playerTurn,
			})
		}
	}
	// Return a slice of possible next states
	return games
}

func (gs *GameState) IsGameOver() bool {
	return checkGameStatus(gs.board) > 0
}

func (gs *GameState) MakeMove(move GameState) *GameState {
	// Apply a move to the current game state and return the new state
	return &GameState{
		move.board,
		move.playerTurn,
	}
}

func (gs *GameState) GetWinner() uint8 {
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
