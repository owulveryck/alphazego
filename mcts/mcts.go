package mcts

import (
	"math/rand"

	"github.com/owulveryck/alphazego/decision"
)

// nodeBatchSize est la taille des blocs pré-alloués de mctsNode.
// Un batch de 256 nœuds réduit le nombre d'allocations heap de N à N/256,
// tout en conservant une localité mémoire favorable au cache CPU.
const nodeBatchSize = 256

// NewMCTS initializes a new MCTS structure.
// Chaque nœud est créé indépendamment (pas de table de transposition)
// pour que la backpropagation remonte correctement via parent.
func NewMCTS() *MCTS {
	return &MCTS{
		rng: rand.New(rand.NewSource(rand.Int63())),
	}
}

// MCTS holds the state for the Monte Carlo Tree Search.
// En mode MCTS pur (créé par [NewMCTS]), le champ evaluator est nil et
// l'algorithme utilise des rollouts aléatoires avec UCB1.
// En mode AlphaZero (créé par [NewAlphaMCTS]), l'evaluator fournit policy et
// value, et la sélection utilise PUCT.
//
// MCTS n'est pas thread-safe : chaque goroutine doit utiliser sa propre instance.
type MCTS struct {
	evaluator Evaluator  // nil = MCTS pur, non-nil = AlphaZero
	cpuct     float64    // constante d'exploration pour PUCT (utilisé uniquement avec evaluator)
	rng       *rand.Rand // générateur aléatoire pour les rollouts (reproductibilité)
	// Allocateur par batch pour les mctsNode. Réduit les allocations
	// individuelles (~140 B/nœud) en blocs contigus de nodeBatchSize.
	nodeBatch []mctsNode
	nodeIdx   int
}

// NewAlphaMCTS initialise un MCTS guidé par un réseau de neurones (style AlphaZero).
// L'évaluateur fournit une policy (priors) et une value pour chaque position,
// remplaçant les rollouts aléatoires. Le paramètre cpuct contrôle l'exploration
// dans la formule PUCT (typiquement entre 1.0 et 5.0).
func NewAlphaMCTS(eval Evaluator, cpuct float64) *MCTS {
	return &MCTS{
		evaluator: eval,
		cpuct:     cpuct,
		rng:       rand.New(rand.NewSource(rand.Int63())),
	}
}

// allocNode retourne un *mctsNode pré-alloué depuis le batch courant.
// Quand le batch est épuisé, un nouveau batch est alloué en une seule
// allocation heap, réduisant la pression GC.
func (m *MCTS) allocNode() *mctsNode {
	if m.nodeIdx >= len(m.nodeBatch) {
		m.nodeBatch = make([]mctsNode, nodeBatchSize)
		m.nodeIdx = 0
	}
	node := &m.nodeBatch[m.nodeIdx]
	m.nodeIdx++
	return node
}

// newNode crée un nouveau nœud pour l'état donné.
func (m *MCTS) newNode(s decision.State, parent *mctsNode) *mctsNode {
	node := m.allocNode()
	*node = mctsNode{
		state:         s,
		parent:        parent,
		mcts:          m,
		previousActor: s.PreviousActor(),
	}
	return node
}

// RunMCTS runs the Monte Carlo Tree Search algorithm for a specified number of iterations.
// It takes the current state 's' and the number of iterations 'iterations' as input.
// It returns the state resulting from the best move found.
func (m *MCTS) RunMCTS(s decision.State, iterations int) decision.State {
	root := m.newNode(s, nil)

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
				if err := node.expandAll(policy); err != nil {
					continue
				}
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
				// Le nœud est terminal : utiliser le résultat caché par isTerminal()
				// au lieu de relancer simulate() qui rappellerait Evaluate().
				node.backpropagate(node.cachedEval)
			}
		}
	}

	bestChild := root.selectBestMove()
	if bestChild == nil {
		return s
	}

	return bestChild.state
}
