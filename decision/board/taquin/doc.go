// Package taquin implémente un puzzle à glissement (taquin) compatible avec
// l'interface [decision.State], permettant sa résolution par le moteur MCTS.
//
// C'est un exemple de problème de décision à un seul acteur : un unique
// décideur déplace les tuiles pour atteindre la configuration cible.
//
// La grille est configurable (rows x cols). Par défaut, un 5-puzzle (3x2) :
//
//	┌───┬───┬───┐
//	│ 1 │ 2 │ 3 │
//	├───┼───┼───┤
//	│ 4 │ 5 │   │
//	└───┴───┴───┘
//
// L'état cible est la configuration ordonnée [1, 2, ..., N, 0] où 0
// représente la case vide. Le mélange se fait par mouvements aléatoires
// depuis l'état résolu, ce qui garantit la solvabilité.
//
// # Compatibilité MCTS
//
// Les deux modes du moteur MCTS fonctionnent correctement :
//   - MCTS pur ([mcts.NewMCTS]) : backpropagate crédite les victoires quand
//     result == PreviousActor(), ce qui est toujours vrai pour un seul acteur.
//   - AlphaZero ([mcts.NewAlphaMCTS]) : backpropagateValue reçoit une map de
//     valeurs par acteur, sans hypothèse de somme nulle.
//
// Une limite d'étapes (maxSteps) borne les rollouts pour éviter les boucles
// infinies. Quand la limite est atteinte, [Evaluate] retourne [decision.Stalemate].
package taquin
