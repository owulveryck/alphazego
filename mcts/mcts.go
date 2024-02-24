package mcts

import (
	"log"

	board "github.com/owulveryck/alphazego/board"
)

type MCTS struct{}

// RunMCTS runs the Monte Carlo Tree Search algorithm, taking the current game state as input.
// This implementation of MCTS is stateless, meaning it does not retain any information between calls.
func (m *MCTS) RunMCST(s board.State) board.State {
	// initialize a new node for this state.
	n := &MCTSNode{
		state:    s, // Current game state
		parent:   nil,
		children: []*MCTSNode{}, // Initialize without any children
		wins:     0,             // No wins initially
		visits:   0,             // No visits initially
	}

	var winRate float64 // Placeholder for the win rate calculation

	// Continue the search until the win rate of the current node is satisfactory (e.g., < 95%).
	// This loop selects the best child based on the UCB1 formula, expands the tree, simulates games from the new nodes,
	// and backpropagates the results to update the statistics of the nodes.
	i := 0
	for i < 500 {
		nn := n.SelectChild()    // Select the best child to explore based on UCB1.
		nn.Expand()              // Expand the tree by adding a new child node for an unexplored move.
		result := nn.Simulate()  // Simulate a random playthrough from the new node to a terminal state.
		nn.Backpropagate(result) // Update the node and its ancestors based on the simulation outcome.
		log.Printf("result: %v, wins: %v, visits: %v, winRate: %v", result, n.wins, n.visits, winRate)
		winRate = n.wins / n.visits // Update win rate after backpropagation.
		i++
	}
	winRate = 0

	var bestState board.State
	for _, n := range n.children {
		var currWinRate float64
		// Calculate the win rate if the node has been visited before.
		// Win rate is calculated as the ratio of wins to total visits.
		if n.visits == 0 {
			currWinRate = 0 // Avoid division by zero for unvisited nodes.
		} else {
			currWinRate = n.wins / n.visits // 𝑊𝑖𝑛𝑅𝑎𝑡𝑒 = 𝑊𝑖𝑛𝑠 / 𝑉𝑖𝑠𝑖𝑡𝑠
		}
		if currWinRate > winRate {
			bestState = n.state
			winRate = currWinRate
		}
	}

	// Return the state associated with the node that has been determined to be the best move.
	return bestState
}
