package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/board"
	"github.com/owulveryck/alphazego/decision/board/taquin"
	"github.com/owulveryck/alphazego/mcts"
)

func main() {
	puzzle := taquin.NewTaquin(5, 5, 50)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	puzzle.Shuffle(15, rng)

	fmt.Println("=== Taquin 3x2 ===")
	fmt.Println("L'IA résout le puzzle avec MCTS...")
	fmt.Println()
	fmt.Println(puzzle)

	m := mcts.NewMCTS()
	step := 0
	for puzzle.Evaluate() == decision.Undecided {
		bestState := m.RunMCTS(puzzle, 5000)
		if bestState == puzzle {
			fmt.Println("MCTS n'a pas trouvé de coup.")
			break
		}
		dir := bestState.(board.ActionRecorder).LastAction()
		dirNames := [4]string{"Haut", "Bas", "Gauche", "Droite"}
		step++
		fmt.Printf("Coup %d : %s\n", step, dirNames[dir])
		if err := puzzle.Play(dir); err != nil {
			fmt.Printf("Erreur : %v\n", err)
			break
		}
		fmt.Println(puzzle)
	}

	fmt.Println(strings.Repeat("─", 20))
	switch puzzle.Evaluate() {
	case taquin.Player:
		fmt.Printf("Puzzle résolu en %d coups !\n", puzzle.Steps())
	case decision.Stalemate:
		fmt.Println("Limite de coups atteinte, puzzle non résolu.")
	default:
		fmt.Println("Fin inattendue.")
	}
}
