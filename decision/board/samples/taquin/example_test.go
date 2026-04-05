package taquin_test

import (
	"fmt"
	"math/rand"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/board"
	"github.com/owulveryck/alphazego/decision/board/samples/taquin"
	"github.com/owulveryck/alphazego/mcts"
)

// Cet exemple montre que le taquin est un problème à un seul acteur :
// CurrentActor() et PreviousActor() retournent toujours [taquin.Player].
func Example_singleActor() {
	t := taquin.NewTaquin(2, 3, 50)
	fmt.Println("CurrentActor:", t.CurrentActor())
	fmt.Println("PreviousActor:", t.PreviousActor())
	fmt.Println("Same actor:", t.CurrentActor() == t.PreviousActor())
	// Output:
	// CurrentActor: 1
	// PreviousActor: 1
	// Same actor: true
}

// Cet exemple montre le MCTS résolvant un taquin 3x2 mélangé de 3 coups.
// Le MCTS trouve la solution en utilisant des rollouts aléatoires bornés
// par maxSteps.
func Example_mctsResolve() {
	puzzle := taquin.NewTaquin(2, 3, 20)
	rng := rand.New(rand.NewSource(42))
	puzzle.Shuffle(3, rng)

	m := mcts.NewMCTS()

	// Jouer coup par coup jusqu'à résolution ou échec
	solved := false
	for puzzle.Evaluate() == decision.Undecided {
		bestState := m.RunMCTS(puzzle, 5000)
		if bestState == puzzle {
			break // état terminal
		}
		dir := bestState.(board.ActionRecorder).LastAction()
		_ = puzzle.Play(dir)
		if puzzle.Evaluate() == taquin.Player {
			solved = true
			break
		}
	}
	fmt.Println("Puzzle solved:", solved)
	// Output:
	// Puzzle solved: true
}

// Cet exemple montre l'utilisation de [board.ActionRecorder] pour extraire
// la direction choisie par le MCTS.
func Example_actionRecorder() {
	puzzle := taquin.NewTaquin(2, 3, 20)
	rng := rand.New(rand.NewSource(42))
	puzzle.Shuffle(3, rng)

	m := mcts.NewMCTS()
	bestState := m.RunMCTS(puzzle, 1000)

	if ar, ok := bestState.(board.ActionRecorder); ok {
		dir := ar.LastAction()
		fmt.Println("Direction is valid:", dir >= taquin.Up && dir <= taquin.Right)
	}
	// Output:
	// Direction is valid: true
}
