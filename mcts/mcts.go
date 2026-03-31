package mcts

import (
	"log"

	board "github.com/owulveryck/alphazego/board"
)

// NewMCTS initializes a new MCTS structure.
// The inventory map stores nodes encountered during the search,
// allowing reuse if the same game state is reached via different paths (transpositions)
// within a single RunMCTS call.
func NewMCTS() *MCTS {
	return &MCTS{
		inventory: make(map[string]*MCTSNode),
	}
}

// MCTS holds the state for the Monte Carlo Tree Search.
type MCTS struct {
	inventory map[string]*MCTSNode // Stores nodes by their state BoardID for potential reuse within a search.
}

// GetOrCreateNode retrieves a node from the inventory or creates a new one if it doesn't exist.
func (m *MCTS) GetOrCreateNode(s board.State, parent *MCTSNode) *MCTSNode {
	boardID := string(s.BoardID())
	if node, ok := m.inventory[boardID]; ok {
		// TODO: Potentially update parent if a shorter path is found? Or handle graph structure explicitly.
		// For now, just return the existing node.
		return node
	}

	// Node not found, create a new one
	newNode := &MCTSNode{
		state:    s,
		parent:   parent,
		children: []*MCTSNode{}, // Initialize empty
		wins:     0,
		visits:   0,
		// untriedActions: s.GetPossibleActions(), // Assuming state can provide actions
		mcts: m, // Pass reference to MCTS for inventory access during expansion
	}
	m.inventory[boardID] = newNode
	return newNode
}

// RunMCTS runs the Monte Carlo Tree Search algorithm for a specified number of iterations.
// It takes the current game state 's' and the number of iterations 'iterations' as input.
// It returns the state resulting from the best move found.
func (m *MCTS) RunMCTS(s board.State, iterations int) board.State {
	// 1. Create or retrieve the root node for the current state.
	root := m.GetOrCreateNode(s, nil) // Root has no parent

	// 2. Perform MCTS iterations
	for i := 0; i < iterations; i++ {
		// a. Selection: Start from root, traverse down the tree using UCB1 until a leaf node is found.
		// A leaf node is one that is terminal or not fully expanded.
		node := root
		for !node.IsTerminal() && node.IsFullyExpanded() {
			child := node.SelectChildUCB() // Select best child based on UCB
			if child == nil {
				// Should not happen if IsFullyExpanded is true and not terminal, but handle defensively.
				log.Printf("Warning: SelectChildUCB returned nil for non-terminal, fully expanded node %v", string(node.state.BoardID()))
				break // Stop traversal for this iteration
			}
			node = child
		}

		// b. Expansion: If the selected node 'node' is not terminal and not fully expanded, expand it by adding one child.
		var nodeToSimulate *MCTSNode
		if !node.IsTerminal() && !node.IsFullyExpanded() {
			// Expand creates and returns the new child node
			expandedNode := node.Expand() // Expand should add the node to m.inventory
			if expandedNode != nil {
				nodeToSimulate = expandedNode
			} else {
				// Expansion failed unexpectedly (e.g., no more valid moves found), simulate from current node.
				log.Printf("Warning: Expansion failed for non-terminal, non-fully-expanded node %v", string(node.state.BoardID()))
				nodeToSimulate = node
			}
		} else {
			// If the node was terminal or already fully expanded (e.g., hit during selection), simulate from it.
			nodeToSimulate = node
		}

		// c. Simulation (Rollout): Simulate a random playout from the 'nodeToSimulate'.
		// The result should be from the perspective of the player whose turn it is in nodeToSimulate.state.
		result := nodeToSimulate.Simulate()

		// d. Backpropagation: Update visit counts and win statistics back up the tree from 'nodeToSimulate' to the root.
		nodeToSimulate.Backpropagate(result)
	}

	// 3. Select the best move from the root node's children.
	// Typically, this is the child with the highest visit count, as it's the most explored path.
	bestChild := root.SelectBestMove() // Implement this method in node.go (usually max visits)

	if bestChild == nil {
		log.Println("Warning: No best child found after MCTS, returning original state.")
		// This might happen if iterations = 0, the root is terminal, or no moves are possible.
		// Consider returning an error or a specific indicator if no move is possible.
		return s // Return the original state as no move could be determined.
	}

	log.Printf("MCTS finished. Root visits: %f. Best child visits: %f, wins: %f", root.visits, bestChild.visits, bestChild.wins)
	// Return the state associated with the best child node.
	return bestChild.state
}
