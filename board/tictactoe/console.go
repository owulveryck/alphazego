package tictactoe

import (
	"strings"

	"github.com/owulveryck/alphazego/board"
)

func (tictactoe *TicTacToe) String() string {
	players := map[board.Agent]string{
		board.EmptyPlace: " ",
		board.Player1:    "X",
		board.Player2:    "O",
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
