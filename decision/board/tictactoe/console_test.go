package tictactoe

import (
	"testing"

	"github.com/owulveryck/alphazego/decision"
)

func TestTicTacToe_String(t *testing.T) {
	ttt := &TicTacToe{
		board: []uint8{
			0, 0, uint8(decision.Actor1),
			uint8(decision.Actor2), 0, uint8(decision.Actor2),
			0, uint8(decision.Actor2), uint8(decision.Actor1),
		},
		ActorTurn: decision.Actor1,
	}
	t.Log(ttt)
}
