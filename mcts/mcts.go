package mcts

import (
	"log"

	board "github.com/owulveryck/alphazego/board"
)

// NewMCTS initializes a new MCTS structure.
// The inventory map stores nodes encountered during the search,
// allowing reuse if the same game state is reached via different paths (transpositions)
// within a single RunMCTS call.
func NewMCTS() *MCTS {
	return &MCTS{
		inventory: make(map[string]*MCTSNode),
	}
}

// MCTS holds the state for the Monte Carlo Tree Search.
// En mode MCTS pur (cree par [NewMCTS]), le champ evaluator est nil et
// l'algorithme utilise des rollouts aleatoires avec UCB1.
// En mode AlphaZero (cree par [NewAlphaMCTS]), l'evaluator fournit policy et
// value, et la selection utilise PUCT.
type MCTS struct {
	inventory map[string]*MCTSNode // Stores nodes by their state ID for potential reuse within a search.
	evaluator board.Evaluator      // nil = MCTS pur, non-nil = AlphaZero
	cpuct     float64              // constante d'exploration pour PUCT (utilise uniquement avec evaluator)
}

// NewAlphaMCTS initialise un MCTS guide par un reseau de neurones (style AlphaZero).
// L'evaluateur fournit une policy (priors) et une value pour chaque position,
// remplacant les rollouts aleatoires. Le parametre cpuct controle l'exploration
// dans la formule PUCT (typiquement entre 1.0 et 5.0).
func NewAlphaMCTS(eval board.Evaluator, cpuct float64) *MCTS {
	return &MCTS{
		inventory: make(map[string]*MCTSNode),
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

// GetOrCreateNode retrieves a node from the inventory or creates a new one if it doesn't exist.
func (m *MCTS) GetOrCreateNode(s board.State, parent *MCTSNode) *MCTSNode {
	boardID := string(s.ID())
	if node, ok := m.inventory[boardID]; ok {
		// TODO: Potentially update parent if a shorter path is found? Or handle graph structure explicitly.
		// For now, just return the existing node.
		return node
	}

	// Node not found, create a new one
	newNode := &MCTSNode{
		state:    s,
		parent:   parent,
		children: []*MCTSNode{}, // Initialize empty
		wins:     0,
		visits:   0,
		// untriedActions: s.GetPossibleActions(), // Assuming state can provide actions
		mcts: m, // Pass reference to MCTS for inventory access during expansion
	}
	m.inventory[boardID] = newNode
	return newNode
}

// RunMCTS runs the Monte Carlo Tree Search algorithm for a specified number of iterations.
// It takes the current game state 's' and the number of iterations 'iterations' as input.
// It returns the state resulting from the best move found.
func (m *MCTS) RunMCTS(s board.State, iterations int) board.State {
	// 1. Create or retrieve the root node for the current state.
	root := m.GetOrCreateNode(s, nil) // Root has no parent

	// 2. Perform MCTS iterations
	for i := 0; i < iterations; i++ {
		// a. Selection: Start from root, traverse down the tree using UCB1 until a leaf node is found.
		// A leaf node is one that is terminal or not fully expanded.
		node := root
		for !node.IsTerminal() && node.IsFullyExpanded() {
			child := node.SelectChildUCB() // Select best child based on UCB
			if child == nil {
				// Should not happen if IsFullyExpanded is true and not terminal, but handle defensively.
				log.Printf("Warning: SelectChildUCB returned nil for non-terminal, fully expanded node %v", string(node.state.ID()))
				break // Stop traversal for this iteration
			}
			node = child
		}

		// b. Expansion + Evaluation + Backpropagation
		if !node.IsTerminal() && !node.IsFullyExpanded() {
			if m.evaluator != nil {
				// AlphaZero path: call the evaluator once to get policy + value,
				// expand all children with priors, backpropagate the value.
				policy, value := m.evaluator.Evaluate(node.state)
				node.ExpandAll(policy)
				node.BackpropagateValue(value)
			} else {
				// Pure MCTS path: expand one child, random rollout, backpropagate result.
				expandedNode := node.Expand()
				if expandedNode == nil {
					log.Printf("Warning: Expansion failed for non-terminal, non-fully-expanded node %v", string(node.state.ID()))
					expandedNode = node
				}
				result := expandedNode.Simulate()
				expandedNode.Backpropagate(result)
			}
		} else {
			// Terminal or fully expanded node.
			if m.evaluator != nil {
				value := terminalValue(node.state)
				node.BackpropagateValue(value)
			} else {
				result := node.Simulate()
				node.Backpropagate(result)
			}
		}
	}

	// 3. Select the best move from the root node's children.
	// Typically, this is the child with the highest visit count, as it's the most explored path.
	bestChild := root.SelectBestMove() // Implement this method in node.go (usually max visits)

	if bestChild == nil {
		log.Println("Warning: No best child found after MCTS, returning original state.")
		// This might happen if iterations = 0, the root is terminal, or no moves are possible.
		// Consider returning an error or a specific indicator if no move is possible.
		return s // Return the original state as no move could be determined.
	}

	log.Printf("MCTS finished. Root visits: %f. Best child visits: %f, wins: %f", root.visits, bestChild.visits, bestChild.wins)
	// Return the state associated with the best child node.
	return bestChild.state
}
