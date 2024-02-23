package mcts

import board "github.com/owulveryck/alphazego/board"

type MCTS struct {
	inventory map[board.State]*MCTSNode
}

// RunMCTS runs the Monte Carlo Tree Search algorithm, taking the current game state as input.
// This implementation of MCTS is stateless, meaning it does not retain any information between calls.
func (m *MCTS) RunMCST(s board.State) board.State {
	var n *MCTSNode // Placeholder for the current node
	var ok bool     // Flag to check existence in inventory

	// Check if the current state is already in the inventory (cache of visited states).
	// If not, initialize a new node for this state.
	if n, ok = m.inventory[s]; !ok {
		n = &MCTSNode{
			state:    s,             // Current game state
			parent:   &MCTSNode{},   // Placeholder parent node
			children: []*MCTSNode{}, // Initialize without any children
			wins:     0,             // No wins initially
			visits:   0,             // No visits initially
		}
		// Add the new node to the inventory for future reference.
		m.inventory[s] = n
	}

	var winRate float64 // Placeholder for the win rate calculation
	// Calculate the win rate if the node has been visited before.
	// Win rate is calculated as the ratio of wins to total visits.
	if n.visits == 0 {
		winRate = 0 // Avoid division by zero for unvisited nodes.
	} else {
		winRate = n.wins / n.visits // 𝑊𝑖𝑛𝑅𝑎𝑡𝑒 = 𝑊𝑖𝑛𝑠 / 𝑉𝑖𝑠𝑖𝑡𝑠
	}

	// Continue the search until the win rate of the current node is satisfactory (e.g., < 95%).
	// This loop selects the best child based on the UCB1 formula, expands the tree, simulates games from the new nodes,
	// and backpropagates the results to update the statistics of the nodes.
	for winRate < 0.95 {
		n = n.SelectChild()         // Select the best child to explore based on UCB1.
		n.Expand()                  // Expand the tree by adding a new child node for an unexplored move.
		result := n.Simulate()      // Simulate a random playthrough from the new node to a terminal state.
		n.Backpropagate(result)     // Update the node and its ancestors based on the simulation outcome.
		winRate = n.wins / n.visits // Update win rate after backpropagation.
	}

	// Return the state associated with the node that has been determined to be the best move.
	return n.state
}
