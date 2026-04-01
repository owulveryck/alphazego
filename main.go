package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/board"
	"github.com/owulveryck/alphazego/decision/board/tictactoe"
	"github.com/owulveryck/alphazego/mcts"
)

func main() {
	ttt := tictactoe.NewTicTacToe()
	var move string
	m := mcts.NewMCTS()
	for ttt.Evaluate() == decision.NoActor {
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
		if ttt.Evaluate() != decision.NoActor {
			break
		}
		aiMove := getNextMoveFromMCTS(m, ttt)
		fmt.Printf("L'IA joue en %d\n", aiMove)
		if err := ttt.Play(aiMove); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(ttt)
	switch ttt.Evaluate() {
	case decision.Actor1:
		fmt.Println("Vous avez gagne !")
	case decision.Actor2:
		fmt.Println("L'IA a gagne !")
	case decision.DrawResult:
		fmt.Println("Match nul !")
	}
}

func getNextMoveFromMCTS(m *mcts.MCTS, s decision.State) uint8 {
	next := m.RunMCTS(s, 1000)
	return uint8(next.(board.ActionRecorder).LastAction())
}
