package mcts

import "github.com/owulveryck/alphazego/decision"

// Evaluator fournit une évaluation d'une position de jeu.
// Il est utilisé par le MCTS pour remplacer les rollouts aléatoires (values)
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
	//   - values : estimation de victoire pour chaque acteur, dans [-1, 1].
	//     La clé est l'[decision.ActorID], la valeur est l'estimation :
	//     1 signifie victoire certaine, -1 défaite certaine, 0 match nul.
	//     La map doit contenir au minimum les ActorIDs présents dans le
	//     problème (CurrentActor, PreviousActor, et tout autre acteur).
	//
	//   Exemples :
	//     - 2 joueurs somme nulle : {Cross: 0.8, Circle: -0.8}
	//     - 1 acteur :              {Player: 0.7}
	//     - 3 acteurs :             {A: 0.6, B: -0.3, C: -0.3}
	Evaluate(state decision.State) (policy []float64, values map[decision.ActorID]float64)
}
