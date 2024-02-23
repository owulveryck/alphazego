package mcts

import (
	"github.com/owulveryck/alphazego/board"
)

// MCTSNode represents a single node in the Monte Carlo Tree Search (MCTS) algorithm.
// Each node corresponds to a specific game state and contains statistical information
// about the outcomes of simulations that have been run through this node. The structure
// of the tree is formed by parent and child relationships between nodes, enabling the
// navigation and expansion of the search tree as the algorithm progresses.
type MCTSNode struct {
	// state holds the current game state that this node represents.
	// The game state includes all necessary information to continue play or simulation
	// from this point, such as the board configuration, the player whose turn it is, etc.
	state board.State

	// parent is a pointer to the parent node in the search tree. The root node of the tree
	// will have a nil parent. This link is used to traverse back up the tree during the
	// backpropagation phase of the MCTS algorithm, updating statistics along the way.
	parent *MCTSNode

	// children is a slice of pointers to the child nodes of this node. Each child represents
	// a possible future game state that can be reached from the current state. The children
	// are the result of expanding the search tree by exploring the outcomes of possible moves
	// from the current state.
	children []*MCTSNode

	// wins records the total number of wins (or other positive outcomes, depending on the
	// game and scoring system) observed in simulations that have passed through this node.
	// This value is used in conjunction with visits to calculate the node's value and
	// determine the most promising paths through the search tree.
	wins float64

	// visits records the total number of times this node has been visited during the
	// simulation phase of the MCTS algorithm. This includes both passing through the node
	// in simulations and selecting it during the selection phase. The visit count is used
	// to balance exploration and exploitation in the selection strategy, ensuring that
	// the search explores a wide range of moves while also concentrating on promising areas.
	visits float64
}
