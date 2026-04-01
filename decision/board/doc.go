// Package board définit les abstractions spécifiques aux jeux de plateau
// et autres problèmes discrets où l'on souhaite enregistrer les actions
// et convertir l'état en tenseur.
//
// L'interface [Boarder] combine [decision.State] et [ActionRecorder] pour
// représenter un état de jeu de plateau complet.
//
// [Tensorizable] permet la conversion en tenseur pour l'évaluation par un
// réseau de neurones.
//
// Deux implémentations de référence sont fournies :
//   - le morpion ([tictactoe]) : problème à deux acteurs
//   - le taquin ([taquin]) : problème à un seul acteur
package board
