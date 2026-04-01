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
		} else if result == decision.Stalemate {
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
// évalué (CurrentActor). Elle est convertie en perspective de PreviousActor
// (l'acteur qui a effectué l'action menant à ce nœud), puis propagée vers la
// racine en inversant le signe uniquement quand l'acteur change entre niveaux.
//
// Pour un problème à un seul acteur (CurrentActor == PreviousActor à chaque
// nœud), aucune inversion n'a lieu : la valeur est propagée telle quelle.
// Pour un jeu à deux acteurs en alternance, le comportement est identique
// à l'alternance systématique classique.
func (node *mctsNode) backpropagateValue(value float64) {
	// Convertir de la perspective de CurrentActor vers celle de PreviousActor.
	// Pour un seul acteur, CurrentActor == PreviousActor : pas d'inversion.
	if node.state.CurrentActor() != node.state.PreviousActor() {
		value = -value
	}
	for n := node; n != nil; n = n.parent {
		n.visits++
		n.wins += value
		// Inverser le signe uniquement si l'acteur qui a joué change
		// entre ce nœud et son parent.
		if n.parent != nil && n.state.PreviousActor() != n.parent.state.PreviousActor() {
			value = -value
		}
	}
}
