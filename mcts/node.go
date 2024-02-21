package mcts

import (
	"github.com/owulveryck/alphazego/board"
)

type MCTSNode struct {
	state    board.State
	parent   *MCTSNode
	children []*MCTSNode
	wins     float64
	visits   float64
}

// MCTSNode methods
