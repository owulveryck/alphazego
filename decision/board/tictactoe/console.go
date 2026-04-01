package tictactoe

import (
	"strings"

	"github.com/owulveryck/alphazego/decision"
)

const (
	reset = "\033[0m"
	red   = "\033[31m"
	blue  = "\033[34m"
)

// String returns a human-readable representation of the board using
// ANSI colors: red for Actor1 (X) and blue for Actor2 (O).
// It also displays whose turn it is.
func (tictactoe *TicTacToe) String() string {
	symbols := map[decision.ActorID]string{
		decision.NoActor: " ",
		decision.Actor1:  reset + red + "X" + reset,
		decision.Actor2:  reset + blue + "O" + reset,
	}
	cellSymbol := func(i int) string {
		return symbols[decision.ActorID(tictactoe.board[i])]
	}
	var b strings.Builder
	b.WriteString("Current actor: " + symbols[tictactoe.ActorTurn] + "\n")
	b.WriteString(" ┌───┬───┬───┐\n")
	b.WriteString(" │ " + cellSymbol(0))
	b.WriteString(" │ ")
	b.WriteString(cellSymbol(1))
	b.WriteString(" │ ")
	b.WriteString(cellSymbol(2))
	b.WriteString(" │\n")
	b.WriteString(" ├───┼───┼───┤\n")
	b.WriteString(" │ " + cellSymbol(3))
	b.WriteString(" │ ")
	b.WriteString(cellSymbol(4))
	b.WriteString(" │ ")
	b.WriteString(cellSymbol(5))
	b.WriteString(" │\n")
	b.WriteString(" ├───┼───┼───┤\n")
	b.WriteString(" │ " + cellSymbol(6))
	b.WriteString(" │ ")
	b.WriteString(cellSymbol(7))
	b.WriteString(" │ ")
	b.WriteString(cellSymbol(8))
	b.WriteString(" │\n")
	b.WriteString(" └───┴───┴───┘\n")
	return b.String()
}
