package tictactoe

import (
	"testing"
)

func TestTicTacToe_String(t *testing.T) {
	ttt := &TicTacToe{
		board: [BoardSize]uint8{
			0, 0, uint8(Cross),
			uint8(Circle), 0, uint8(Circle),
			0, uint8(Circle), uint8(Cross),
		},
		actorTurn: Cross,
	}
	t.Log(ttt)
}
