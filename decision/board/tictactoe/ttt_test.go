package tictactoe

import (
	"testing"

	"github.com/owulveryck/alphazego/decision"
)

func TestNewTicTacToe(t *testing.T) {
	ttt := NewTicTacToe()
	if ttt.ActorTurn != decision.Actor1 {
		t.Errorf("expected Actor1 to start, got %d", ttt.ActorTurn)
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
	if ttt.board[0] != uint8(decision.Actor1) {
		t.Errorf("expected Actor1 at position 0, got %d", ttt.board[0])
	}
	if ttt.ActorTurn != decision.Actor2 {
		t.Errorf("expected Actor2 turn after Actor1 plays, got %d", ttt.ActorTurn)
	}
	ttt.Play(4)
	if ttt.board[4] != uint8(decision.Actor2) {
		t.Errorf("expected Actor2 at position 4, got %d", ttt.board[4])
	}
	if ttt.ActorTurn != decision.Actor1 {
		t.Errorf("expected Actor1 turn after Actor2 plays, got %d", ttt.ActorTurn)
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
	// Case deja occupee
	if err := ttt.Play(0); err == nil {
		t.Error("expected error for occupied cell")
	}
}

func TestCurrentActor(t *testing.T) {
	ttt := NewTicTacToe()
	if ttt.CurrentActor() != decision.Actor1 {
		t.Errorf("expected Actor1, got %d", ttt.CurrentActor())
	}
	ttt.Play(0)
	if ttt.CurrentActor() != decision.Actor2 {
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
	if ttt.Evaluate() != decision.NoActor {
		t.Errorf("expected NoActor for empty board, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Actor1WinsRow(t *testing.T) {
	ttt := &TicTacToe{
		board:     []uint8{1, 1, 1, 0, 0, 0, 0, 0, 0},
		ActorTurn: decision.Actor2,
	}
	if ttt.Evaluate() != decision.Actor1 {
		t.Errorf("expected Actor1 wins for top row, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Actor2WinsColumn(t *testing.T) {
	ttt := &TicTacToe{
		board:     []uint8{2, 0, 0, 2, 0, 0, 2, 0, 0},
		ActorTurn: decision.Actor1,
	}
	if ttt.Evaluate() != decision.Actor2 {
		t.Errorf("expected Actor2 wins for left column, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Actor1WinsDiagonal(t *testing.T) {
	ttt := &TicTacToe{
		board:     []uint8{1, 0, 0, 0, 1, 0, 0, 0, 1},
		ActorTurn: decision.Actor2,
	}
	if ttt.Evaluate() != decision.Actor1 {
		t.Errorf("expected Actor1 wins for diagonal, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Actor1WinsAntiDiagonal(t *testing.T) {
	ttt := &TicTacToe{
		board:     []uint8{0, 0, 1, 0, 1, 0, 1, 0, 0},
		ActorTurn: decision.Actor2,
	}
	if ttt.Evaluate() != decision.Actor1 {
		t.Errorf("expected Actor1 wins for anti-diagonal, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Draw(t *testing.T) {
	ttt := &TicTacToe{
		// X O X
		// X X O
		// O X O
		board:     []uint8{1, 2, 1, 1, 1, 2, 2, 1, 2},
		ActorTurn: decision.Actor1,
	}
	if ttt.Evaluate() != decision.DrawResult {
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
		if s.CurrentActor() != decision.Actor1 {
			t.Errorf("expected Actor1 turn in child state, got %d", s.CurrentActor())
		}
	}
}

func TestPossibleMoves_FullBoard(t *testing.T) {
	ttt := &TicTacToe{
		board:     []uint8{1, 2, 1, 1, 1, 2, 2, 1, 2},
		ActorTurn: decision.Actor1,
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

func TestToDecisionState(t *testing.T) {
	games := []*TicTacToe{NewTicTacToe(), NewTicTacToe()}
	states := toDecisionState(games)
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
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
		board  []uint8
		winner decision.ActorID
	}{
		{"row0-a1", []uint8{1, 1, 1, 0, 0, 0, 0, 0, 0}, decision.Actor1},
		{"row1-a1", []uint8{0, 0, 0, 1, 1, 1, 0, 0, 0}, decision.Actor1},
		{"row2-a1", []uint8{0, 0, 0, 0, 0, 0, 1, 1, 1}, decision.Actor1},
		{"col0-a2", []uint8{2, 0, 0, 2, 0, 0, 2, 0, 0}, decision.Actor2},
		{"col1-a2", []uint8{0, 2, 0, 0, 2, 0, 0, 2, 0}, decision.Actor2},
		{"col2-a2", []uint8{0, 0, 2, 0, 0, 2, 0, 0, 2}, decision.Actor2},
		{"diag-a2", []uint8{2, 0, 0, 0, 2, 0, 0, 0, 2}, decision.Actor2},
		{"anti-a2", []uint8{0, 0, 2, 0, 2, 0, 2, 0, 0}, decision.Actor2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttt := &TicTacToe{board: tt.board, ActorTurn: decision.Actor1}
			if ttt.Evaluate() != tt.winner {
				t.Errorf("expected %d, got %d", tt.winner, ttt.Evaluate())
			}
		})
	}
}
