package mcts

import (
	"github.com/owulveryck/alphazego/gamestate"
)

type MCTSNode struct {
	state        gamestate.GameState
	parent       *MCTSNode
	children     []*MCTSNode
	wins         float64
	visits       float64
	untriedMoves []gamestate.GameState
}

// MCTSNode methods
