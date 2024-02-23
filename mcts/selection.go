package mcts

import "math"

// SelectChild iteratively selects the best child node based on the UCB1 score,
// aiming to balance exploration and exploitation.
func (node *MCTSNode) SelectChild() *MCTSNode {
	// If the current node has no children, it's either a leaf node or
	// exploration has not yet expanded this part of the tree.
	if len(node.children) == 0 {
		return node // Returns the node itself if it has no children
	}

	bestScore := math.Inf(-1) // Initialize bestScore with negative infinity
	var bestChild *MCTSNode   // Pointer to track the child node with the best UCB1 score

	// Iterate over all children to find the one with the highest UCB1 score
	for _, child := range node.children {
		score := child.UCB1() // Calculate the UCB1 score for the child
		if score > bestScore {
			bestScore = score // Update the best score found so far
			bestChild = child // Update the best child based on the new best score
		}
	}

	// After identifying the best child based on UCB1 score, we recursively
	// select nodes down the tree until a leaf node is reached.
	// This process ensures a path is chosen that balances between exploring
	// less visited nodes and exploiting nodes with high average rewards.
	return bestChild.SelectChild() // Corrected to SelectChild for consistency and clarity
}
