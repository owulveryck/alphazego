package tictactoe_test

import (
	"fmt"

	"github.com/owulveryck/alphazego/board"
	"github.com/owulveryck/alphazego/board/tictactoe"
)

func ExampleNewTicTacToe() {
	game := tictactoe.NewTicTacToe()
	fmt.Println("Current player:", game.CurrentPlayer())
	fmt.Println("Game result:", game.Evaluate())
	// Output:
	// Current player: 1
	// Game result: 0
}

func ExampleTicTacToe_Play() {
	game := tictactoe.NewTicTacToe()

	// Player1 plays at center (position 4)
	game.Play(4)
	fmt.Println("After Player1 plays at 4, current player:", game.CurrentPlayer())

	// Player2 plays at top-left (position 0)
	game.Play(0)
	fmt.Println("After Player2 plays at 0, current player:", game.CurrentPlayer())
	// Output:
	// After Player1 plays at 4, current player: 2
	// After Player2 plays at 0, current player: 1
}

func ExampleTicTacToe_Evaluate_gameOn() {
	game := tictactoe.NewTicTacToe()
	game.Play(0) // X at top-left
	game.Play(4) // O at center

	result := game.Evaluate()
	fmt.Println("Game still in progress:", result == board.NoPlayer)
	// Output:
	// Game still in progress: true
}

func ExampleTicTacToe_Evaluate_player1Wins() {
	game := tictactoe.NewTicTacToe()
	// Player1 takes the top row: 0, 1, 2
	game.Play(0) // X
	game.Play(3) // O
	game.Play(1) // X
	game.Play(4) // O
	game.Play(2) // X wins!

	result := game.Evaluate()
	fmt.Println("Player1 wins:", result == board.Player1)
	// Output:
	// Player1 wins: true
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
	fmt.Println("Draw:", result == board.DrawResult)
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
	game := tictactoe.NewTicTacToe() // Player1's turn
	moves := game.PossibleMoves()

	// Each child state has the other player's turn
	for _, m := range moves {
		if m.CurrentPlayer() != board.Player2 {
			fmt.Println("ERROR: expected Player2's turn in child state")
			return
		}
	}
	fmt.Println("All child states have Player2's turn: true")
	// Output:
	// All child states have Player2's turn: true
}

func ExampleTicTacToe_PreviousPlayer() {
	game := tictactoe.NewTicTacToe()
	// Sur un plateau vierge, le joueur courant est Player1 ;
	// PreviousPlayer retourne Player2 (le "dernier" dans l'alternance).
	fmt.Println("Previous player (initial):", game.PreviousPlayer())

	game.Play(4) // Player1 joue au centre
	fmt.Println("Previous player after P1 plays:", game.PreviousPlayer())

	game.Play(0) // Player2 joue en haut a gauche
	fmt.Println("Previous player after P2 plays:", game.PreviousPlayer())
	// Output:
	// Previous player (initial): 2
	// Previous player after P1 plays: 1
	// Previous player after P2 plays: 2
}

func ExampleTicTacToe_LastMove() {
	game := tictactoe.NewTicTacToe()
	game.Play(0) // Player1 at position 0
	// Now it's Player2's turn. Each possible next state knows its LastMove.

	moves := game.PossibleMoves()
	// Find the state where position 4 was played
	for _, next := range moves {
		if next.LastMove() == 4 {
			fmt.Println("Found move at position:", next.LastMove())
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

func ExampleTicTacToe_CurrentPlayer() {
	game := tictactoe.NewTicTacToe()
	fmt.Println("First player:", game.CurrentPlayer())
	game.Play(0)
	fmt.Println("Second player:", game.CurrentPlayer())
	game.Play(1)
	fmt.Println("Back to first:", game.CurrentPlayer())
	// Output:
	// First player: 1
	// Second player: 2
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
