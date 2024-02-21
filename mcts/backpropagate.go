package mcts

func (node *MCTSNode) Backpropagate(result uint8) {
	// Loop to update nodes up to the root
	for n := node; n != nil; n = n.parent {
		n.visits += 1
		// If the result matches the playerTurn of this node, it's a win for this node
		if n.state.CurrentPlayer() == result {
			n.wins += 1
		}
		// For Tic-Tac-Toe, you might also need to handle draws specifically
		// depending on how you want to treat them in your win/loss statistics
	}
	// Update this node and its ancestors with the simulation result
}
