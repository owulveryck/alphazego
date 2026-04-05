package tictactoe_test

import (
	"fmt"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/board"
	"github.com/owulveryck/alphazego/decision/board/samples/tictactoe"
)

func ExampleNewTicTacToe() {
	game := tictactoe.NewTicTacToe()
	fmt.Println("Current actor:", game.CurrentActor())
	fmt.Println("Game result:", game.Evaluate())
	// Output:
	// Current actor: 1
	// Game result: 0
}

func ExampleTicTacToe_Play() {
	game := tictactoe.NewTicTacToe()

	// Actor1 plays at center (position 4)
	game.Play(4)
	fmt.Println("After Actor1 plays at 4, current actor:", game.CurrentActor())

	// Actor2 plays at top-left (position 0)
	game.Play(0)
	fmt.Println("After Actor2 plays at 0, current actor:", game.CurrentActor())
	// Output:
	// After Actor1 plays at 4, current actor: 2
	// After Actor2 plays at 0, current actor: 1
}

func ExampleTicTacToe_Evaluate_gameOn() {
	game := tictactoe.NewTicTacToe()
	game.Play(0) // X at top-left
	game.Play(4) // O at center

	result := game.Evaluate()
	fmt.Println("Game still in progress:", result == decision.Undecided)
	// Output:
	// Game still in progress: true
}

func ExampleTicTacToe_Evaluate_actor1Wins() {
	game := tictactoe.NewTicTacToe()
	// Actor1 takes the top row: 0, 1, 2
	game.Play(0) // X
	game.Play(3) // O
	game.Play(1) // X
	game.Play(4) // O
	game.Play(2) // X wins!

	result := game.Evaluate()
	fmt.Println("Actor1 wins:", result == tictactoe.Cross)
	// Output:
	// Actor1 wins: true
}

func ExampleTicTacToe_Evaluate_draw() {
	game := tictactoe.NewTicTacToe()
	// Play a sequence leading to a draw:
	// X O X
	// X X O
	// O X O
	game.Play(0) // X
	game.Play(1) // O
	game.Play(2) // X
	game.Play(5) // O
	game.Play(3) // X
	game.Play(6) // O
	game.Play(4) // X
	game.Play(8) // O
	game.Play(7) // X

	result := game.Evaluate()
	fmt.Println("Draw:", result == decision.Stalemate)
	// Output:
	// Draw: true
}

func ExampleTicTacToe_PossibleMoves() {
	game := tictactoe.NewTicTacToe()

	// Empty board has 9 possible moves
	moves := game.PossibleMoves()
	fmt.Println("Moves from empty board:", len(moves))

	// After one move, 8 remain
	game.Play(4)
	moves = game.PossibleMoves()
	fmt.Println("Moves after one play:", len(moves))
	// Output:
	// Moves from empty board: 9
	// Moves after one play: 8
}

func ExampleTicTacToe_PossibleMoves_alternation() {
	game := tictactoe.NewTicTacToe() // Actor1's turn
	moves := game.PossibleMoves()

	// Each child state has the other actor's turn
	for _, m := range moves {
		if m.CurrentActor() != tictactoe.Circle {
			fmt.Println("ERROR: expected Actor2's turn in child state")
			return
		}
	}
	fmt.Println("All child states have Actor2's turn: true")
	// Output:
	// All child states have Actor2's turn: true
}

func ExampleTicTacToe_PreviousActor() {
	game := tictactoe.NewTicTacToe()
	// Sur un plateau vierge, l'acteur courant est Actor1 ;
	// PreviousActor retourne Actor2 (le "dernier" dans l'alternance).
	fmt.Println("Previous actor (initial):", game.PreviousActor())

	game.Play(4) // Actor1 joue au centre
	fmt.Println("Previous actor after A1 plays:", game.PreviousActor())

	game.Play(0) // Actor2 joue en haut à gauche
	fmt.Println("Previous actor after A2 plays:", game.PreviousActor())
	// Output:
	// Previous actor (initial): 2
	// Previous actor after A1 plays: 1
	// Previous actor after A2 plays: 2
}

func ExampleTicTacToe_LastAction() {
	game := tictactoe.NewTicTacToe()
	game.Play(0) // Actor1 at position 0
	// Now it's Actor2's turn. Each possible next state knows its LastAction.

	moves := game.PossibleMoves()
	// Find the state where position 4 was played
	for _, next := range moves {
		s := next.(*tictactoe.TicTacToe)
		if s.LastAction() == 4 {
			fmt.Println("Found move at position:", s.LastAction())
			break
		}
	}
	// Output:
	// Found move at position: 4
}

func ExampleTicTacToe_ID() {
	game1 := tictactoe.NewTicTacToe()
	game2 := tictactoe.NewTicTacToe()

	// Same state produces the same ID
	fmt.Println("Same state, same ID:", game1.ID() == game2.ID())

	// Different state produces a different ID
	game2.Play(0)
	fmt.Println("Different state, same ID:", game1.ID() == game2.ID())
	// Output:
	// Same state, same ID: true
	// Different state, same ID: false
}

func ExampleTicTacToe_CurrentActor() {
	game := tictactoe.NewTicTacToe()
	fmt.Println("First actor:", game.CurrentActor())
	game.Play(0)
	fmt.Println("Second actor:", game.CurrentActor())
	game.Play(1)
	fmt.Println("Back to first:", game.CurrentActor())
	// Output:
	// First actor: 1
	// Second actor: 2
	// Back to first: 1
}

func ExampleTicTacToe_String() {
	game := tictactoe.NewTicTacToe()
	game.Play(4) // X at center
	game.Play(0) // O at top-left

	// String() returns a board with ANSI colors
	s := game.String()
	fmt.Println("Board contains grid:", len(s) > 0)
	// Output:
	// Board contains grid: true
}

func ExampleTicTacToe_RandomMove() {
	game := tictactoe.NewTicTacToe()
	game.Play(4) // X at center
	game.Play(0) // O at top-left

	// Choisir un coup aléatoire parmi les 7 cases vides
	rng := func(n int) int { return 0 } // toujours la première case vide
	next := game.RandomMove(rng)
	fmt.Println("Next actor:", next.CurrentActor())
	fmt.Println("Game still going:", next.Evaluate() == decision.Undecided)
	// Output:
	// Next actor: 2
	// Game still going: true
}

// Verify that TicTacToe implements Boarder (State + ActionRecorder).
var _ board.Boarder = (*tictactoe.TicTacToe)(nil)

// Verify that TicTacToe implements RandomMover.
var _ decision.RandomMover = (*tictactoe.TicTacToe)(nil)
