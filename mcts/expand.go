package mcts

// Expand adds one new child node for an untried move from the current game state.
// It returns the newly created child node, or nil if no untried moves remain.
func (node *MCTSNode) Expand() *MCTSNode {
	possibleMoves := node.state.PossibleMoves()

	// Build a set of already-expanded board IDs
	existingIDs := make(map[string]bool)
	for _, child := range node.children {
		existingIDs[string(child.state.BoardID())] = true
	}

	// Find the first untried move and expand it
	for _, move := range possibleMoves {
		if !existingIDs[string(move.BoardID())] {
			child := &MCTSNode{
				state:    move,
				parent:   node,
				children: []*MCTSNode{},
				mcts:     node.mcts,
			}
			node.children = append(node.children, child)
			return child
		}
	}
	return nil
}
