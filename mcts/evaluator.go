package mcts

import "github.com/owulveryck/alphazego/board"

// Evaluator fournit une evaluation d'une position de jeu.
// Il est utilise par le MCTS pour remplacer les rollouts aleatoires (value)
// et guider l'exploration (policy).
//
// Pour un MCTS pur, un evaluateur effectue des rollouts aleatoires avec
// une policy uniforme. Pour AlphaZero, un reseau de neurones fournit
// les deux en un seul appel.
type Evaluator interface {
	// Evaluate prend un etat de jeu et retourne :
	//   - policy : probabilite a priori pour chaque coup legal,
	//     dans le meme ordre que [board.State.PossibleMoves].
	//     La somme des elements doit etre egale a 1.
	//   - value : estimation de victoire pour le joueur courant, dans [-1, 1].
	//     1 signifie victoire certaine du joueur courant,
	//     -1 signifie defaite certaine, 0 signifie match nul.
	Evaluate(state board.State) (policy []float64, value float64)
}
