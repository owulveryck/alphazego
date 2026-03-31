package mcts

import "math"

// puct calcule le score Polynomial Upper Confidence Trees pour ce noeud.
// Cette formule est utilisee dans AlphaZero pour la selection, en remplacement
// de ucb1. Elle integre la probabilite a priori P(s,a) fournie par le
// policy network pour guider l'exploration vers les coups les plus prometteurs.
//
// Formule : Q(s,a) + C_puct * P(s,a) * sqrt(N(parent)) / (1 + N(s,a))
//
// Contrairement a UCB1, les noeuds non visites recoivent un score fini
// proportionnel a leur prior, permettant un elagage implicite des coups
// juges peu prometteurs par le reseau.
func (n *mctsNode) puct() float64 {
	if n.visits == 0 {
		if n.parent == nil {
			return n.prior
		}
		return n.mcts.cpuct * n.prior * math.Sqrt(n.parent.visits)
	}

	q := n.wins / n.visits
	if n.parent == nil {
		return q
	}

	exploration := n.mcts.cpuct * n.prior * math.Sqrt(n.parent.visits) / (1 + n.visits)
	return q + exploration
}
