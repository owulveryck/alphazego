package taquin

import (
	"fmt"
	"math/rand"

	"github.com/owulveryck/alphazego/decision"
)

// Player est l'unique acteur du taquin. Dans un problème à un seul acteur,
// CurrentActor() et PreviousActor() retournent toujours cette valeur.
const Player decision.ActorID = 1

// Directions de déplacement de la case vide.
// Ce sont les valeurs retournées par [LastAction].
const (
	Up    = 0
	Down  = 1
	Left  = 2
	Right = 3
)

// dirNames associe chaque direction à son nom pour l'affichage.
var dirNames = [4]string{"Up", "Down", "Left", "Right"}

// MaxBoardSize est la taille maximale du plateau (5x5 = 25 cases).
const MaxBoardSize = 25

// Taquin représente l'état d'un puzzle à glissement.
// Il implémente [decision.State] et [board.ActionRecorder].
type Taquin struct {
	board    [MaxBoardSize]uint8 // valeurs des tuiles (0 = case vide)
	rows     int
	cols     int
	blank    int // position de la case vide
	steps    int // nombre de coups joués depuis l'état initial
	maxSteps int // limite pour borner les rollouts MCTS
	lastDir  int // direction du dernier mouvement
}

// NewTaquin crée un taquin résolu de taille rows x cols avec une limite
// de maxSteps coups. La grille est en configuration cible : [1, 2, ..., N, 0].
//
// Panics si rows*cols > [MaxBoardSize] ou si rows < 2 ou cols < 2.
func NewTaquin(rows, cols, maxSteps int) *Taquin {
	size := rows * cols
	if size > MaxBoardSize {
		panic(fmt.Sprintf("taquin: board size %d exceeds MaxBoardSize %d", size, MaxBoardSize))
	}
	if rows < 2 || cols < 2 {
		panic("taquin: rows and cols must be at least 2")
	}
	t := &Taquin{
		rows:     rows,
		cols:     cols,
		blank:    size - 1,
		maxSteps: maxSteps,
	}
	for i := 0; i < size-1; i++ {
		t.board[i] = uint8(i + 1)
	}
	// board[size-1] = 0 (case vide, déjà la valeur zéro)
	return t
}

// Shuffle mélange le taquin en effectuant n mouvements aléatoires depuis
// l'état résolu. Cela garantit que le puzzle est toujours solvable.
// Le compteur de steps est remis à zéro après le mélange.
func (t *Taquin) Shuffle(n int, rng *rand.Rand) {
	for i := 0; i < n; i++ {
		dirs := t.validDirections()
		dir := dirs[rng.Intn(len(dirs))]
		t.move(dir)
	}
	t.steps = 0
	t.lastDir = 0
}

// Play effectue un mouvement dans la direction dir (Up, Down, Left, Right).
// Retourne une erreur si la direction est invalide ou si le puzzle est terminé.
func (t *Taquin) Play(dir int) error {
	if dir < Up || dir > Right {
		panic(fmt.Sprintf("taquin: direction %d invalide (0-3)", dir))
	}
	if t.Evaluate() != decision.Undecided {
		return fmt.Errorf("le puzzle est terminé")
	}
	if !t.canMove(dir) {
		return fmt.Errorf("mouvement %s impossible depuis la position %d", dirNames[dir], t.blank)
	}
	t.move(dir)
	return nil
}

// move déplace la case vide dans la direction dir sans vérification.
func (t *Taquin) move(dir int) {
	target := t.target(dir)
	t.board[t.blank], t.board[target] = t.board[target], t.board[t.blank]
	t.blank = target
	t.steps++
	t.lastDir = dir
}

// target retourne la position cible pour un déplacement dans la direction dir.
func (t *Taquin) target(dir int) int {
	switch dir {
	case Up:
		return t.blank - t.cols
	case Down:
		return t.blank + t.cols
	case Left:
		return t.blank - 1
	case Right:
		return t.blank + 1
	}
	panic("unreachable")
}

// canMove vérifie si la case vide peut se déplacer dans la direction dir.
func (t *Taquin) canMove(dir int) bool {
	row, col := t.blank/t.cols, t.blank%t.cols
	switch dir {
	case Up:
		return row > 0
	case Down:
		return row < t.rows-1
	case Left:
		return col > 0
	case Right:
		return col < t.cols-1
	}
	return false
}

// validDirections retourne les directions valides depuis la position courante
// de la case vide.
func (t *Taquin) validDirections() []int {
	dirs := make([]int, 0, 4)
	for d := Up; d <= Right; d++ {
		if t.canMove(d) {
			dirs = append(dirs, d)
		}
	}
	return dirs
}

// CurrentActor retourne [Player]. Dans un problème à un seul acteur,
// c'est toujours le même décideur.
func (t *Taquin) CurrentActor() decision.ActorID { return Player }

// PreviousActor retourne [Player]. Dans un problème à un seul acteur,
// l'acteur précédent est le même que l'acteur courant.
func (t *Taquin) PreviousActor() decision.ActorID { return Player }

// LastAction retourne la direction du dernier mouvement effectué (Up, Down,
// Left, Right). Pour l'état initial, la valeur n'est pas significative.
// Implémente [board.ActionRecorder].
func (t *Taquin) LastAction() int { return t.lastDir }

// Evaluate retourne l'issue du puzzle :
//   - [Player] si le puzzle est résolu (configuration cible atteinte)
//   - [decision.Stalemate] si la limite de steps est atteinte sans solution
//   - [decision.Undecided] sinon
func (t *Taquin) Evaluate() decision.ActorID {
	if t.isSolved() {
		return Player
	}
	if t.steps >= t.maxSteps {
		return decision.Stalemate
	}
	return decision.Undecided
}

// isSolved vérifie si le plateau est en configuration cible : [1, 2, ..., N, 0].
func (t *Taquin) isSolved() bool {
	size := t.rows * t.cols
	for i := 0; i < size-1; i++ {
		if t.board[i] != uint8(i+1) {
			return false
		}
	}
	return t.board[size-1] == 0
}

// ID retourne un identifiant unique pour cet état. L'ID encode le plateau
// et le nombre de steps pour que l'inventaire MCTS distingue le même plateau
// atteint à des profondeurs différentes.
func (t *Taquin) ID() string {
	size := t.rows * t.cols
	// board + 2 bytes pour steps (little-endian, supporte jusqu'à 65535)
	id := make([]byte, size+2)
	copy(id, t.board[:size])
	id[size] = byte(t.steps)
	id[size+1] = byte(t.steps >> 8)
	return string(id)
}

// PossibleMoves retourne tous les états atteignables en un mouvement.
// Chaque état enfant est une copie indépendante grâce au tableau fixe
// [MaxBoardSize]uint8.
func (t *Taquin) PossibleMoves() []decision.State {
	if t.Evaluate() != decision.Undecided {
		return nil
	}
	dirs := t.validDirections()
	moves := make([]decision.State, 0, len(dirs))
	for _, dir := range dirs {
		child := &Taquin{
			board:    t.board, // copie par valeur (tableau fixe)
			rows:     t.rows,
			cols:     t.cols,
			blank:    t.blank,
			steps:    t.steps,
			maxSteps: t.maxSteps,
		}
		child.move(dir)
		child.lastDir = dir
		moves = append(moves, child)
	}
	return moves
}

// Steps retourne le nombre de coups joués depuis l'état initial.
func (t *Taquin) Steps() int { return t.steps }

// MaxSteps retourne la limite de coups configurée.
func (t *Taquin) MaxSteps() int { return t.maxSteps }

// Rows retourne le nombre de lignes de la grille.
func (t *Taquin) Rows() int { return t.rows }

// Cols retourne le nombre de colonnes de la grille.
func (t *Taquin) Cols() int { return t.cols }

// Features retourne l'état du taquin sous forme de tenseur aplati [1 * rows * cols].
// Chaque case contient la valeur de la tuile normalisée par le nombre total de tuiles.
// La case vide est représentée par 0.
func (t *Taquin) Features() []float32 {
	size := t.rows * t.cols
	features := make([]float32, size)
	norm := float32(size)
	for i := 0; i < size; i++ {
		features[i] = float32(t.board[i]) / norm
	}
	return features
}

// FeatureShape retourne les dimensions du tenseur : 1 canal, rows x cols.
func (t *Taquin) FeatureShape() [3]int {
	return [3]int{1, t.rows, t.cols}
}

// ActionSize retourne le nombre total de directions possibles (4 : Up, Down, Left, Right).
func (t *Taquin) ActionSize() int {
	return 4
}
