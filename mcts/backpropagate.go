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
		n.visits += 1

		// Credit wins to the player who made the move leading to this node.
		// CurrentPlayer() returns who is about to play, so the player who moved here
		// is the opponent: 3 - CurrentPlayer().
		// This ensures UCB1 correctly evaluates children from the parent's perspective.
		playerWhoMovedHere := 3 - n.state.CurrentPlayer()
		if result == playerWhoMovedHere {
			n.wins += 1
		} else if result == board.Draw {
			n.wins += 0.5
		}
	}

	// This method systematically updates the visit and win counts for each node from the
	// current node back to the root. These updated statistics influence the selection of
	// nodes in future iterations of the MCTS, guiding the search towards more promising paths.
}
