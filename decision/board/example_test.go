package board_test

import (
	"fmt"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/board"
	"github.com/owulveryck/alphazego/mcts"
)

// Constantes locales pour les acteurs de cet exemple.
const (
	actor1 decision.ActorID = 1
	actor2 decision.ActorID = 2
)

// ttt est une implémentation minimale de [decision.State] et
// [board.ActionRecorder] pour le morpion.
// Elle illustre les champs nécessaires : le plateau, l'acteur
// courant, et l'action qui a produit cet état.
//
// Le plateau utilise un tableau fixe [9]uint8 (pas un slice) :
// l'affectation d'un tableau copie les données automatiquement,
// ce qui simplifie [decision.State.PossibleMoves].
type ttt struct {
	cells      [9]uint8         // 0=vide, 1=acteur1, 2=acteur2
	turn       decision.ActorID // acteur dont c'est le tour
	lastAction int              // action qui a produit cet état
}

func (t *ttt) CurrentActor() decision.ActorID  { return t.turn }
func (t *ttt) PreviousActor() decision.ActorID { return 3 - t.turn }
func (t *ttt) LastAction() int                 { return t.lastAction }

// ID inclut l'acteur courant : même plateau + acteur différent = ID différent.
func (t *ttt) ID() string {
	var id [10]byte
	copy(id[:], t.cells[:])
	id[9] = byte(t.turn)
	return string(id[:])
}

// Evaluate vérifie les 8 combinaisons gagnantes, puis le match nul.
func (t *ttt) Evaluate() decision.ActorID {
	for _, l := range [][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // lignes
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // colonnes
		{0, 4, 8}, {2, 4, 6}, // diagonales
	} {
		if t.cells[l[0]] != 0 &&
			t.cells[l[0]] == t.cells[l[1]] &&
			t.cells[l[1]] == t.cells[l[2]] {
			return decision.ActorID(t.cells[l[0]])
		}
	}
	for _, c := range t.cells {
		if c == 0 {
			return decision.Undecided
		}
	}
	return decision.Stalemate
}

// PossibleMoves retourne un état par case vide.
// Chaque enfant est une copie indépendante grâce au tableau [9]uint8 :
// l'affectation child.cells = t.cells copie les 9 octets.
func (t *ttt) PossibleMoves() []decision.State {
	if t.Evaluate() != decision.Undecided {
		return nil // état terminal : aucun coup
	}
	var moves []decision.State
	for i, c := range t.cells {
		if c == 0 {
			child := &ttt{
				cells:      t.cells,    // copie automatique (tableau, pas slice)
				turn:       3 - t.turn, // acteur suivant
				lastAction: i,          // action qui a produit cet état
			}
			child.cells[i] = uint8(t.turn) // placer le pion de l'acteur courant
			moves = append(moves, child)
		}
	}
	return moves
}

// Cet exemple montre une implémentation complète de [decision.State] pour le
// morpion (tic-tac-toe), connectée au moteur MCTS.
//
// L'implémentation tient en ~50 lignes grâce à deux choix :
//   - un tableau fixe [9]uint8 pour le plateau (copie automatique)
//   - l'acteur suivant calculé par 3 - acteur (alternance 1↔2)
//
// Voir le type ttt dans le code source pour l'implémentation complète.
func Example() {
	game := &ttt{turn: actor1}

	// Connecter au MCTS : une seule ligne
	m := mcts.NewMCTS()
	bestState := m.RunMCTS(game, 1000)

	// ActionRecorder permet d'extraire l'action choisie par le MCTS
	move := bestState.(board.ActionRecorder).LastAction()
	fmt.Println("MCTS chose a valid position:", move >= 0 && move <= 8)
	fmt.Println("Next actor after MCTS move:", bestState.CurrentActor())
	// Output:
	// MCTS chose a valid position: true
	// Next actor after MCTS move: 2
}

// Cet exemple montre comment utiliser [board.ActionRecorder] pour extraire
// l'action choisie par le MCTS. L'interface [decision.State] est générique et
// n'impose pas de notion d'action ; [board.ActionRecorder] est une commodité
// pour les jeux de plateau où l'on veut connaître le coup joué.
func Example_actionRecorder() {
	game := &ttt{turn: actor1}
	m := mcts.NewMCTS()

	bestState := m.RunMCTS(game, 1000)

	// Le MCTS retourne un decision.State. Pour obtenir l'action,
	// on utilise une assertion de type vers ActionRecorder.
	if ar, ok := bestState.(board.ActionRecorder); ok {
		action := ar.LastAction()
		fmt.Println("Action extraite:", action >= 0 && action <= 8)
	}
	// Output:
	// Action extraite: true
}

// Cet exemple montre une partie complète MCTS vs MCTS.
func Example_fullGame() {
	game := &ttt{turn: actor1}
	m := mcts.NewMCTS()

	for game.Evaluate() == decision.Undecided {
		bestState := m.RunMCTS(game, 500)
		move := bestState.(board.ActionRecorder).LastAction()
		// Appliquer le coup
		game.cells[move] = uint8(game.turn)
		game.turn = 3 - game.turn
		game.lastAction = move
	}

	result := game.Evaluate()
	fmt.Println("Game finished:", result != decision.Undecided)
	// Output:
	// Game finished: true
}
