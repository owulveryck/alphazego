package mcts

import "fmt"

// expand adds one new child node for an untried move from the current game state.
// It returns the newly created child node, or nil if no untried moves remain.
// La recherche de doublons utilise un scan linéaire sur les enfants existants,
// ce qui est plus efficace qu'une map pour les branching factors typiques (< 20).
func (node *mctsNode) expand() *mctsNode {
	possibleMoves := node.getPossibleMoves()

	// Find the first untried move via linear scan of existing children
	for _, move := range possibleMoves {
		moveID := move.ID()
		found := false
		for _, child := range node.children {
			if child.state.ID() == moveID {
				found = true
				break
			}
		}
		if !found {
			child := &mctsNode{
				state:  move,
				parent: node,
				mcts:   node.mcts,
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
	possibleMoves := node.getPossibleMoves()
	if len(policy) != len(possibleMoves) {
		panic(fmt.Sprintf("mcts: policy length %d does not match possible moves count %d", len(policy), len(possibleMoves)))
	}
	node.children = make([]*mctsNode, 0, len(possibleMoves))
	for i, move := range possibleMoves {
		child := &mctsNode{
			state:  move,
			parent: node,
			prior:  policy[i],
			mcts:   node.mcts,
		}
		node.children = append(node.children, child)
	}
}
