package tictactoe

import (
	"strings"

	"github.com/owulveryck/alphazego/board"
)

const (
	reset = "\033[0m"
	red   = "\033[31m"
	blue  = "\033[34m"
)

func (tictactoe *TicTacToe) String() string {
	players := map[board.Agent]string{
		board.EmptyPlace: " ",
		board.Player1:    reset + red + "X" + reset,
		board.Player2:    reset + blue + "O" + reset,
	}
	var b strings.Builder
	b.WriteString("Current player: " + players[tictactoe.PlayerTurn] + "\n")
	b.WriteString(" ┌───┬───┬───┐\n")
	b.WriteString(" │ " + players[tictactoe.board[0]])
	b.WriteString(" │ ")
	b.WriteString(players[tictactoe.board[1]])
	b.WriteString(" │ ")
	b.WriteString(players[tictactoe.board[2]])
	b.WriteString(" │\n")
	b.WriteString(" ├───┼───┼───┤\n")
	b.WriteString(" │ " + players[tictactoe.board[3]])
	b.WriteString(" │ ")
	b.WriteString(players[tictactoe.board[4]])
	b.WriteString(" │ ")
	b.WriteString(players[tictactoe.board[5]])
	b.WriteString(" │\n")
	b.WriteString(" ├───┼───┼───┤\n")
	b.WriteString(" │ " + players[tictactoe.board[6]])
	b.WriteString(" │ ")
	b.WriteString(players[tictactoe.board[7]])
	b.WriteString(" │ ")
	b.WriteString(players[tictactoe.board[8]])
	b.WriteString(" │\n")
	b.WriteString(" └───┴───┴───┘\n")
	return b.String()
}
