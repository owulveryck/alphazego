package tictactoe

import (
	"testing"

	"github.com/owulveryck/alphazego/board"
)

func TestTicTacToe_String(t *testing.T) {
	ttt := &TicTacToe{
		board: []uint8{
			board.EmptyPlace, board.EmptyPlace, board.Player1,
			board.Player2, board.EmptyPlace, board.Player2,
			board.EmptyPlace, board.Player2, board.Player1,
		},
		PlayerTurn: board.Player1,
	}
	t.Log(ttt)
}
