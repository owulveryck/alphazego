package mcts

// Expand adds one new child node for an untried move from the current game state.
// It returns the newly created child node, or nil if no untried moves remain.
func (node *MCTSNode) Expand() *MCTSNode {
	possibleMoves := node.state.PossibleMoves()

	// Build a set of already-expanded board IDs
	existingIDs := make(map[string]bool)
	for _, child := range node.children {
		existingIDs[string(child.state.ID())] = true
	}

	// Find the first untried move and expand it
	for _, move := range possibleMoves {
		if !existingIDs[string(move.ID())] {
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

// ExpandAll cree des noeuds enfants pour tous les coups possibles,
// en leur attribuant les priors fournis par le policy network.
// Cette methode est utilisee par le chemin AlphaZero, ou le reseau
// retourne une distribution de probabilites sur tous les coups legaux
// en un seul appel. Les priors doivent etre dans le meme ordre que
// [board.State.PossibleMoves].
func (node *MCTSNode) ExpandAll(policy []float64) {
	possibleMoves := node.state.PossibleMoves()
	for i, move := range possibleMoves {
		child := &MCTSNode{
			state:    move,
			parent:   node,
			children: []*MCTSNode{},
			prior:    policy[i],
			mcts:     node.mcts,
		}
		node.children = append(node.children, child)
	}
}
