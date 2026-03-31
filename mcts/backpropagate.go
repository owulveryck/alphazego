package mcts

import (
	"github.com/owulveryck/alphazego/board"
)

// backpropagate updates the statistics for this node and its ancestors up to the root node
// after a game simulation is completed. The statistics updated include the number of visits
// and wins, which are used to calculate the node's value in future selections.
func (node *mctsNode) backpropagate(result board.PlayerID) {
	// Starting from the current node, loop through all ancestors until the root node is reached.
	// The loop uses 'n' to traverse the tree upwards, with 'n.parent' moving to each parent node.
	for n := node; n != nil; n = n.parent {
		n.visits += 1

		// Credit wins to the player who made the move leading to this node.
		// PreviousPlayer() retourne le joueur qui a effectue le coup menant a cet etat,
		// quelle que soit la logique de tour (2 joueurs, N joueurs, etc.).
		playerWhoMovedHere := n.state.PreviousPlayer()
		if result == playerWhoMovedHere {
			n.wins += 1
		} else if result == board.DrawResult {
			n.wins += 0.5
		}
	}

	// This method systematically updates the visit and win counts for each node from the
	// current node back to the root. These updated statistics influence the selection of
	// nodes in future iterations of the MCTS, guiding the search towards more promising paths.
}

// backpropagateValue propage une valeur continue v ∈ [-1, 1] depuis ce noeud
// jusqu'a la racine. Cette methode est utilisee par le chemin AlphaZero, ou la
// value provient du reseau de neurones au lieu d'un rollout aleatoire.
//
// La valeur initiale est exprimee du point de vue du joueur courant au noeud
// evalue (CurrentPlayer). Elle est d'abord inversee pour etre stockee du point
// de vue du joueur qui a joue le coup menant a ce noeud (convention coherente
// avec backpropagate), puis alternee a chaque niveau en remontant.
//
// L'inversion de signe suppose un jeu a somme nulle a deux joueurs. Pour les
// jeux a N joueurs (N > 2), utiliser backpropagate avec un [board.PlayerID]
// discret.
func (node *mctsNode) backpropagateValue(value float64) {
	// Inverser pour passer de la perspective du joueur courant a celle du joueur
	// qui a effectue le coup menant a ce noeud (= PreviousPlayer).
	value = -value
	for n := node; n != nil; n = n.parent {
		n.visits++
		n.wins += value
		value = -value
	}
}
