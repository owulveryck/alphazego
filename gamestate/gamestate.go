package gamestate

type GameState struct {
	board      [3][3]int
	playerTurn int
}

// Placeholder for GameState methods
func (gs *GameState) PossibleMoves() []GameState {
	// Return a slice of possible next states
	return nil
}

func (gs *GameState) IsGameOver() bool {
	// Return true if the game is over, false otherwise
	return false
}

func (gs *GameState) MakeMove(move GameState) *GameState {
	// Apply a move to the current game state and return the new state
	return &GameState{}
}

func (gs *GameState) GetWinner() int {
	// Determine the winner of the game; return 0 for draw, 1 for Player X, 2 for Player O
	return 0
}
