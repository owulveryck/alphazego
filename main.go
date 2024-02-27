package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/owulveryck/alphazego/board"
	"github.com/owulveryck/alphazego/board/tictactoe"
	"github.com/owulveryck/alphazego/mcts"
)

func main() {
	ttt := tictactoe.NewTicTacToe()
	var move string
	m := mcts.NewMCTS()
	for ttt.Evaluate() == 0 {
		fmt.Println(ttt)
		fmt.Print("Enter your move: ")
		fmt.Scan(&move)
		// Convert string to uint64
		val, err := strconv.ParseUint(move, 10, 8) // Base 10, and up to 8 bits
		if err != nil {
			log.Fatal(err)
		}
		// Convert uint64 to uint8 since
		ttt.Play(uint8(val))
		ttt.Play(getNextMoveFromMCTS(m, ttt))
		// Now run the MCTS and change the state
	}
	fmt.Println(ttt)
}

func getNextMoveFromMCTS(m *mcts.MCTS, s board.State) board.Move {
	next := m.RunMCST(s)
	return s.(board.Playable).GetMoveFromState(next)
}
