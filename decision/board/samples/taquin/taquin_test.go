package taquin

import (
	"math/rand"
	"testing"

	"github.com/owulveryck/alphazego/decision"
)

func TestNewTaquin_Solved(t *testing.T) {
	taq := NewTaquin(3, 2, 50)
	if taq.Evaluate() != Player {
		t.Fatalf("new taquin should be solved, got %v", taq.Evaluate())
	}
	if taq.rows != 3 || taq.cols != 2 {
		t.Fatalf("expected 3x2, got %dx%d", taq.rows, taq.cols)
	}
	if taq.blank != 5 {
		t.Fatalf("expected blank at 5, got %d", taq.blank)
	}
}

func TestNewTaquin_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for oversized board")
		}
	}()
	NewTaquin(6, 6, 50) // 36 > MaxBoardSize
}

func TestNewTaquin_PanicsTooSmall(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for 1x3 board")
		}
	}()
	NewTaquin(1, 3, 50)
}

func TestShuffle(t *testing.T) {
	taq := NewTaquin(2, 3, 100)
	rng := rand.New(rand.NewSource(42))
	taq.Shuffle(20, rng)

	if taq.steps != 0 {
		t.Fatalf("steps should be 0 after shuffle, got %d", taq.steps)
	}
	if taq.Evaluate() == Player {
		t.Fatal("shuffled taquin should not be solved (very unlikely)")
	}
}

func TestCurrentActor_AlwaysPlayer(t *testing.T) {
	taq := NewTaquin(2, 2, 10)
	if taq.CurrentActor() != Player {
		t.Fatal("expected Player")
	}
	if taq.PreviousActor() != Player {
		t.Fatal("expected Player")
	}
}

func TestEvaluate_Stalemate(t *testing.T) {
	taq := NewTaquin(2, 2, 0) // maxSteps=0, déjà résolu → Player
	if taq.Evaluate() != Player {
		t.Fatalf("solved puzzle should return Player even at maxSteps=0")
	}

	// Mélanger puis vérifier stalemate
	taq2 := NewTaquin(2, 2, 2)
	rng := rand.New(rand.NewSource(1))
	taq2.Shuffle(5, rng)
	// Jouer jusqu'à la limite
	for taq2.Evaluate() == decision.Undecided {
		dirs := taq2.validDirections()
		taq2.move(dirs[0])
	}
	result := taq2.Evaluate()
	if result != Player && result != decision.Stalemate {
		t.Fatalf("expected Player or Stalemate, got %v", result)
	}
}

func TestPlay_Valid(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	taq.Shuffle(5, rng)
	dirs := taq.validDirections()
	err := taq.Play(dirs[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if taq.steps != 1 {
		t.Fatalf("expected 1 step, got %d", taq.steps)
	}
	if taq.lastDir != dirs[0] {
		t.Fatalf("expected lastDir=%d, got %d", dirs[0], taq.lastDir)
	}
}

func TestPlay_InvalidDirection(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	err := taq.Play(5)
	if err == nil {
		t.Fatal("expected error for invalid direction")
	}
}

func TestPlay_ImpossibleMove(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	// Case vide en position 5 (row=1, col=2), ne peut pas aller Down ni Right
	err := taq.Play(Down)
	if err == nil {
		t.Fatal("expected error for impossible move")
	}
}

func TestPlay_GameOver(t *testing.T) {
	// Créer un puzzle terminé par stalemate
	taq := NewTaquin(2, 2, 1)
	rng := rand.New(rand.NewSource(1))
	taq.Shuffle(3, rng)
	taq.move(taq.validDirections()[0]) // 1 step → stalemate si pas résolu
	if taq.Evaluate() == decision.Undecided {
		t.Skip("puzzle happened to be solved in 1 move")
	}
	err := taq.Play(Up)
	if err == nil {
		t.Fatal("expected error on terminated puzzle")
	}
}

func TestPossibleMoves_Solved(t *testing.T) {
	taq := NewTaquin(3, 3, 50)
	// Puzzle résolu → état terminal → pas de mouvements
	moves := taq.PossibleMoves()
	if len(moves) != 0 {
		t.Fatalf("solved puzzle should have 0 moves, got %d", len(moves))
	}
}

func TestPossibleMoves_Count(t *testing.T) {
	taq := NewTaquin(3, 3, 50)
	rng := rand.New(rand.NewSource(42))
	taq.Shuffle(50, rng)
	if taq.isSolved() {
		t.Skip("shuffle returned to solved state (extremely unlikely)")
	}
	moves := taq.PossibleMoves()
	dirs := taq.validDirections()
	if len(moves) != len(dirs) {
		t.Fatalf("expected %d moves, got %d", len(dirs), len(moves))
	}
}

func TestPossibleMoves_Independence(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	taq.Shuffle(10, rng)

	moves := taq.PossibleMoves()
	if len(moves) < 2 {
		t.Skip("not enough moves to test independence")
	}

	// Modifier le premier enfant et vérifier que le parent et le second
	// enfant ne sont pas affectés.
	child0 := moves[0].(*Taquin)
	child1 := moves[1].(*Taquin)
	parentBoard := taq.board

	child0.board[0] = 99
	if taq.board[0] != parentBoard[0] {
		t.Fatal("modifying child affected parent")
	}
	if child1.board[0] == 99 {
		t.Fatal("modifying child0 affected child1")
	}
}

func TestPossibleMoves_Terminal(t *testing.T) {
	taq := NewTaquin(2, 2, 1)
	rng := rand.New(rand.NewSource(1))
	taq.Shuffle(3, rng)
	// Épuiser les steps
	taq.steps = taq.maxSteps
	moves := taq.PossibleMoves()
	if len(moves) != 0 {
		t.Fatalf("terminal state should have 0 moves, got %d", len(moves))
	}
}

func TestID_DifferentSteps(t *testing.T) {
	taq1 := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	taq1.Shuffle(5, rng)
	id1 := taq1.ID()

	// Même config mais steps différent
	taq2 := *taq1
	taq2.steps = 10
	id2 := taq2.ID()

	if id1 == id2 {
		t.Fatal("IDs should differ when steps differ")
	}
}

func TestID_SameState(t *testing.T) {
	taq1 := NewTaquin(2, 3, 50)
	taq2 := NewTaquin(2, 3, 50)
	if taq1.ID() != taq2.ID() {
		t.Fatal("identical states should have identical IDs")
	}
}

func TestLastAction(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	taq.Shuffle(5, rng)

	moves := taq.PossibleMoves()
	for _, m := range moves {
		child := m.(*Taquin)
		dir := child.LastAction()
		if dir < Up || dir > Right {
			t.Fatalf("invalid direction %d", dir)
		}
	}
}

func TestFeatures(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	features := taq.Features()
	if len(features) != 6 {
		t.Fatalf("expected 6 features, got %d", len(features))
	}
	// Première tuile = 1, normalisée par 6
	expected := float32(1) / float32(6)
	if features[0] != expected {
		t.Fatalf("expected features[0] = %f, got %f", expected, features[0])
	}
	// Dernière case = 0 (vide)
	if features[5] != 0 {
		t.Fatalf("expected features[5] = 0, got %f", features[5])
	}
}

func TestFeatureShape(t *testing.T) {
	taq := NewTaquin(3, 4, 50)
	shape := taq.FeatureShape()
	if shape != [3]int{1, 3, 4} {
		t.Fatalf("expected [1, 3, 4], got %v", shape)
	}
}

func TestActionSize(t *testing.T) {
	taq := NewTaquin(2, 2, 10)
	if taq.ActionSize() != 4 {
		t.Fatalf("expected 4, got %d", taq.ActionSize())
	}
}

func TestSteps(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	if taq.Steps() != 0 {
		t.Fatalf("expected 0, got %d", taq.Steps())
	}
	if taq.MaxSteps() != 50 {
		t.Fatalf("expected 50, got %d", taq.MaxSteps())
	}
}

func TestRowsCols(t *testing.T) {
	taq := NewTaquin(3, 4, 50)
	if taq.Rows() != 3 || taq.Cols() != 4 {
		t.Fatalf("expected 3x4, got %dx%d", taq.Rows(), taq.Cols())
	}
}

func TestCanMove_Corners(t *testing.T) {
	// 3x3, case vide au centre (position 4) : 4 directions possibles
	taq := NewTaquin(3, 3, 50)
	taq.blank = 4
	dirs := taq.validDirections()
	if len(dirs) != 4 {
		t.Fatalf("center should have 4 directions, got %d", len(dirs))
	}

	// Coin haut-gauche (position 0) : Down et Right seulement
	taq.blank = 0
	dirs = taq.validDirections()
	if len(dirs) != 2 {
		t.Fatalf("top-left corner should have 2 directions, got %d", len(dirs))
	}
}

func TestRandomMove(t *testing.T) {
	taq := NewTaquin(3, 3, 50)
	rng := rand.New(rand.NewSource(42))
	taq.Shuffle(10, rng)

	child := taq.RandomMove(func(n int) int { return 0 }).(*Taquin)

	// L'enfant doit avoir un step de plus
	if child.steps != taq.steps+1 {
		t.Errorf("expected %d steps, got %d", taq.steps+1, child.steps)
	}

	// Le plateau doit être différent
	if child.board == taq.board && child.blank == taq.blank {
		t.Error("child board should differ from parent")
	}
}

func TestRandomMove_Independence(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	taq.Shuffle(10, rng)

	child := taq.RandomMove(func(n int) int { return 0 }).(*Taquin)
	parentBoard := taq.board

	child.board[0] = 99
	if taq.board[0] != parentBoard[0] {
		t.Error("mutating RandomMove child affected the parent state")
	}
}

func TestRandomMove_Distribution(t *testing.T) {
	taq := NewTaquin(3, 3, 50)
	// Case vide au centre (position 4) : 4 directions possibles
	taq.blank = 4
	// Remettre le board dans un état cohérent
	taq.board[4], taq.board[8] = taq.board[8], taq.board[4]

	dirs := taq.validDirections()
	for idx := range dirs {
		child := taq.RandomMove(func(n int) int { return idx }).(*Taquin)
		if child.LastAction() != dirs[idx] {
			t.Errorf("rng(%d): expected LastAction %d, got %d", idx, dirs[idx], child.LastAction())
		}
	}
}

func TestIsSolved(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	if !taq.isSolved() {
		t.Fatal("new taquin should be solved")
	}
	// Déplacer avec move() (interne, pas de check terminal)
	taq.move(Up) // case vide monte
	if taq.isSolved() {
		t.Fatal("should not be solved after a move")
	}
}
