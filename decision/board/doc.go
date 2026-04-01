// Package board definit les abstractions specifiques aux jeux de plateau
// et autres problemes discrets ou l'on souhaite enregistrer les actions
// et convertir l'etat en tenseur.
//
// L'interface [Boarder] combine [decision.State] et [ActionRecorder] pour
// representer un etat de jeu de plateau complet.
//
// [Tensorizable] permet la conversion en tenseur pour l'evaluation par un
// reseau de neurones.
//
// L'implementation de reference est le morpion dans le sous-package
// [tictactoe].
package board
