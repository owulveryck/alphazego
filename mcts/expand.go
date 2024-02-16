package mcts

func (node *MCTSNode) Expand() *MCTSNode {
	// Expand the tree by creating a new child node for one of the untried moves
	unexploredMoves := node.state.PossibleMoves() // Assume this returns a list of game states
	for _, move := range unexploredMoves {
		newState := node.state.MakeMove(move) // Assume this applies the move and returns a new state
		child := &MCTSNode{
			state:    newState,
			parent:   node,
			wins:     0,
			visits:   0,
			children: []*MCTSNode{}, // No children yet
		}
		node.children = append(node.children, child)
	}
	return nil
}
