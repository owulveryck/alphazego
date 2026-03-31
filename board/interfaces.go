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

// Move represente une action ou une position. Dans un jeu de plateau,
// c'est une case. Dans un probleme generique, c'est un identifiant d'action.
type Move = uint8

// ID est un identifiant unique pour un etat. Deux etats identiques
// (meme configuration, meme agent courant) doivent produire le meme ID.
type ID = []byte

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
// Un State doit etre immuable du point de vue du MCTS : les methodes comme
// [State.PossibleMoves] doivent retourner de nouveaux State sans modifier
// le recepteur.
//
// # Mapping vers d'autres domaines
//
//   - Jeu de plateau : CurrentPlayer = joueur, PossibleMoves = coups legaux
//   - Negociation : CurrentPlayer = partie, PossibleMoves = propositions/contre-offres
//   - Diagnostic : CurrentPlayer = optique (patient/medecin), PossibleMoves = examens/traitements
//   - Planification : CurrentPlayer = agent, PossibleMoves = actions possibles
type State interface {
	// CurrentPlayer retourne le joueur dont c'est le tour d'agir.
	CurrentPlayer() PlayerID
	// PreviousPlayer retourne le joueur qui a effectue le coup menant a cet etat.
	// Pour l'etat initial (aucun coup joue), le comportement est defini par
	// l'implementation (typiquement [NoPlayer] ou le dernier joueur dans l'ordre de jeu).
	// Cette methode permet au moteur MCTS de determiner "qui a joue ici" sans
	// connaitre la logique de tour (alternance, round-robin, etc.).
	PreviousPlayer() PlayerID
	// Evaluate retourne l'issue du probleme :
	//   - [NoPlayer] (0) si le probleme est en cours
	//   - [DrawResult] (-1) en cas de match nul
	//   - un PlayerID positif si ce joueur a gagne
	Evaluate() PlayerID
	// PossibleMoves retourne tous les etats atteignables depuis l'etat courant
	// en effectuant une action. Chaque State retourne au prochain agent comme
	// CurrentPlayer.
	PossibleMoves() []State
	// ID retourne un identifiant unique pour cet etat. Deux etats avec la meme
	// configuration et le meme agent courant doivent retourner des ID identiques.
	ID() ID
}

// Playable is implemented by game states that can extract the move
// that was played between two consecutive states.
type Playable interface {
	// GetMoveFromState compares the receiver with another State and returns
	// the Move (board position) that differs between the two.
	GetMoveFromState(State) Move
}

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
	//     dans le meme ordre que [State.PossibleMoves].
	//     La somme des elements doit etre egale a 1.
	//   - value : estimation de victoire pour le joueur courant, dans [-1, 1].
	//     1 signifie victoire certaine du joueur courant,
	//     -1 signifie defaite certaine, 0 signifie match nul.
	Evaluate(state State) (policy []float64, value float64)
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
