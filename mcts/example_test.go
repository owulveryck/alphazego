package mcts_test

import (
	"fmt"
	"math/rand"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/board"
	"github.com/owulveryck/alphazego/decision/board/taquin"
	"github.com/owulveryck/alphazego/decision/board/tictactoe"
	"github.com/owulveryck/alphazego/mcts"
)

func ExampleNewMCTS() {
	m := mcts.NewMCTS()
	fmt.Println("MCTS instance created:", m != nil)
	// Output:
	// MCTS instance created: true
}

func ExampleMCTS_RunMCTS() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Run 1000 iterations to find the best move
	bestState := m.RunMCTS(game, 1000)

	// The result is a new state with one move played
	fmt.Println("Best state is not nil:", bestState != nil)
	fmt.Println("Next actor after MCTS move:", bestState.CurrentActor())
	// Output:
	// Best state is not nil: true
	// Next actor after MCTS move: 2
}

func ExampleMCTS_RunMCTS_extractMove() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Run MCTS to find best move
	bestState := m.RunMCTS(game, 1000)

	// Extract the actual move number (0-8) from the state change
	move := bestState.(board.ActionRecorder).LastAction()
	fmt.Println("MCTS chose a valid position:", move >= 0 && move <= 8)
	// Output:
	// MCTS chose a valid position: true
}

func ExampleMCTS_RunMCTS_takesWin() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Set up a board where Actor1 can win at position 2:
	// X X _ | Actor1's turn
	// O _ _
	// _ _ O
	game.Play(0) // X
	game.Play(3) // O
	game.Play(1) // X
	game.Play(8) // O

	// MCTS should find the winning move
	bestState := m.RunMCTS(game, 5000)
	move := bestState.(board.ActionRecorder).LastAction()
	fmt.Println("MCTS plays winning move at position 2:", move == 2)
	// Output:
	// MCTS plays winning move at position 2: true
}

func ExampleMCTS_RunMCTS_blocksOpponent() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Set up a board where Actor1 threatens to win at position 2:
	// X X _ | Actor2's turn
	// _ _ _
	// _ _ O
	game.Play(0) // X
	game.Play(8) // O
	game.Play(1) // X

	// MCTS (playing as Actor2) should block at position 2
	bestState := m.RunMCTS(game, 5000)
	move := bestState.(board.ActionRecorder).LastAction()
	fmt.Println("MCTS blocks at position 2:", move == 2)
	// Output:
	// MCTS blocks at position 2: true
}

func ExampleMCTS_RunMCTS_terminalState() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Play until Actor1 wins
	game.Play(0) // X
	game.Play(3) // O
	game.Play(1) // X
	game.Play(4) // O
	game.Play(2) // X wins (top row)

	// MCTS on a terminal state returns the same state
	result := m.RunMCTS(game, 100)
	fmt.Println("Returns original state for terminal:", result == game)
	// Output:
	// Returns original state for terminal: true
}

func ExampleMCTS_RunMCTS_fullGame() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Let MCTS play both sides until the game ends
	moves := 0
	for game.Evaluate() == decision.Undecided {
		bestState := m.RunMCTS(game, 500)
		move := bestState.(board.ActionRecorder).LastAction()
		game.Play(uint8(move))
		moves++
	}

	result := game.Evaluate()
	fmt.Println("Game finished:", result != decision.Undecided)
	fmt.Println("Moves played:", moves >= 5 && moves <= 9) // Min 5 moves for a win, max 9
	// Output:
	// Game finished: true
	// Moves played: true
}

func ExampleNewAlphaMCTS() {
	// Create a simple evaluator for testing
	eval := &exampleEvaluator{}

	// Create an AlphaZero-style MCTS with the evaluator and cpuct=1.5
	m := mcts.NewAlphaMCTS(eval, 1.5)
	game := tictactoe.NewTicTacToe()

	// Run MCTS with the evaluator guiding the search
	bestState := m.RunMCTS(game, 100)
	fmt.Println("AlphaMCTS result is not nil:", bestState != nil)
	fmt.Println("Next actor after move:", bestState.CurrentActor())
	// Output:
	// AlphaMCTS result is not nil: true
	// Next actor after move: 2
}

// Cet exemple montre AlphaMCTS résolvant un taquin (problème à un seul acteur).
// Grâce à la correction de backpropagateValue (pas d'alternance de signe quand
// CurrentActor == PreviousActor), le mode AlphaZero fonctionne pour les
// problèmes à un acteur.
func ExampleNewAlphaMCTS_singleActor() {
	eval := &exampleEvaluator{}
	m := mcts.NewAlphaMCTS(eval, 1.5)
	puzzle := taquin.NewTaquin(2, 3, 20)
	rng := rand.New(rand.NewSource(42))
	puzzle.Shuffle(3, rng)

	solved := false
	for puzzle.Evaluate() == decision.Undecided {
		bestState := m.RunMCTS(puzzle, 3000)
		if bestState == puzzle {
			break
		}
		dir := bestState.(board.ActionRecorder).LastAction()
		puzzle.Play(dir)
		if puzzle.Evaluate() == taquin.Player {
			solved = true
			break
		}
	}
	fmt.Println("AlphaMCTS solved taquin:", solved)
	// Output:
	// AlphaMCTS solved taquin: true
}

// exampleEvaluator is a simple evaluator with uniform policy and neutral values.
type exampleEvaluator struct{}

func (e *exampleEvaluator) Evaluate(state decision.State) ([]float64, map[decision.ActorID]float64) {
	moves := state.PossibleMoves()
	n := len(moves)
	if n == 0 {
		return nil, map[decision.ActorID]float64{}
	}
	policy := make([]float64, n)
	for i := range policy {
		policy[i] = 1.0 / float64(n)
	}
	values := map[decision.ActorID]float64{
		state.CurrentActor():  0.0,
		state.PreviousActor(): 0.0,
	}
	return policy, values
}
