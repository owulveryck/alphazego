package mcts

import "github.com/owulveryck/alphazego/decision"

// Evaluator fournit une évaluation d'une position de jeu.
// Il est utilisé par le MCTS pour remplacer les rollouts aléatoires (value)
// et guider l'exploration (policy).
//
// Pour un MCTS pur, un évaluateur effectue des rollouts aléatoires avec
// une policy uniforme. Pour AlphaZero, un réseau de neurones fournit
// les deux en un seul appel.
type Evaluator interface {
	// Evaluate prend un état et retourne :
	//   - policy : probabilité a priori pour chaque action légale,
	//     dans le même ordre que [decision.State.PossibleMoves].
	//     La somme des éléments doit être égale à 1.
	//   - value : estimation de victoire pour l'acteur courant, dans [-1, 1].
	//     1 signifie victoire certaine de l'acteur courant,
	//     -1 signifie défaite certaine, 0 signifie match nul.
	Evaluate(state decision.State) (policy []float64, value float64)
}
