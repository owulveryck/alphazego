package mcts

import (
	"errors"
	"fmt"
)

// expand adds one new child node for an untried move from the current game state.
// It returns the newly created child node, or nil if no untried moves remain.
//
// L'expansion utilise un index incrémental (expandedIndex) plutôt qu'une
// détection de doublons par ID. Le prochain coup à explorer est simplement
// possibleMoves[expandedIndex]. Cela est correct car :
//
//  1. getPossibleMoves() est cachée sur le nœud (cachedMoves) et retourne
//     toujours le même slice dans le même ordre.
//  2. Le contrat de decision.State.PossibleMoves() garantit un ordre
//     déterministe pour un état donné.
//  3. expand() est le seul chemin qui ajoute des enfants un par un
//     (expandAll utilise son propre mécanisme).
//
// Les nœuds enfants sont alloués par batch via MCTS.allocNode() pour
// réduire la pression GC (660K allocs → ~2600 pour RunMCTS_10000).
func (node *mctsNode) expand() *mctsNode {
	possibleMoves := node.getPossibleMoves()
	if node.expandedIndex >= len(possibleMoves) {
		return nil
	}

	// Pré-allouer le slice children au premier expand pour éviter les resize.
	if node.children == nil {
		node.children = make([]*mctsNode, 0, len(possibleMoves))
	}
	childState := possibleMoves[node.expandedIndex]
	child := node.mcts.allocNode()
	*child = mctsNode{
		state:         childState,
		parent:        node,
		mcts:          node.mcts,
		previousActor: childState.PreviousActor(),
	}
	node.children = append(node.children, child)
	node.expandedIndex++
	return child
}

// expandAll crée des nœuds enfants pour tous les coups possibles,
// en leur attribuant les priors fournis par le policy network.
// Cette méthode est utilisée par le chemin AlphaZero, où le réseau
// retourne une distribution de probabilités sur tous les coups légaux
// en un seul appel. Les priors doivent être dans le même ordre que
// [decision.State.PossibleMoves].
//
// Retourne une erreur si la taille de policy ne correspond pas au nombre
// de coups possibles, ou si policy est nil.
func (node *mctsNode) expandAll(policy []float64) error {
	possibleMoves := node.getPossibleMoves()
	if policy == nil {
		return errors.New("mcts: policy is nil")
	}
	if len(policy) != len(possibleMoves) {
		return fmt.Errorf("mcts: policy length %d does not match possible moves count %d", len(policy), len(possibleMoves))
	}
	// Vérifier que la policy est approximativement normalisée.
	// La normalisation est la responsabilité de l'Evaluator. Une policy
	// mal normalisée biaisera les scores PUCT mais n'est pas une erreur
	// fatale. Pas de log ici : log.Printf est incontrôlable en code
	// bibliothèque et pollue les benchmarks.

	node.children = make([]*mctsNode, 0, len(possibleMoves))
	for i, move := range possibleMoves {
		child := node.mcts.allocNode()
		*child = mctsNode{
			state:         move,
			parent:        node,
			prior:         policy[i],
			mcts:          node.mcts,
			previousActor: move.PreviousActor(),
		}
		node.children = append(node.children, child)
	}
	node.expandedIndex = len(possibleMoves)
	return nil
}
