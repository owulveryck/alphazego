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

// expandAll cree des noeuds enfants pour tous les coups possibles,
// en leur attribuant les priors fournis par le policy network.
// Cette methode est utilisee par le chemin AlphaZero, ou le reseau
// retourne une distribution de probabilites sur tous les coups legaux
// en un seul appel. Les priors doivent etre dans le meme ordre que
// [board.State.PossibleMoves].
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
