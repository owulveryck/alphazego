package mcts

import (
	"math/rand"

	"github.com/owulveryck/alphazego/board"
)

func (node *MCTSNode) Simulate() board.Result {
	// Simulate a random playthrough from this node to a terminal state
	currentState := node.state
	for currentState.Evaluate() == board.GameOn {
		possibleMoves := currentState.PossibleMoves()
		currentState = possibleMoves[rand.Intn(len(possibleMoves))] // Randomly select a move

	}
	return currentState.Evaluate() // Return the outcome of the simulation
}
