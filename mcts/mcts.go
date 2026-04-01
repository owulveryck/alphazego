package mcts

import (
	"math/rand"

	"github.com/owulveryck/alphazego/decision"
)

// NewMCTS initializes a new MCTS structure.
// The inventory map stores nodes encountered during the search,
// allowing reuse if the same game state is reached via different paths (transpositions)
// within a single RunMCTS call.
func NewMCTS() *MCTS {
	return &MCTS{
		inventory:   make(map[string]*mctsNode),
		selectionFn: (*mctsNode).ucb1,
		rng:         rand.New(rand.NewSource(rand.Int63())),
	}
}

// selectionFunc calcule le score d'un nœud enfant pour la phase de sélection.
// UCB1 est utilisé en MCTS pur, PUCT en mode AlphaZero.
type selectionFunc func(child *mctsNode) float64

// MCTS holds the state for the Monte Carlo Tree Search.
// En mode MCTS pur (créé par [NewMCTS]), le champ evaluator est nil et
// l'algorithme utilise des rollouts aléatoires avec UCB1.
// En mode AlphaZero (créé par [NewAlphaMCTS]), l'evaluator fournit policy et
// value, et la sélection utilise PUCT.
type MCTS struct {
	inventory   map[string]*mctsNode // Stores nodes by their state ID for potential reuse within a search.
	evaluator   Evaluator            // nil = MCTS pur, non-nil = AlphaZero
	cpuct       float64              // constante d'exploration pour PUCT (utilisé uniquement avec evaluator)
	selectionFn selectionFunc        // stratégie de sélection (ucb1 ou puct)
	rng         *rand.Rand           // générateur aléatoire pour les rollouts (reproductibilité)
}

// NewAlphaMCTS initialise un MCTS guidé par un réseau de neurones (style AlphaZero).
// L'évaluateur fournit une policy (priors) et une value pour chaque position,
// remplaçant les rollouts aléatoires. Le paramètre cpuct contrôle l'exploration
// dans la formule PUCT (typiquement entre 1.0 et 5.0).
func NewAlphaMCTS(eval Evaluator, cpuct float64) *MCTS {
	return &MCTS{
		inventory:   make(map[string]*mctsNode),
		evaluator:   eval,
		cpuct:       cpuct,
		selectionFn: (*mctsNode).puct,
		rng:         rand.New(rand.NewSource(rand.Int63())),
	}
}

// getOrCreateNode retrieves a node from the inventory or creates a new one if it doesn't exist.
func (m *MCTS) getOrCreateNode(s decision.State, parent *mctsNode) *mctsNode {
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
// It takes the current state 's' and the number of iterations 'iterations' as input.
// It returns the state resulting from the best move found.
func (m *MCTS) RunMCTS(s decision.State, iterations int) decision.State {
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

		// Expansion + Évaluation + Backpropagation
		if !node.isTerminal() && !node.isFullyExpanded() {
			if m.evaluator != nil {
				policy, values := m.evaluator.Evaluate(node.state)
				node.expandAll(policy)
				node.backpropagateValue(values)
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
				node.backpropagateTerminal()
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
