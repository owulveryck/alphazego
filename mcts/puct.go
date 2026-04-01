package mcts

import "math"

// puct calcule le score Polynomial Upper Confidence Trees pour ce nœud.
// Cette formule est utilisée dans AlphaZero pour la sélection, en remplacement
// de ucb1. Elle intègre la probabilité a priori P(s,a) fournie par le
// policy network pour guider l'exploration vers les coups les plus prometteurs.
//
// Formule : Q(s,a) + C_puct * P(s,a) * sqrt(N(parent)) / (1 + N(s,a))
//
// Contrairement à UCB1, les nœuds non visités reçoivent un score fini
// proportionnel à leur prior, permettant un élagage implicite des coups
// jugés peu prometteurs par le réseau.
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
