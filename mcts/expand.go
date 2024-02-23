package mcts

// Expand adds new child nodes to the current node for each untried move in the game state.
// This process increases the breadth of the search tree, allowing the algorithm to explore
// more potential future states of the game.
func (node *MCTSNode) Expand() {
	// Retrieve a list of all possible moves from the current game state.
	// These moves are potential actions that can be taken from this state,
	// leading to new game states that have not yet been explored in the search tree.
	unexploredMoves := node.state.PossibleMoves()

	// Iterate through each unexplored move to create new child nodes.
	// Each child node represents a potential future state of the game
	// that results from making one of the possible moves.
	for _, move := range unexploredMoves {
		// Create a new MCTSNode for the new state. This child node starts with zero wins
		// and zero visits, indicating that it is a new, unexplored node. The parent of this
		// new node is set to the current node (`node`), establishing the link in the tree structure.
		child := &MCTSNode{
			state:    move,
			parent:   node,
			wins:     0,
			visits:   0,
			children: []*MCTSNode{}, // Initialize with no children, as this is a newly expanded node.
		}

		// Add the newly created child node to the list of children of the current node.
		// This step effectively expands the tree, adding a new branch for each unexplored move.
		node.children = append(node.children, child)
	}
}
