package board

// Agent is a player
type Agent = uint8

// Move is a possible value on a board
type Move = uint8

// Result is the result of the evaluation of a state
type Result = uint8

const (
	// Player1 ...
	Player1 Agent = 1
	// Player2 ...
	Player2 Agent = 2
	// EmptyPlace is the value on a board that has not been played yet
	EmptyPlace Move = 0
	// GameOn when we did not reach an end state
	GameOn Result = 0
	// Player1Wins
	Player1Wins Result = Player1
	// Player2Wins
	Player2Wins Result = Player2
	// Draw is when neither player wins, often due to the board being filled without meeting any player's win condition
	Draw Result = 3
	// Stalemat is specific type of draw in some games, where one player cannot make a legal move, but the game's win conditions are not met.
	Stalemat Result = 4
)

type State interface {
	// CurrentPlayer is the player that will play on the current board
	CurrentPlayer() Agent
	// Evaluate the state and returns gameon, or a winner
	Evaluate() Result
	// PossibleMoves ...
	PossibleMoves() []State
}

type Playable interface {
	GetMoveFromState(State) Move
}
