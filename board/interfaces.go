// Package board defines the interfaces and types for sequential decision
// problems solvable by the MCTS (Monte Carlo Tree Search) engine.
//
// The core abstraction is [State], qui represente un instantane d'un
// probleme de decision sequentiel a un ou plusieurs agents. Les jeux de
// plateau (morpion, Go, echecs) en sont l'exemple naturel, mais les
// interfaces s'appliquent a tout probleme modelisable comme un arbre de
// decisions : negociations, planification adversariale, diagnostics
// medicaux, composition creative, etc.
//
// Le nombre de joueurs n'est pas fixe par le framework. L'interface [State]
// expose [State.PreviousPlayer] pour que le moteur MCTS determine qui a
// effectue le dernier coup sans connaitre la logique de tour. Ainsi, un jeu
// a 2 joueurs (morpion), a 1 joueur (planification) ou a N joueurs
// (negociation multilaterale) peut etre resolu par le meme moteur.
//
// # Types
//
// [PlayerID] identifie un decideur (joueur, partie, acteur). C'est un type
// distinct base sur int. La methode [State.Evaluate] retourne directement un
// PlayerID : le gagnant si la partie est finie, [NoPlayer] si elle est en
// cours, ou [DrawResult] en cas de match nul. Il n'y a pas de type Result
// separe : le resultat EST l'identifiant du gagnant.
//
// Pour utiliser ce framework, implementer [State] avec la logique
// specifique au probleme. L'implementation [tictactoe.TicTacToe] sert
// d'exemple de reference pour un jeu a deux joueurs.
package board

// PlayerID identifie un decideur dans le probleme. Dans un jeu de plateau,
// c'est un joueur. Dans une negociation, c'est une partie. Dans un probleme
// de planification, c'est un acteur.
//
// PlayerID est aussi utilise comme resultat de [State.Evaluate] : la valeur
// retournee est le PlayerID du gagnant, [NoPlayer] si le jeu est en cours,
// ou [DrawResult] en cas de match nul.
type PlayerID int

const (
	// NoPlayer indique qu'aucun joueur n'est concerne. Utilise comme valeur de
	// retour de [State.Evaluate] pour indiquer que le jeu est en cours, et comme
	// contenu d'une position vide sur un plateau.
	NoPlayer PlayerID = 0
	// DrawResult indique un match nul : la partie est terminee mais aucun
	// joueur n'a gagne.
	DrawResult PlayerID = -1
	// Player1 est le premier joueur (typiquement X au morpion).
	Player1 PlayerID = 1
	// Player2 est le second joueur (typiquement O au morpion).
	Player2 PlayerID = 2
)

// State represente l'etat d'un probleme de decision sequentiel a un ou
// plusieurs agents.
//
// Tout probleme implementant cette interface peut etre resolu par le moteur
// MCTS. Les jeux de plateau en sont un cas particulier : le morpion
// ([tictactoe.TicTacToe]) en est l'implementation de reference pour deux
// joueurs.
//
// # Implementer State
//
// Le struct d'etat doit contenir au minimum trois informations :
//
//   - L'etat du probleme (plateau, configuration, etc.)
//   - Le joueur courant (qui doit agir)
//   - Le dernier coup joue (pour [State.LastMove])
//
// Astuce : utiliser un tableau fixe ([N]uint8) plutot qu'un slice ([]uint8)
// pour le plateau simplifie [State.PossibleMoves] : l'affectation d'un
// tableau copie les donnees automatiquement, sans appel a copy().
//
// # Contrat
//
// Les invariants suivants doivent etre respectes :
//
//   - [State.PossibleMoves] ne doit jamais modifier le recepteur. Chaque etat
//     retourne doit etre une copie independante (c'est le piege le plus courant).
//   - Chaque etat retourne par [State.PossibleMoves] doit avoir
//     [State.CurrentPlayer] defini sur le joueur suivant et
//     [State.LastMove] defini sur l'action qui a produit cet etat.
//   - [State.PossibleMoves] retourne un slice vide (ou nil) quand
//     [State.Evaluate] != [NoPlayer] : un etat terminal n'a pas de coups.
//   - [State.ID] doit etre deterministe et inclure le joueur courant.
//     Meme plateau + joueur different = ID different.
//
// Voir l'exemple du package pour une implementation complete (morpion).
//
// # Mapping vers d'autres domaines
//
//   - Jeu de plateau : CurrentPlayer = joueur, PossibleMoves = coups legaux
//   - Negociation : CurrentPlayer = partie, PossibleMoves = propositions/contre-offres
//   - Diagnostic : CurrentPlayer = optique (patient/medecin), PossibleMoves = examens/traitements
//   - Planification : CurrentPlayer = agent, PossibleMoves = actions possibles
type State interface {
	// CurrentPlayer retourne le joueur dont c'est le tour d'agir.
	// Pour un jeu a deux joueurs en alternance, le joueur suivant est 3 - joueur.
	CurrentPlayer() PlayerID

	// PreviousPlayer retourne le joueur qui a effectue le coup menant a cet etat.
	// Pour un jeu a deux joueurs : 3 - CurrentPlayer().
	// Pour l'etat initial, le comportement est defini par l'implementation
	// (typiquement le dernier joueur dans l'ordre de jeu).
	// Le moteur MCTS utilise cette methode pour crediter les victoires
	// au bon joueur lors de la retropropagation.
	PreviousPlayer() PlayerID

	// Evaluate retourne l'issue du probleme :
	//   - [NoPlayer] (0) si le probleme est en cours
	//   - [DrawResult] (-1) en cas de match nul
	//   - un PlayerID positif si ce joueur a gagne
	//
	// Le moteur MCTS appelle cette methode a chaque noeud pour detecter
	// les etats terminaux. Un etat ou Evaluate != NoPlayer ne doit pas
	// avoir de coups possibles (PossibleMoves retourne nil ou vide).
	Evaluate() PlayerID

	// PossibleMoves retourne tous les etats atteignables depuis l'etat courant
	// en effectuant une action. C'est la methode la plus importante a implementer
	// correctement.
	//
	// Chaque etat retourne doit :
	//   - etre une copie independante (ne pas partager de donnees mutables avec le recepteur)
	//   - avoir CurrentPlayer() defini sur le joueur suivant
	//   - avoir LastMove() defini sur l'action qui a produit cet etat
	//
	// Retourne nil ou un slice vide si Evaluate() != [NoPlayer].
	//
	// Astuce : utiliser un tableau fixe ([N]uint8) au lieu d'un slice ([]uint8)
	// pour le plateau rend la copie automatique par simple affectation.
	PossibleMoves() []State

	// ID retourne un identifiant unique pour cet etat. Deux etats avec la meme
	// configuration et le meme joueur courant doivent retourner des ID identiques.
	// Inclure le joueur courant dans l'ID : meme plateau avec un joueur
	// different constitue un etat different.
	ID() string

	// LastMove retourne le coup (position) qui a ete joue pour atteindre cet etat.
	// Chaque etat retourne par [State.PossibleMoves] doit avoir ce champ
	// correctement initialise.
	// Pour l'etat initial (aucun coup joue), la valeur n'est pas significative.
	LastMove() uint8
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
