package mcts

import (
	"github.com/owulveryck/alphazego/board"
)

// NewMCTS initializes a new MCTS structure.
// The inventory map stores nodes encountered during the search,
// allowing reuse if the same game state is reached via different paths (transpositions)
// within a single RunMCTS call.
func NewMCTS() *MCTS {
	return &MCTS{
		inventory: make(map[string]*mctsNode),
	}
}

// MCTS holds the state for the Monte Carlo Tree Search.
// En mode MCTS pur (cree par [NewMCTS]), le champ evaluator est nil et
// l'algorithme utilise des rollouts aleatoires avec UCB1.
// En mode AlphaZero (cree par [NewAlphaMCTS]), l'evaluator fournit policy et
// value, et la selection utilise PUCT.
type MCTS struct {
	inventory map[string]*mctsNode // Stores nodes by their state ID for potential reuse within a search.
	evaluator board.Evaluator      // nil = MCTS pur, non-nil = AlphaZero
	cpuct     float64              // constante d'exploration pour PUCT (utilise uniquement avec evaluator)
}

// NewAlphaMCTS initialise un MCTS guide par un reseau de neurones (style AlphaZero).
// L'evaluateur fournit une policy (priors) et une value pour chaque position,
// remplacant les rollouts aleatoires. Le parametre cpuct controle l'exploration
// dans la formule PUCT (typiquement entre 1.0 et 5.0).
func NewAlphaMCTS(eval board.Evaluator, cpuct float64) *MCTS {
	return &MCTS{
		inventory: make(map[string]*mctsNode),
		evaluator: eval,
		cpuct:     cpuct,
	}
}

// terminalValue convertit le resultat d'un etat terminal en valeur continue
// du point de vue du joueur courant (celui qui est a jouer).
// Retourne 1.0 si le joueur courant a gagne, -1.0 s'il a perdu, 0.0 pour un nul.
func terminalValue(s board.State) float64 {
	result := s.Evaluate()
	// Le joueur qui a joue le dernier coup
	playerWhoMovedHere := s.PreviousPlayer()
	if result == playerWhoMovedHere {
		// L'adversaire (qui a joue le dernier coup) a gagne → defaite pour le joueur courant
		return -1.0
	}
	if result == board.DrawResult {
		return 0.0
	}
	// Le joueur courant a gagne (cas rare dans un etat terminal ou c'est a lui de jouer)
	return 1.0
}

// getOrCreateNode retrieves a node from the inventory or creates a new one if it doesn't exist.
func (m *MCTS) getOrCreateNode(s board.State, parent *mctsNode) *mctsNode {
	boardID := s.ID()
	if node, ok := m.inventory[boardID]; ok {
		return node
	}

	newNode := &mctsNode{
		state:    s,
		parent:   parent,
		children: []*mctsNode{},
		mcts:     m,
	}
	m.inventory[boardID] = newNode
	return newNode
}

// RunMCTS runs the Monte Carlo Tree Search algorithm for a specified number of iterations.
// It takes the current game state 's' and the number of iterations 'iterations' as input.
// It returns the state resulting from the best move found.
func (m *MCTS) RunMCTS(s board.State, iterations int) board.State {
	root := m.getOrCreateNode(s, nil)

	for i := 0; i < iterations; i++ {
		// Selection: descend the tree using UCB until a leaf node is found.
		node := root
		for !node.isTerminal() && node.isFullyExpanded() {
			child := node.selectChildUCB()
			if child == nil {
				break
			}
			node = child
		}

		// Expansion + Evaluation + Backpropagation
		if !node.isTerminal() && !node.isFullyExpanded() {
			if m.evaluator != nil {
				policy, value := m.evaluator.Evaluate(node.state)
				node.expandAll(policy)
				node.backpropagateValue(value)
			} else {
				expandedNode := node.expand()
				if expandedNode == nil {
					expandedNode = node
				}
				result := expandedNode.simulate()
				expandedNode.backpropagate(result)
			}
		} else {
			if m.evaluator != nil {
				value := terminalValue(node.state)
				node.backpropagateValue(value)
			} else {
				result := node.simulate()
				node.backpropagate(result)
			}
		}
	}

	bestChild := root.selectBestMove()
	if bestChild == nil {
		return s
	}

	return bestChild.state
}
