package tictactoe

import (
	"strings"
	"testing"
)

func TestTicTacToe_String_NonEmpty(t *testing.T) {
	ttt := &TicTacToe{
		board: [BoardSize]uint8{
			0, 0, uint8(Cross),
			uint8(Circle), 0, uint8(Circle),
			0, uint8(Circle), uint8(Cross),
		},
		actorTurn: Cross,
	}
	s := ttt.String()
	if len(s) == 0 {
		t.Fatal("String() returned empty string")
	}
	// Vérifier que le rendu contient les symboles X et O (entourés de codes ANSI).
	if !strings.Contains(s, "X") {
		t.Error("String() should contain 'X' for Cross positions")
	}
	if !strings.Contains(s, "O") {
		t.Error("String() should contain 'O' for Circle positions")
	}
}

func TestTicTacToe_String_Empty(t *testing.T) {
	ttt := NewTicTacToe()
	s := ttt.String()
	if len(s) == 0 {
		t.Fatal("String() returned empty string for new game")
	}
	// Un plateau vide doit contenir les caractères de grille.
	if !strings.Contains(s, "┌") || !strings.Contains(s, "┘") {
		t.Error("String() of empty board should contain grid characters")
	}
}
