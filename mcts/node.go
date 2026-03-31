package mcts

import (
	"math"

	"github.com/owulveryck/alphazego/board"
)

// mctsNode represents a single node in the Monte Carlo Tree Search (MCTS) algorithm.
// Each node corresponds to a specific game state and contains statistical information
// about the outcomes of simulations that have been run through this node. The structure
// of the tree is formed by parent and child relationships between nodes, enabling the
// navigation and expansion of the search tree as the algorithm progresses.
type mctsNode struct {
	// state holds the current game state that this node represents.
	// The game state includes all necessary information to continue play or simulation
	// from this point, such as the board configuration, the player whose turn it is, etc.
	state board.State

	// parent is a pointer to the parent node in the search tree. The root node of the tree
	// will have a nil parent. This link is used to traverse back up the tree during the
	// backpropagation phase of the MCTS algorithm, updating statistics along the way.
	parent *mctsNode

	// children is a slice of pointers to the child nodes of this node. Each child represents
	// a possible future game state that can be reached from the current state. The children
	// are the result of expanding the search tree by exploring the outcomes of possible moves
	// from the current state.
	children []*mctsNode

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

	// prior est la probabilite a priori P(s,a) attribuee par le policy network.
	// En MCTS pur, cette valeur est 0 (non utilisee). En mode AlphaZero,
	// elle est fixee lors de l'expansion par l'Evaluator et utilisee dans la
	// formule PUCT pour guider la selection.
	prior float64

	// mcts holds a reference back to the MCTS instance for inventory access during expansion.
	mcts *MCTS
}

// isTerminal returns true if this node represents a terminal game state (win, loss, or draw).
func (n *mctsNode) isTerminal() bool {
	return n.state.Evaluate() != board.NoPlayer
}

// isFullyExpanded returns true if all possible moves from this state have been expanded as children.
func (n *mctsNode) isFullyExpanded() bool {
	return len(n.children) >= len(n.state.PossibleMoves())
}

// selectChildUCB selects the immediate child with the highest score.
// When an Evaluator is configured, it uses puct (with prior probabilities).
// Otherwise, it uses ucb1 (pure MCTS).
func (n *mctsNode) selectChildUCB() *mctsNode {
	bestScore := math.Inf(-1)
	var bestChild *mctsNode
	for _, child := range n.children {
		var score float64
		if n.mcts != nil && n.mcts.evaluator != nil {
			score = child.puct()
		} else {
			score = child.ucb1()
		}
		if score > bestScore {
			bestScore = score
			bestChild = child
		}
	}
	return bestChild
}

// selectBestMove returns the child with the highest visit count (most explored path).
func (n *mctsNode) selectBestMove() *mctsNode {
	var bestChild *mctsNode
	bestVisits := float64(-1)
	for _, child := range n.children {
		if child.visits > bestVisits {
			bestVisits = child.visits
			bestChild = child
		}
	}
	return bestChild
}
