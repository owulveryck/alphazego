package mcts

import (
	"math"

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

	// prior est la probabilite a priori P(s,a) attribuee par le policy network.
	// En MCTS pur, cette valeur est 0 (non utilisee). En mode AlphaZero,
	// elle est fixee lors de l'expansion par l'Evaluator et utilisee dans la
	// formule PUCT pour guider la selection.
	prior float64

	// mcts holds a reference back to the MCTS instance for inventory access during expansion.
	mcts *MCTS
}

// IsTerminal returns true if this node represents a terminal game state (win, loss, or draw).
func (n *MCTSNode) IsTerminal() bool {
	return n.state.Evaluate() != board.NoPlayer
}

// IsFullyExpanded returns true if all possible moves from this state have been expanded as children.
func (n *MCTSNode) IsFullyExpanded() bool {
	return len(n.children) >= len(n.state.PossibleMoves())
}

// SelectChildUCB selects the immediate child with the highest score.
// When an [Evaluator] is configured, it uses [MCTSNode.PUCT] (with prior probabilities).
// Otherwise, it uses [MCTSNode.UCB1] (pure MCTS).
func (n *MCTSNode) SelectChildUCB() *MCTSNode {
	bestScore := math.Inf(-1)
	var bestChild *MCTSNode
	for _, child := range n.children {
		var score float64
		if n.mcts != nil && n.mcts.evaluator != nil {
			score = child.PUCT()
		} else {
			score = child.UCB1()
		}
		if score > bestScore {
			bestScore = score
			bestChild = child
		}
	}
	return bestChild
}

// SelectBestMove returns the child with the highest visit count (most explored path).
func (n *MCTSNode) SelectBestMove() *MCTSNode {
	var bestChild *MCTSNode
	bestVisits := float64(-1)
	for _, child := range n.children {
		if child.visits > bestVisits {
			bestVisits = child.visits
			bestChild = child
		}
	}
	return bestChild
}
