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
		n.visits++

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

// backpropagateValue propage des valeurs continues ∈ [-1, 1] depuis ce nœud
// jusqu'à la racine. Cette méthode est utilisée par le chemin AlphaZero, où les
// valeurs proviennent du réseau de neurones au lieu d'un rollout aléatoire.
//
// La map values associe chaque [decision.ActorID] à sa valeur. À chaque nœud,
// la valeur de PreviousActor() (l'acteur qui a effectué l'action menant à ce
// nœud) est ajoutée aux wins. Cette logique est identique à [backpropagate]
// mais avec des valeurs continues.
//
// Cette approche fonctionne pour tout nombre d'acteurs (1, 2, N) sans
// hypothèse de somme nulle.
func (node *mctsNode) backpropagateValue(values map[decision.ActorID]float64) {
	for n := node; n != nil; n = n.parent {
		n.visits++
		n.wins += values[n.state.PreviousActor()]
	}
}

// backpropagateTerminal propage le résultat d'un état terminal depuis ce nœud
// jusqu'à la racine. Chaque nœud reçoit 1.0 si PreviousActor a gagné, -1.0
// s'il a perdu, ou 0.0 en cas de match nul (convention [-1, 1]).
//
// Cette méthode remplace l'ancien terminalValue + backpropagateValue pour les
// nœuds terminaux, en calculant la valeur à la volée pour chaque acteur.
func (node *mctsNode) backpropagateTerminal() {
	result := node.state.Evaluate()
	for n := node; n != nil; n = n.parent {
		n.visits++
		actor := n.state.PreviousActor()
		if result == actor {
			n.wins += 1.0
		} else if result != decision.Stalemate {
			n.wins += -1.0
		}
		// Stalemate : wins += 0.0 (implicite)
	}
}
