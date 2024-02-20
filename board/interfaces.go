package board

// Agent is a player, usually is 1 and 2
type Agent = uint8

// Move is a possible value on a board
type Move = uint8

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
	Player1Wins Result = 1
	// Player2Wins
	Player2Wins Result = 2
	// Draw
	Draw Result = 3
)

type Arena interface {
	GetBoard() []Move
	GetPlayerTurn() Agent
}

// State is a special representation of an Arena
type State = Arena

type Game interface {
	IsGameOver() bool
	GetWinner() Agent
	MakeMove(Arena) Arena
	PossibleMoves() []Arena
}
