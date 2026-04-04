package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/board"
	"github.com/owulveryck/alphazego/decision/board/samples/tictactoe"
	"github.com/owulveryck/alphazego/mcts"
)

func main() {
	ttt := tictactoe.NewTicTacToe()
	var move string
	m := mcts.NewMCTS()
	for ttt.Evaluate() == decision.Undecided {
		fmt.Println(ttt)
		fmt.Print("Enter your move (0-8): ")
		fmt.Scan(&move)
		val, err := strconv.ParseUint(move, 10, 8)
		if err != nil {
			log.Fatal(err)
		}
		if err := ttt.Play(uint8(val)); err != nil {
			fmt.Println("Coup invalide :", err)
			continue
		}
		if ttt.Evaluate() != decision.Undecided {
			break
		}
		aiState := m.RunMCTS(ttt, 1000)
		aiMove := uint8(aiState.(board.ActionRecorder).LastAction())
		fmt.Printf("L'IA joue en %d\n", aiMove)
		if err := ttt.Play(aiMove); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(ttt)
	switch ttt.Evaluate() {
	case tictactoe.Cross:
		fmt.Println("Vous avez gagné !")
	case tictactoe.Circle:
		fmt.Println("L'IA a gagné !")
	case decision.Stalemate:
		fmt.Println("Match nul !")
	}
}
