package mcts_test

import (
	"fmt"

	"github.com/owulveryck/alphazego/board"
	"github.com/owulveryck/alphazego/board/tictactoe"
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
	fmt.Println("Next player after MCTS move:", bestState.CurrentPlayer())
	// Output:
	// Best state is not nil: true
	// Next player after MCTS move: 2
}

func ExampleMCTS_RunMCTS_extractMove() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Run MCTS to find best move
	bestState := m.RunMCTS(game, 1000)

	// Extract the actual move number (0-8) from the state change
	move := board.State(game).(board.Playable).GetMoveFromState(bestState)
	fmt.Println("MCTS chose a valid position:", move >= 0 && move <= 8)
	// Output:
	// MCTS chose a valid position: true
}

func ExampleMCTS_RunMCTS_takesWin() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Set up a board where Player1 can win at position 2:
	// X X _ | Player1's turn
	// O _ _
	// _ _ O
	game.Play(0) // X
	game.Play(3) // O
	game.Play(1) // X
	game.Play(8) // O

	// MCTS should find the winning move
	bestState := m.RunMCTS(game, 5000)
	move := board.State(game).(board.Playable).GetMoveFromState(bestState)
	fmt.Println("MCTS plays winning move at position 2:", move == 2)
	// Output:
	// MCTS plays winning move at position 2: true
}

func ExampleMCTS_RunMCTS_blocksOpponent() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Set up a board where Player1 threatens to win at position 2:
	// X X _ | Player2's turn
	// _ _ _
	// _ _ O
	game.Play(0) // X
	game.Play(8) // O
	game.Play(1) // X

	// MCTS (playing as Player2) should block at position 2
	bestState := m.RunMCTS(game, 5000)
	move := board.State(game).(board.Playable).GetMoveFromState(bestState)
	fmt.Println("MCTS blocks at position 2:", move == 2)
	// Output:
	// MCTS blocks at position 2: true
}

func ExampleMCTS_RunMCTS_terminalState() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Play until Player1 wins
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
	for game.Evaluate() == board.NoPlayer {
		bestState := m.RunMCTS(game, 500)
		move := board.State(game).(board.Playable).GetMoveFromState(bestState)
		game.Play(move)
		moves++
	}

	result := game.Evaluate()
	fmt.Println("Game finished:", result != board.NoPlayer)
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
	fmt.Println("Next player after move:", bestState.CurrentPlayer())
	// Output:
	// AlphaMCTS result is not nil: true
	// Next player after move: 2
}

// exampleEvaluator is a simple evaluator with uniform policy and neutral value.
type exampleEvaluator struct{}

func (e *exampleEvaluator) Evaluate(state board.State) ([]float64, float64) {
	moves := state.PossibleMoves()
	n := len(moves)
	if n == 0 {
		return nil, 0.0
	}
	policy := make([]float64, n)
	for i := range policy {
		policy[i] = 1.0 / float64(n)
	}
	return policy, 0.0
}

func ExampleMCTS_GetOrCreateNode() {
	m := mcts.NewMCTS()
	game := tictactoe.NewTicTacToe()

	// Create a root node
	node := m.GetOrCreateNode(game, nil)
	fmt.Println("Node created:", node != nil)

	// Requesting the same state returns the same node
	sameNode := m.GetOrCreateNode(game, nil)
	fmt.Println("Same node returned:", node == sameNode)
	// Output:
	// Node created: true
	// Same node returned: true
}
