package tictactoe

import (
	"testing"

	"github.com/owulveryck/alphazego/decision"
)

func TestNewTicTacToe(t *testing.T) {
	ttt := NewTicTacToe()
	if ttt.CurrentActor() != Cross {
		t.Errorf("expected Actor1 to start, got %d", ttt.CurrentActor())
	}
	for i, cell := range ttt.board {
		if cell != 0 {
			t.Errorf("expected empty cell at %d, got %d", i, cell)
		}
	}
}

func TestPlay(t *testing.T) {
	ttt := NewTicTacToe()
	ttt.Play(0)
	if ttt.board[0] != uint8(Cross) {
		t.Errorf("expected Actor1 at position 0, got %d", ttt.board[0])
	}
	if ttt.CurrentActor() != Circle {
		t.Errorf("expected Actor2 turn after Actor1 plays, got %d", ttt.CurrentActor())
	}
	ttt.Play(4)
	if ttt.board[4] != uint8(Circle) {
		t.Errorf("expected Actor2 at position 4, got %d", ttt.board[4])
	}
	if ttt.CurrentActor() != Cross {
		t.Errorf("expected Actor1 turn after Actor2 plays, got %d", ttt.CurrentActor())
	}
}

func TestPlay_Validation(t *testing.T) {
	ttt := NewTicTacToe()
	// Position hors limites
	if err := ttt.Play(9); err == nil {
		t.Error("expected error for out-of-bounds position")
	}
	// Coup valide
	if err := ttt.Play(0); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Case déjà occupée
	if err := ttt.Play(0); err == nil {
		t.Error("expected error for occupied cell")
	}
}

func TestCurrentActor(t *testing.T) {
	ttt := NewTicTacToe()
	if ttt.CurrentActor() != Cross {
		t.Errorf("expected Actor1, got %d", ttt.CurrentActor())
	}
	ttt.Play(0)
	if ttt.CurrentActor() != Circle {
		t.Errorf("expected Actor2, got %d", ttt.CurrentActor())
	}
}

func TestID(t *testing.T) {
	ttt := NewTicTacToe()
	id := ttt.ID()
	if len(id) != BoardSize+1 {
		t.Errorf("expected ID length %d, got %d", BoardSize+1, len(id))
	}

	// Two different states should have different IDs
	ttt2 := NewTicTacToe()
	ttt2.Play(0)
	id2 := ttt2.ID()
	if id == id2 {
		t.Error("expected different IDs for different states")
	}
}

func TestEvaluate_GameOn(t *testing.T) {
	ttt := NewTicTacToe()
	if ttt.Evaluate() != decision.Undecided {
		t.Errorf("expected NoActor for empty board, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Actor1WinsRow(t *testing.T) {
	ttt := &TicTacToe{
		board:     [BoardSize]uint8{1, 1, 1, 0, 0, 0, 0, 0, 0},
		actorTurn: Circle,
	}
	if ttt.Evaluate() != Cross {
		t.Errorf("expected Actor1 wins for top row, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Actor2WinsColumn(t *testing.T) {
	ttt := &TicTacToe{
		board:     [BoardSize]uint8{2, 0, 0, 2, 0, 0, 2, 0, 0},
		actorTurn: Cross,
	}
	if ttt.Evaluate() != Circle {
		t.Errorf("expected Actor2 wins for left column, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Actor1WinsDiagonal(t *testing.T) {
	ttt := &TicTacToe{
		board:     [BoardSize]uint8{1, 0, 0, 0, 1, 0, 0, 0, 1},
		actorTurn: Circle,
	}
	if ttt.Evaluate() != Cross {
		t.Errorf("expected Actor1 wins for diagonal, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Actor1WinsAntiDiagonal(t *testing.T) {
	ttt := &TicTacToe{
		board:     [BoardSize]uint8{0, 0, 1, 0, 1, 0, 1, 0, 0},
		actorTurn: Circle,
	}
	if ttt.Evaluate() != Cross {
		t.Errorf("expected Actor1 wins for anti-diagonal, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Draw(t *testing.T) {
	ttt := &TicTacToe{
		// X O X
		// X X O
		// O X O
		board:     [BoardSize]uint8{1, 2, 1, 1, 1, 2, 2, 1, 2},
		actorTurn: Cross,
	}
	if ttt.Evaluate() != decision.Stalemate {
		t.Errorf("expected DrawResult, got %d", ttt.Evaluate())
	}
}

func TestPossibleMoves(t *testing.T) {
	ttt := NewTicTacToe()
	moves := ttt.PossibleMoves()
	if len(moves) != 9 {
		t.Errorf("expected 9 possible moves for empty board, got %d", len(moves))
	}

	// After one move, should have 8 possible moves
	ttt.Play(0)
	moves = ttt.PossibleMoves()
	if len(moves) != 8 {
		t.Errorf("expected 8 possible moves, got %d", len(moves))
	}

	// Each move should have the correct current actor
	for _, m := range moves {
		s := m.(*TicTacToe)
		if s.CurrentActor() != Cross {
			t.Errorf("expected Actor1 turn in child state, got %d", s.CurrentActor())
		}
	}
}

func TestPossibleMoves_FullBoard(t *testing.T) {
	ttt := &TicTacToe{
		board:     [BoardSize]uint8{1, 2, 1, 1, 1, 2, 2, 1, 2},
		actorTurn: Cross,
	}
	moves := ttt.PossibleMoves()
	if len(moves) != 0 {
		t.Errorf("expected 0 possible moves for full board, got %d", len(moves))
	}
}

func TestLastAction(t *testing.T) {
	ttt := NewTicTacToe()
	ttt.Play(4) // Actor1 plays at 4
	if ttt.LastAction() != 4 {
		t.Errorf("expected LastAction 4, got %d", ttt.LastAction())
	}

	// PossibleMoves should set LastAction on each child
	ttt2 := NewTicTacToe()
	moves := ttt2.PossibleMoves()
	for i, m := range moves {
		child := m.(*TicTacToe)
		if child.LastAction() != i {
			t.Errorf("expected child LastAction %d, got %d", i, child.LastAction())
		}
	}
}

func TestFeatures(t *testing.T) {
	ttt := NewTicTacToe()
	features := ttt.Features()
	if len(features) != 27 {
		t.Fatalf("expected 27 features, got %d", len(features))
	}
	// Empty board: all features should be 0 except plan 2 (actor indicator)
	for i := 0; i < 18; i++ {
		if features[i] != 0 {
			t.Errorf("expected 0 at index %d for empty board, got %f", i, features[i])
		}
	}
	// Actor1 starts, so plan 2 should be all 1.0
	for i := 18; i < 27; i++ {
		if features[i] != 1.0 {
			t.Errorf("expected 1.0 at index %d (actor indicator), got %f", i, features[i])
		}
	}

	// After Actor1 plays at 0, Actor2 plays at 4
	ttt.Play(0)
	ttt.Play(4)
	// Now it's Actor1's turn
	features = ttt.Features()
	// Plan 0 (current = Actor1): position 0 should be 1.0
	if features[0] != 1.0 {
		t.Errorf("expected 1.0 at plan0[0] (Actor1 piece), got %f", features[0])
	}
	// Plan 1 (opponent = Actor2): position 4 should be 1.0
	if features[9+4] != 1.0 {
		t.Errorf("expected 1.0 at plan1[4] (Actor2 piece), got %f", features[9+4])
	}
	// Actor1's piece should NOT appear in plan 1
	if features[9+0] != 0 {
		t.Errorf("expected 0 at plan1[0], got %f", features[9+0])
	}
}

func TestFeatures_Actor2Turn(t *testing.T) {
	ttt := NewTicTacToe()
	ttt.Play(0) // Actor1 plays, now Actor2's turn
	features := ttt.Features()
	// Plan 2: Actor2's turn, so indicator should be 0.0
	for i := 18; i < 27; i++ {
		if features[i] != 0.0 {
			t.Errorf("expected 0.0 at index %d (Actor2 turn), got %f", i, features[i])
		}
	}
	// Plan 0 (current = Actor2): position 0 should be 0 (that's Actor1's piece)
	if features[0] != 0 {
		t.Errorf("expected 0 at plan0[0] (not current actor's piece), got %f", features[0])
	}
	// Plan 1 (opponent = Actor1): position 0 should be 1.0
	if features[9+0] != 1.0 {
		t.Errorf("expected 1.0 at plan1[0] (opponent piece), got %f", features[9+0])
	}
}

func TestFeatureShape(t *testing.T) {
	ttt := NewTicTacToe()
	shape := ttt.FeatureShape()
	expected := [3]int{3, 3, 3}
	if shape != expected {
		t.Errorf("expected %v, got %v", expected, shape)
	}
}

func TestActionSize(t *testing.T) {
	ttt := NewTicTacToe()
	if ttt.ActionSize() != 9 {
		t.Errorf("expected ActionSize 9, got %d", ttt.ActionSize())
	}
}

func TestEvaluate_AllWinningPositions(t *testing.T) {
	// Test all rows, columns, diagonals for both actors
	tests := []struct {
		name   string
		board  [BoardSize]uint8
		winner decision.ActorID
	}{
		{"row0-a1", [BoardSize]uint8{1, 1, 1, 0, 0, 0, 0, 0, 0}, Cross},
		{"row1-a1", [BoardSize]uint8{0, 0, 0, 1, 1, 1, 0, 0, 0}, Cross},
		{"row2-a1", [BoardSize]uint8{0, 0, 0, 0, 0, 0, 1, 1, 1}, Cross},
		{"col0-a2", [BoardSize]uint8{2, 0, 0, 2, 0, 0, 2, 0, 0}, Circle},
		{"col1-a2", [BoardSize]uint8{0, 2, 0, 0, 2, 0, 0, 2, 0}, Circle},
		{"col2-a2", [BoardSize]uint8{0, 0, 2, 0, 0, 2, 0, 0, 2}, Circle},
		{"diag-a2", [BoardSize]uint8{2, 0, 0, 0, 2, 0, 0, 0, 2}, Circle},
		{"anti-a2", [BoardSize]uint8{0, 0, 2, 0, 2, 0, 2, 0, 0}, Circle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttt := &TicTacToe{board: tt.board, actorTurn: Cross}
			if ttt.Evaluate() != tt.winner {
				t.Errorf("expected %d, got %d", tt.winner, ttt.Evaluate())
			}
		})
	}
}

// TestPossibleMoves_Independence vérifie que muter un état retourné par
// PossibleMoves n'affecte pas les autres états ni l'état parent.
func TestPossibleMoves_Independence(t *testing.T) {
	ttt := NewTicTacToe()
	ttt.Play(0)
	moves := ttt.PossibleMoves()
	if len(moves) < 2 {
		t.Fatal("expected at least 2 moves")
	}

	// Sauvegarder l'ID du parent et du second enfant
	parentID := ttt.ID()
	secondChildID := moves[1].ID()

	// Muter le premier enfant
	first := moves[0].(*TicTacToe)
	first.board[8] = 42 // valeur arbitraire

	// Vérifier que le parent n'est pas affecté
	if ttt.ID() != parentID {
		t.Error("mutating a child affected the parent state")
	}

	// Vérifier que le second enfant n'est pas affecté
	if moves[1].ID() != secondChildID {
		t.Error("mutating one child affected another child")
	}
}

// TestPossibleMoves_SingleMove vérifie le comportement avec un seul coup légal.
func TestRandomMove(t *testing.T) {
	ttt := NewTicTacToe()
	ttt.Play(0) // X at 0, now O's turn
	ttt.Play(4) // O at 4, now X's turn

	// Utiliser un rng déterministe qui retourne toujours 0
	child := ttt.RandomMove(func(n int) int { return 0 }).(*TicTacToe)

	// L'enfant doit avoir une case de plus remplie
	parentFilled := 0
	childFilled := 0
	for i := 0; i < BoardSize; i++ {
		if ttt.board[i] != 0 {
			parentFilled++
		}
		if child.board[i] != 0 {
			childFilled++
		}
	}
	if childFilled != parentFilled+1 {
		t.Errorf("expected %d filled cells in child, got %d", parentFilled+1, childFilled)
	}

	// L'acteur doit avoir changé
	if child.CurrentActor() != Circle {
		t.Errorf("expected Circle's turn in child, got %d", child.CurrentActor())
	}
}

func TestRandomMove_Independence(t *testing.T) {
	ttt := NewTicTacToe()
	ttt.Play(0)

	child := ttt.RandomMove(func(n int) int { return 0 }).(*TicTacToe)
	parentID := ttt.ID()

	// Muter l'enfant ne doit pas affecter le parent
	child.board[8] = 42
	if ttt.ID() != parentID {
		t.Error("mutating RandomMove child affected the parent state")
	}
}

func TestRandomMove_Distribution(t *testing.T) {
	ttt := NewTicTacToe()
	ttt.Play(0) // X at 0
	ttt.Play(4) // O at 4
	// 7 cases vides : 1,2,3,5,6,7,8

	emptyCells := []int{1, 2, 3, 5, 6, 7, 8}
	for idx, expected := range emptyCells {
		child := ttt.RandomMove(func(n int) int { return idx }).(*TicTacToe)
		if child.LastAction() != expected {
			t.Errorf("rng(%d): expected LastAction %d, got %d", idx, expected, child.LastAction())
		}
	}
}

func TestPossibleMoves_SingleMove(t *testing.T) {
	// Plateau presque complet : une seule case vide (position 8)
	ttt := &TicTacToe{
		board:     [BoardSize]uint8{1, 2, 1, 2, 1, 2, 2, 1, 0},
		actorTurn: Cross,
	}
	moves := ttt.PossibleMoves()
	if len(moves) != 1 {
		t.Fatalf("expected 1 possible move, got %d", len(moves))
	}
	child := moves[0].(*TicTacToe)
	if child.LastAction() != 8 {
		t.Errorf("expected LastAction 8, got %d", child.LastAction())
	}
}
