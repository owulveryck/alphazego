package tictactoe

import (
	"testing"

	"github.com/owulveryck/alphazego/board"
)

func TestNewTicTacToe(t *testing.T) {
	ttt := NewTicTacToe()
	if ttt.PlayerTurn != board.Player1 {
		t.Errorf("expected Player1 to start, got %d", ttt.PlayerTurn)
	}
	for i, cell := range ttt.board {
		if cell != board.EmptyPlace {
			t.Errorf("expected empty cell at %d, got %d", i, cell)
		}
	}
}

func TestPlay(t *testing.T) {
	ttt := NewTicTacToe()
	ttt.Play(0)
	if ttt.board[0] != board.Player1 {
		t.Errorf("expected Player1 at position 0, got %d", ttt.board[0])
	}
	if ttt.PlayerTurn != board.Player2 {
		t.Errorf("expected Player2 turn after Player1 plays, got %d", ttt.PlayerTurn)
	}
	ttt.Play(4)
	if ttt.board[4] != board.Player2 {
		t.Errorf("expected Player2 at position 4, got %d", ttt.board[4])
	}
	if ttt.PlayerTurn != board.Player1 {
		t.Errorf("expected Player1 turn after Player2 plays, got %d", ttt.PlayerTurn)
	}
}

func TestCurrentPlayer(t *testing.T) {
	ttt := NewTicTacToe()
	if ttt.CurrentPlayer() != board.Player1 {
		t.Errorf("expected Player1, got %d", ttt.CurrentPlayer())
	}
	ttt.Play(0)
	if ttt.CurrentPlayer() != board.Player2 {
		t.Errorf("expected Player2, got %d", ttt.CurrentPlayer())
	}
}

func TestBoardID(t *testing.T) {
	ttt := NewTicTacToe()
	id := ttt.BoardID()
	if len(id) != BoardSize+1 {
		t.Errorf("expected BoardID length %d, got %d", BoardSize+1, len(id))
	}
	// Last byte should be the current player
	if id[BoardSize] != board.Player1 {
		t.Errorf("expected last byte to be Player1, got %d", id[BoardSize])
	}

	// Two different states should have different IDs
	ttt2 := NewTicTacToe()
	ttt2.Play(0)
	id2 := ttt2.BoardID()
	if string(id) == string(id2) {
		t.Error("expected different BoardIDs for different states")
	}
}

func TestEvaluate_GameOn(t *testing.T) {
	ttt := NewTicTacToe()
	if ttt.Evaluate() != board.GameOn {
		t.Errorf("expected GameOn for empty board, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Player1WinsRow(t *testing.T) {
	ttt := &TicTacToe{
		board:      []uint8{1, 1, 1, 0, 0, 0, 0, 0, 0},
		PlayerTurn: board.Player2,
	}
	if ttt.Evaluate() != board.Player1Wins {
		t.Errorf("expected Player1Wins for top row, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Player2WinsColumn(t *testing.T) {
	ttt := &TicTacToe{
		board:      []uint8{2, 0, 0, 2, 0, 0, 2, 0, 0},
		PlayerTurn: board.Player1,
	}
	if ttt.Evaluate() != board.Player2Wins {
		t.Errorf("expected Player2Wins for left column, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Player1WinsDiagonal(t *testing.T) {
	ttt := &TicTacToe{
		board:      []uint8{1, 0, 0, 0, 1, 0, 0, 0, 1},
		PlayerTurn: board.Player2,
	}
	if ttt.Evaluate() != board.Player1Wins {
		t.Errorf("expected Player1Wins for diagonal, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Player1WinsAntiDiagonal(t *testing.T) {
	ttt := &TicTacToe{
		board:      []uint8{0, 0, 1, 0, 1, 0, 1, 0, 0},
		PlayerTurn: board.Player2,
	}
	if ttt.Evaluate() != board.Player1Wins {
		t.Errorf("expected Player1Wins for anti-diagonal, got %d", ttt.Evaluate())
	}
}

func TestEvaluate_Draw(t *testing.T) {
	ttt := &TicTacToe{
		// X O X
		// X X O
		// O X O
		board:      []uint8{1, 2, 1, 1, 1, 2, 2, 1, 2},
		PlayerTurn: board.Player1,
	}
	if ttt.Evaluate() != board.Draw {
		t.Errorf("expected Draw, got %d", ttt.Evaluate())
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

	// Each move should have the correct current player
	for _, m := range moves {
		s := m.(*TicTacToe)
		if s.CurrentPlayer() != board.Player1 {
			t.Errorf("expected Player1 turn in child state, got %d", s.CurrentPlayer())
		}
	}
}

func TestPossibleMoves_FullBoard(t *testing.T) {
	ttt := &TicTacToe{
		board:      []uint8{1, 2, 1, 1, 1, 2, 2, 1, 2},
		PlayerTurn: board.Player1,
	}
	moves := ttt.PossibleMoves()
	if len(moves) != 0 {
		t.Errorf("expected 0 possible moves for full board, got %d", len(moves))
	}
}

func TestGetMoveFromState(t *testing.T) {
	ttt := NewTicTacToe()
	ttt.Play(0) // Player1 plays at 0

	next := &TicTacToe{
		board:      []uint8{1, 0, 0, 0, 2, 0, 0, 0, 0},
		PlayerTurn: board.Player1,
	}
	move := ttt.GetMoveFromState(next)
	if move != 4 {
		t.Errorf("expected move 4, got %d", move)
	}
}

func TestToBoardState(t *testing.T) {
	games := []*TicTacToe{NewTicTacToe(), NewTicTacToe()}
	states := toBoardState(games)
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}
}

func TestEvaluate_AllWinningPositions(t *testing.T) {
	// Test all rows, columns, diagonals for both players
	tests := []struct {
		name   string
		board  []uint8
		winner board.Result
	}{
		{"row0-p1", []uint8{1, 1, 1, 0, 0, 0, 0, 0, 0}, board.Player1Wins},
		{"row1-p1", []uint8{0, 0, 0, 1, 1, 1, 0, 0, 0}, board.Player1Wins},
		{"row2-p1", []uint8{0, 0, 0, 0, 0, 0, 1, 1, 1}, board.Player1Wins},
		{"col0-p2", []uint8{2, 0, 0, 2, 0, 0, 2, 0, 0}, board.Player2Wins},
		{"col1-p2", []uint8{0, 2, 0, 0, 2, 0, 0, 2, 0}, board.Player2Wins},
		{"col2-p2", []uint8{0, 0, 2, 0, 0, 2, 0, 0, 2}, board.Player2Wins},
		{"diag-p2", []uint8{2, 0, 0, 0, 2, 0, 0, 0, 2}, board.Player2Wins},
		{"anti-p2", []uint8{0, 0, 2, 0, 2, 0, 2, 0, 0}, board.Player2Wins},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttt := &TicTacToe{board: tt.board, PlayerTurn: board.Player1}
			if ttt.Evaluate() != tt.winner {
				t.Errorf("expected %d, got %d", tt.winner, ttt.Evaluate())
			}
		})
	}
}
