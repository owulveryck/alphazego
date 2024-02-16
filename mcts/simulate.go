package mcts

import "math/rand"

func (node *MCTSNode) Simulate() uint8 {
	// Simulate a random playthrough from this node to a terminal state
	currentState := node.state
	for !currentState.IsGameOver() {
		possibleMoves := currentState.PossibleMoves()
		move := possibleMoves[rand.Intn(len(possibleMoves))] // Randomly select a move
		currentState = currentState.MakeMove(move)           // Apply the move
	}
	return currentState.GetWinner() // Return the outcome of the simulation
}
