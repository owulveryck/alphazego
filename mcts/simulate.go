package mcts

import (
	"math/rand"

	"github.com/owulveryck/alphazego/decision"
)

// simulate performs a random playthrough from the current state until a terminal state is reached.
// It selects moves randomly and advances the state until it can be evaluated as a win, lose, or draw.
func (node *mctsNode) simulate() decision.ActorID {
	// Start from the current state of the node.
	currentState := node.state

	// Continue simulating random moves until the game reaches a terminal state.
	// The game is in a terminal state if it is not in the NoActor state anymore.
	for currentState.Evaluate() == decision.NoActor {
		possibleMoves := currentState.PossibleMoves() // Get all possible moves from the current state.

		// Randomly select one of the possible moves.
		// This approach simulates a playthrough with random decisions, mimicking an unpredictable game.
		currentState = possibleMoves[rand.Intn(len(possibleMoves))]
	}

	// After reaching a terminal state, evaluate and return the outcome of the game.
	// The outcome is determined based on the rules defined in the State's Evaluate method.
	return currentState.Evaluate()
}
