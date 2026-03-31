package tictactoe

import (
	"testing"

	"github.com/owulveryck/alphazego/board"
)

func TestTicTacToe_String(t *testing.T) {
	ttt := &TicTacToe{
		board: []uint8{
			0, 0, uint8(board.Player1),
			uint8(board.Player2), 0, uint8(board.Player2),
			0, uint8(board.Player2), uint8(board.Player1),
		},
		PlayerTurn: board.Player1,
	}
	t.Log(ttt)
}
