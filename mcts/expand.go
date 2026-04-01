package mcts

// expand adds one new child node for an untried move from the current game state.
// It returns the newly created child node, or nil if no untried moves remain.
func (node *mctsNode) expand() *mctsNode {
	possibleMoves := node.state.PossibleMoves()

	// Build a set of already-expanded board IDs
	existingIDs := make(map[string]bool)
	for _, child := range node.children {
		existingIDs[child.state.ID()] = true
	}

	// Find the first untried move and expand it
	for _, move := range possibleMoves {
		if !existingIDs[move.ID()] {
			child := &mctsNode{
				state:    move,
				parent:   node,
				children: []*mctsNode{},
				mcts:     node.mcts,
			}
			node.children = append(node.children, child)
			return child
		}
	}
	return nil
}

// expandAll crée des nœuds enfants pour tous les coups possibles,
// en leur attribuant les priors fournis par le policy network.
// Cette méthode est utilisée par le chemin AlphaZero, où le réseau
// retourne une distribution de probabilités sur tous les coups légaux
// en un seul appel. Les priors doivent être dans le même ordre que
// [decision.State.PossibleMoves].
func (node *mctsNode) expandAll(policy []float64) {
	possibleMoves := node.state.PossibleMoves()
	for i, move := range possibleMoves {
		child := &mctsNode{
			state:    move,
			parent:   node,
			children: []*mctsNode{},
			prior:    policy[i],
			mcts:     node.mcts,
		}
		node.children = append(node.children, child)
	}
}
