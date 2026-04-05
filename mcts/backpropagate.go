package mcts

import (
	"math"

	"github.com/owulveryck/alphazego/decision"
)

// backpropagate updates the statistics for this node and its ancestors up to the root node
// after a game simulation is completed. The statistics updated include the number of visits
// and wins, which are used to calculate the node's value in future selections.
//
// Convention de récompense [0, 0.5, 1] (MCTS pur) :
//   - 1.0 si previousActor a gagné
//   - 0.5 en cas de match nul (Stalemate)
//   - 0.0 sinon (perte ou non-participation)
//
// Cette convention [0, 1] est standard pour UCB1, où avgReward = wins/visits
// reste dans [0, 1]. Elle diffère de [backpropagateTerminal] qui utilise [-1, 1]
// pour le chemin AlphaZero.
//
// previousActor est lu depuis le champ caché du nœud (calculé à la création)
// pour éviter un appel d'interface à chaque étape de la remontée.
// logVisits est mis à jour incrémentalement pour éviter un recalcul dans selectChildUCB.
func (node *mctsNode) backpropagate(result decision.ActorID) {
	for n := node; n != nil; n = n.parent {
		n.visits++
		n.logVisits = math.Log(n.visits)

		if result == n.previousActor {
			n.wins += 1
		} else if result == decision.Stalemate {
			n.wins += 0.5
		}
	}
}

// backpropagateValue propage des valeurs continues ∈ [-1, 1] depuis ce nœud
// jusqu'à la racine. Cette méthode est utilisée par le chemin AlphaZero, où les
// valeurs proviennent du réseau de neurones au lieu d'un rollout aléatoire.
//
// La map values associe chaque [decision.ActorID] à sa valeur. À chaque nœud,
// la valeur de previousActor (l'acteur qui a effectué l'action menant à ce
// nœud) est ajoutée aux wins.
func (node *mctsNode) backpropagateValue(values map[decision.ActorID]float64) {
	for n := node; n != nil; n = n.parent {
		n.visits++
		n.logVisits = math.Log(n.visits)
		n.wins += values[n.previousActor]
	}
}

// backpropagateTerminal propage le résultat d'un état terminal depuis ce nœud
// jusqu'à la racine. Chaque nœud reçoit 1.0 si PreviousActor a gagné, -1.0
// s'il a perdu, ou 0.0 en cas de match nul (convention [-1, 1]).
//
// Cette convention [-1, 1] est celle du chemin AlphaZero, cohérente avec les
// valeurs retournées par [Evaluator.Evaluate]. Elle diffère de [backpropagate]
// qui utilise [0, 0.5, 1] pour le MCTS pur avec UCB1.
func (node *mctsNode) backpropagateTerminal() {
	result := node.state.Evaluate()
	for n := node; n != nil; n = n.parent {
		n.visits++
		n.logVisits = math.Log(n.visits)
		if result == n.previousActor {
			n.wins += 1.0
		} else if result != decision.Stalemate {
			n.wins += -1.0
		}
	}
}
