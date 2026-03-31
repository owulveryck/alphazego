// Package board defines the interfaces and types for board games compatible
// with the MCTS (Monte Carlo Tree Search) engine.
//
// Any game can be plugged into the MCTS engine by implementing the [State] interface.
// The package also provides common type aliases and constants for two-player games.
package board

// Agent represents a player in the game.
type Agent = uint8

// Move represents a position or action on the board.
type Move = uint8

// Result represents the outcome of evaluating a game state.
type Result = uint8

// ID is a unique identifier for a board position.
type ID = []byte

const (
	// Player1 is the first player (typically X in tic-tac-toe).
	Player1 Agent = 1
	// Player2 is the second player (typically O in tic-tac-toe).
	Player2 Agent = 2
	// EmptyPlace represents an unoccupied cell on the board.
	EmptyPlace Move = 0
	// GameOn indicates the game is still in progress.
	GameOn Result = 0
	// Player1Wins indicates that Player1 has won the game.
	Player1Wins Result = Player1
	// Player2Wins indicates that Player2 has won the game.
	Player2Wins Result = Player2
	// Draw indicates neither player has won and the board is full.
	Draw Result = 3
	// Stalemat indicates a stalemate: a player cannot make a legal move,
	// but no win condition is met. Used in some games but not in tic-tac-toe.
	Stalemat Result = 4
)

// State represents a game state. Any game that implements this interface
// can be used with the MCTS engine.
//
// A State must be immutable from the MCTS perspective: methods like
// [State.PossibleMoves] must return new State values without modifying
// the receiver.
type State interface {
	// CurrentPlayer returns the player whose turn it is to play.
	CurrentPlayer() Agent
	// Evaluate returns the current result of the game: [GameOn] if the game
	// is still in progress, or the outcome ([Player1Wins], [Player2Wins], [Draw]).
	Evaluate() Result
	// PossibleMoves returns all possible next states reachable from this state
	// by making one move. Each returned State has the opponent as CurrentPlayer.
	PossibleMoves() []State
	// BoardID returns a unique byte identifier for this board position.
	// Two states with the same board configuration and current player
	// must return identical BoardIDs.
	BoardID() ID
}

// Playable is implemented by game states that can extract the move
// that was played between two consecutive states.
type Playable interface {
	// GetMoveFromState compares the receiver with another State and returns
	// the Move (board position) that differs between the two.
	GetMoveFromState(State) Move
}
