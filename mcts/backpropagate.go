package mcts

import (
	"github.com/owulveryck/alphazego/board"
)

// Backpropagate updates the statistics for this node and its ancestors up to the root node
// after a game simulation is completed. The statistics updated include the number of visits
// and wins, which are used to calculate the node's value in future selections.
func (node *MCTSNode) Backpropagate(result board.Result) {
	// Starting from the current node, loop through all ancestors until the root node is reached.
	// The loop uses 'n' to traverse the tree upwards, with 'n.parent' moving to each parent node.
	for n := node; n != nil; n = n.parent {
		n.visits += 1 // Increment the visits count for each node on the path back to the root.

		// Check if the simulation result matches this node's player turn.
		// The assumption here is that 'result' is coded in a way to match the player's identifier
		// in the node's state, e.g., '1' for one player and '2' for the other in a two-player game.
		// If they match, it means this node (and thus the decision leading to it) was on the winning path.
		if n.state.CurrentPlayer() == result {
			n.wins += 1 // Increment the win count for the node.
		}
		// Optional: Handling draws.
		// Depending on your game's rules, you might need to handle draws explicitly.
		// This could involve checking if the result indicates a draw and then deciding
		// whether to count that as a half-win, a full win, or something else for the node.
		// Example:
		/*
			if result == board.Draw {
				n.wins += 0.5
			}
		*/
	}

	// This method systematically updates the visit and win counts for each node from the
	// current node back to the root. These updated statistics influence the selection of
	// nodes in future iterations of the MCTS, guiding the search towards more promising paths.
}
