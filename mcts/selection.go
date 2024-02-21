package mcts

import "math"

func (node *MCTSNode) SelectChild() *MCTSNode {
	// C is the exploration parameter, which determines the balance between exploration and exploitation. A higher value of encourages more exploration.
	// Common values for are in the range of sqrt(2), but the optimal value can depend on the specific game and context
	var C float64 = 1
	var bestScore float64 = -1
	var bestChild *MCTSNode

	for _, child := range node.children {
		wins := child.wins
		if node.state.CurrentPlayer() != child.state.CurrentPlayer() {
			// Adjust for the perspective of the player to make a move
			wins = child.visits - child.wins
		}
		ucb1 := wins/child.visits + C*math.Sqrt(math.Log(node.visits)/child.visits)
		if ucb1 > bestScore {
			bestScore = ucb1
			bestChild = child
		}
	}
	return bestChild
}
