package main

import (
	"github.com/owulveryck/alphazego/board"
	"github.com/owulveryck/alphazego/board/tictactoe"
	"github.com/owulveryck/alphazego/mcts"
)

func main() {
	var ttt board.State
	ttt = &tictactoe.TicTacToe{
		PlayerTurn: board.Player1,
	}
	mcts := &mcts.MCTS{}
	for ttt.Evaluate() == board.GameOn {
		ttt = mcts.RunMCST(ttt)
	}
}
