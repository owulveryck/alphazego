package mcts

import "math"

// UCB1
func (node *MCTSNode) UCB1() float64 {
	if node.visits == 0 {
		return math.Inf(1) // Return positive infinity to prioritize unvisited nodes
	}
	return node.wins/node.visits + math.Sqrt(2*math.Log(node.parent.visits)/node.visits)
}
