package mcts

import (
	"github.com/owulveryck/alphazego/decision"
)

// backpropagate updates the statistics for this node and its ancestors up to the root node
// after a game simulation is completed. The statistics updated include the number of visits
// and wins, which are used to calculate the node's value in future selections.
func (node *mctsNode) backpropagate(result decision.ActorID) {
	// Starting from the current node, loop through all ancestors until the root node is reached.
	// The loop uses 'n' to traverse the tree upwards, with 'n.parent' moving to each parent node.
	for n := node; n != nil; n = n.parent {
		n.visits += 1

		// Credit wins to the actor who made the move leading to this node.
		// PreviousActor() retourne l'acteur qui a effectué l'action menant à cet état,
		// quelle que soit la logique de tour (2 acteurs, N acteurs, etc.).
		actorWhoMovedHere := n.state.PreviousActor()
		if result == actorWhoMovedHere {
			n.wins += 1
		} else if result == decision.DrawResult {
			n.wins += 0.5
		}
	}

	// This method systematically updates the visit and win counts for each node from the
	// current node back to the root. These updated statistics influence the selection of
	// nodes in future iterations of the MCTS, guiding the search towards more promising paths.
}

// backpropagateValue propage une valeur continue v ∈ [-1, 1] depuis ce nœud
// jusqu'à la racine. Cette méthode est utilisée par le chemin AlphaZero, où la
// value provient du réseau de neurones au lieu d'un rollout aléatoire.
//
// La valeur initiale est exprimée du point de vue de l'acteur courant au nœud
// évalué (CurrentActor). Elle est d'abord inversée pour être stockée du point
// de vue de l'acteur qui a effectué l'action menant à ce nœud (convention
// cohérente avec backpropagate), puis alternée à chaque niveau en remontant.
//
// L'inversion de signe suppose un jeu à somme nulle à deux acteurs. Pour les
// jeux à N acteurs (N > 2), utiliser backpropagate avec un [decision.ActorID]
// discret.
func (node *mctsNode) backpropagateValue(value float64) {
	// Inverser pour passer de la perspective de l'acteur courant à celle de l'acteur
	// qui a effectué l'action menant à ce nœud (= PreviousActor).
	value = -value
	for n := node; n != nil; n = n.parent {
		n.visits++
		n.wins += value
		value = -value
	}
}
