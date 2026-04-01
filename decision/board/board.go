package board

import "github.com/owulveryck/alphazego/decision"

// Boarder represente un etat de jeu de plateau : un probleme de decision
// sequentiel ([decision.State]) enrichi d'un enregistreur d'action
// ([ActionRecorder]).
//
// Les implementations de Boarder peuvent etre utilisees directement avec
// le moteur MCTS via leur interface [decision.State], puis interrogees
// via [ActionRecorder] pour extraire l'action choisie :
//
//	bestState := m.RunMCTS(currentState, 1000)
//	move := bestState.(board.Boarder).LastAction()
type Boarder interface {
	decision.State
	ActionRecorder
}

// ActionRecorder est une interface complementaire a [decision.State] pour les
// problemes ou il est utile de connaitre l'action qui a produit un etat donne.
//
// C'est une commodite pour les jeux de plateau et autres problemes discrets.
// Le moteur MCTS n'utilise pas cette interface : il travaille uniquement avec
// [decision.State]. C'est l'appelant (boucle de jeu, UI, enregistrement de
// parties) qui l'utilise pour extraire l'action choisie par le MCTS :
//
//	bestState := m.RunMCTS(currentState, 1000)
//	move := bestState.(board.ActionRecorder).LastAction()
//
// Le type de retour est int (pas uint8) pour supporter des espaces d'action
// superieurs a 256 (Go 19x19 : 362 actions, echecs : ~4672 actions possibles).
type ActionRecorder interface {
	// LastAction retourne l'action qui a ete effectuee pour atteindre cet etat.
	// Chaque etat retourne par [decision.State.PossibleMoves] doit avoir ce
	// champ correctement initialise.
	// Pour l'etat initial (aucune action), la valeur n'est pas significative.
	LastAction() int
}

// Tensorizable est implemente par les etats de jeu qui peuvent etre
// convertis en tenseur pour l'evaluation par un reseau de neurones.
type Tensorizable interface {
	// Features retourne l'etat du jeu sous forme de tenseur aplati.
	// Le format attendu est [C * H * W] en row-major order,
	// ou C est le nombre de canaux (plans de features),
	// H la hauteur et W la largeur du plateau.
	Features() []float32

	// FeatureShape retourne les dimensions du tenseur [C, H, W].
	FeatureShape() [3]int

	// ActionSize retourne le nombre total d'actions possibles dans le jeu
	// (pas seulement les actions legales dans l'etat courant).
	// Morpion : 9, Go 19x19 : 362 (361 + passe).
	ActionSize() int
}
