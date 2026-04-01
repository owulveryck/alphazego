package board

import "github.com/owulveryck/alphazego/decision"

// Boarder représente un état de jeu de plateau : un problème de décision
// séquentiel ([decision.State]) enrichi d'un enregistreur d'action
// ([ActionRecorder]).
//
// Les implémentations de Boarder peuvent être utilisées directement avec
// le moteur MCTS via leur interface [decision.State], puis interrogées
// via [ActionRecorder] pour extraire l'action choisie :
//
//	bestState := m.RunMCTS(currentState, 1000)
//	move := bestState.(board.Boarder).LastAction()
type Boarder interface {
	decision.State
	ActionRecorder
}

// ActionRecorder est une interface complémentaire à [decision.State] pour les
// problèmes où il est utile de connaître l'action qui a produit un état donné.
//
// C'est une commodité pour les jeux de plateau et autres problèmes discrets.
// Le moteur MCTS n'utilise pas cette interface : il travaille uniquement avec
// [decision.State]. C'est l'appelant (boucle de jeu, UI, enregistrement de
// parties) qui l'utilise pour extraire l'action choisie par le MCTS :
//
//	bestState := m.RunMCTS(currentState, 1000)
//	move := bestState.(board.ActionRecorder).LastAction()
//
// Le type de retour est int (pas uint8) pour supporter des espaces d'action
// supérieurs à 256 (Go 19x19 : 362 actions, échecs : ~4672 actions possibles).
type ActionRecorder interface {
	// LastAction retourne l'action qui a été effectuée pour atteindre cet état.
	// Chaque état retourné par [decision.State.PossibleMoves] doit avoir ce
	// champ correctement initialisé.
	// Pour l'état initial (aucune action), la valeur n'est pas significative.
	LastAction() int
}

// Tensorizable est implémenté par les états de jeu qui peuvent être
// convertis en tenseur pour l'évaluation par un réseau de neurones.
type Tensorizable interface {
	// Features retourne l'état du jeu sous forme de tenseur aplati.
	// Le format attendu est [C * H * W] en row-major order,
	// où C est le nombre de canaux (plans de features),
	// H la hauteur et W la largeur du plateau.
	Features() []float32

	// FeatureShape retourne les dimensions du tenseur [C, H, W].
	FeatureShape() [3]int

	// ActionSize retourne le nombre total d'actions possibles dans le jeu
	// (pas seulement les actions légales dans l'état courant).
	// Morpion : 9, Go 19x19 : 362 (361 + passe).
	ActionSize() int
}
