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
	for ttt.Evaluate() == board.NoPlayer {
		fmt.Println(ttt)
		fmt.Print("Enter your move (0-8): ")
		fmt.Scan(&move)
		val, err := strconv.ParseUint(move, 10, 8)
		if err != nil {
			log.Fatal(err)
		}
		ttt.Play(uint8(val))
		if ttt.Evaluate() != board.NoPlayer {
			break
		}
		aiState := m.RunMCTS(ttt, 1000)
		aiMove := board.State(ttt).(board.Playable).GetMoveFromState(aiState)
		fmt.Printf("L'IA joue en %d\n", aiMove)
		ttt.Play(aiMove)
	}
	fmt.Println(ttt)
	switch ttt.Evaluate() {
	case board.Player1:
		fmt.Println("Vous avez gagne !")
	case board.Player2:
		fmt.Println("L'IA a gagne !")
	case board.DrawResult:
		fmt.Println("Match nul !")
	}
}
