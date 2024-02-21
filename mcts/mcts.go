package mcts

import board "github.com/owulveryck/alphazego/board"

type MCTS struct {
	inventory map[board.State]*MCTSNode
}

// RunMCTS runs the Tree Search Algorithm taking the state as input
// The MCTS is stateless in the sense that it does not store any session
func (m *MCTS) RunMCST(s board.State) board.State {
	var n *MCTSNode
	var ok bool
	if n, ok = m.inventory[s]; !ok {
		n = &MCTSNode{
			state:        s,
			parent:       &MCTSNode{},
			children:     []*MCTSNode{},
			wins:         0,
			visits:       0,
			untriedMoves: nil,
		}
	}
	n.Expand()
	n.SelectChild()
	return nil
}
