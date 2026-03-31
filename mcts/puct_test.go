package mcts

import (
	"math"
	"math/rand"
	"testing"

	"github.com/owulveryck/alphazego/board"
	"github.com/owulveryck/alphazego/board/tictactoe"
)

// --- PUCT Tests ---

func TestPUCT_UnvisitedWithParent(t *testing.T) {
	m := &MCTS{cpuct: 1.5}
	parent := &mctsNode{visits: 100, mcts: m}
	child := &mctsNode{visits: 0, prior: 0.3, parent: parent, mcts: m}

	score := child.puct()
	// C_puct * P * sqrt(N_parent) = 1.5 * 0.3 * sqrt(100) = 1.5 * 0.3 * 10 = 4.5
	expected := 1.5 * 0.3 * math.Sqrt(100)
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("expected PUCT ~ %f for unvisited node, got %f", expected, score)
	}
}

func TestPUCT_UnvisitedRoot(t *testing.T) {
	m := &MCTS{cpuct: 1.5}
	root := &mctsNode{visits: 0, prior: 0.5, parent: nil, mcts: m}

	score := root.puct()
	if math.Abs(score-0.5) > 1e-9 {
		t.Errorf("expected prior 0.5 for unvisited root, got %f", score)
	}
}

func TestPUCT_VisitedWithParent(t *testing.T) {
	m := &MCTS{cpuct: 2.0}
	parent := &mctsNode{visits: 100, mcts: m}
	child := &mctsNode{visits: 10, wins: 6, prior: 0.4, parent: parent, mcts: m}

	score := child.puct()
	q := 6.0 / 10.0
	exploration := 2.0 * 0.4 * math.Sqrt(100) / (1 + 10)
	expected := q + exploration
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("expected PUCT ~ %f, got %f", expected, score)
	}
}

func TestPUCT_VisitedRoot(t *testing.T) {
	m := &MCTS{cpuct: 1.0}
	root := &mctsNode{visits: 10, wins: 5, parent: nil, mcts: m}

	score := root.puct()
	if math.Abs(score-0.5) > 1e-9 {
		t.Errorf("expected 0.5 for visited root, got %f", score)
	}
}

func TestPUCT_HigherPriorGivesHigherScore(t *testing.T) {
	m := &MCTS{cpuct: 1.0}
	parent := &mctsNode{visits: 100, mcts: m}
	lowPrior := &mctsNode{visits: 0, prior: 0.1, parent: parent, mcts: m}
	highPrior := &mctsNode{visits: 0, prior: 0.9, parent: parent, mcts: m}

	if lowPrior.puct() >= highPrior.puct() {
		t.Error("expected higher prior to give higher PUCT score")
	}
}

// --- BackpropagateValue Tests ---

func TestBackpropagateValue_SignInversion(t *testing.T) {
	root := &mctsNode{state: tictactoe.NewTicTacToe()}
	ttt1 := tictactoe.NewTicTacToe()
	ttt1.Play(0)
	child := &mctsNode{state: ttt1, parent: root}
	ttt2 := tictactoe.NewTicTacToe()
	ttt2.Play(0)
	ttt2.Play(1)
	grandchild := &mctsNode{state: ttt2, parent: child}

	// value=0.8 from grandchild.CurrentPlayer()'s perspective
	// BackpropagateValue negates first to store from "player who moved here" perspective
	grandchild.backpropagateValue(0.8)

	// grandchild: initial negate → wins += -0.8 (from opponent's perspective = player who moved here)
	if math.Abs(grandchild.wins-(-0.8)) > 1e-9 {
		t.Errorf("expected grandchild wins=-0.8, got %f", grandchild.wins)
	}
	// child: wins += 0.8 (alternation)
	if math.Abs(child.wins-0.8) > 1e-9 {
		t.Errorf("expected child wins=0.8, got %f", child.wins)
	}
	// root: wins += -0.8 (alternation)
	if math.Abs(root.wins-(-0.8)) > 1e-9 {
		t.Errorf("expected root wins=-0.8, got %f", root.wins)
	}
}

func TestBackpropagateValue_UpdatesVisits(t *testing.T) {
	root := &mctsNode{state: tictactoe.NewTicTacToe()}
	ttt1 := tictactoe.NewTicTacToe()
	ttt1.Play(0)
	child := &mctsNode{state: ttt1, parent: root}

	child.backpropagateValue(1.0)

	if child.visits != 1 {
		t.Errorf("expected child visits=1, got %f", child.visits)
	}
	if root.visits != 1 {
		t.Errorf("expected root visits=1, got %f", root.visits)
	}
	// child: value negated first → wins = -1.0
	if math.Abs(child.wins-(-1.0)) > 1e-9 {
		t.Errorf("expected child wins=-1.0, got %f", child.wins)
	}
	// root: alternated → wins = 1.0
	if math.Abs(root.wins-1.0) > 1e-9 {
		t.Errorf("expected root wins=1.0, got %f", root.wins)
	}
}

// --- ExpandAll Tests ---

func TestExpandAll_CreatesAllChildren(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.getOrCreateNode(ttt, nil)

	policy := make([]float64, 9)
	for i := range policy {
		policy[i] = 1.0 / 9.0
	}
	node.expandAll(policy)

	if len(node.children) != 9 {
		t.Errorf("expected 9 children, got %d", len(node.children))
	}
}

func TestExpandAll_AssignsPriors(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.getOrCreateNode(ttt, nil)

	policy := []float64{0.5, 0.1, 0.1, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05}
	node.expandAll(policy)

	for i, child := range node.children {
		if math.Abs(child.prior-policy[i]) > 1e-9 {
			t.Errorf("child %d: expected prior %f, got %f", i, policy[i], child.prior)
		}
		if child.parent != node {
			t.Errorf("child %d: expected parent to be the expanded node", i)
		}
		if child.mcts != m {
			t.Errorf("child %d: expected mcts reference", i)
		}
	}
}

func TestExpandAll_IsFullyExpanded(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.getOrCreateNode(ttt, nil)

	policy := make([]float64, 9)
	for i := range policy {
		policy[i] = 1.0 / 9.0
	}
	node.expandAll(policy)

	if !node.isFullyExpanded() {
		t.Error("expected fully expanded after ExpandAll")
	}
}

// --- SelectChildUCB with PUCT dispatch Tests ---

func TestSelectChildUCB_UsesPUCTWithEvaluator(t *testing.T) {
	eval := &uniformEvaluator{}
	m := NewAlphaMCTS(eval, 1.0)
	ttt := tictactoe.NewTicTacToe()
	node := m.getOrCreateNode(ttt, nil)
	node.visits = 100

	// Create two children with different priors
	ttt1 := tictactoe.NewTicTacToe()
	ttt1.Play(0)
	child1 := &mctsNode{state: ttt1, parent: node, prior: 0.1, visits: 0, mcts: m}
	ttt2 := tictactoe.NewTicTacToe()
	ttt2.Play(4)
	child2 := &mctsNode{state: ttt2, parent: node, prior: 0.9, visits: 0, mcts: m}
	node.children = []*mctsNode{child1, child2}

	best := node.selectChildUCB()
	if best != child2 {
		t.Error("expected child with higher prior to be selected with PUCT")
	}
}

// --- NewAlphaMCTS Tests ---

func TestNewAlphaMCTS(t *testing.T) {
	eval := &uniformEvaluator{}
	m := NewAlphaMCTS(eval, 1.5)
	if m == nil {
		t.Fatal("expected non-nil MCTS")
	}
	if m.evaluator != eval {
		t.Error("expected evaluator to be set")
	}
	if m.cpuct != 1.5 {
		t.Errorf("expected cpuct=1.5, got %f", m.cpuct)
	}
}

// --- RunMCTS with Evaluator Integration Tests ---

func TestRunMCTS_WithEvaluator_ReturnsValidState(t *testing.T) {
	eval := &uniformEvaluator{}
	m := NewAlphaMCTS(eval, 1.0)
	ttt := tictactoe.NewTicTacToe()

	result := m.RunMCTS(ttt, 100)
	if result == nil {
		t.Fatal("expected non-nil result state")
	}
	if result.CurrentPlayer() != board.Player2 {
		t.Errorf("expected Player2's turn after MCTS move, got %d", result.CurrentPlayer())
	}
}

func TestRunMCTS_WithEvaluator_BlocksWin(t *testing.T) {
	eval := &rolloutEvaluator{}
	m := NewAlphaMCTS(eval, 1.0)
	// Player2's turn. Player1 has positions 0,1 - about to win at 2.
	ttt := playMoves(0, 3, 1, 4)

	result := m.RunMCTS(ttt, 5000)

	move := result.LastMove()
	if move != 2 {
		t.Errorf("expected AlphaMCTS to block at position 2, got move %d", move)
	}
}

func TestRunMCTS_WithEvaluator_TakesWin(t *testing.T) {
	eval := &rolloutEvaluator{}
	m := NewAlphaMCTS(eval, 1.0)
	// Player1's turn, can win at position 2
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0) // P1
	ttt.Play(3) // P2
	ttt.Play(1) // P1
	ttt.Play(7) // P2

	result := m.RunMCTS(ttt, 5000)

	move := result.LastMove()
	if move != 2 {
		t.Errorf("expected AlphaMCTS to win at position 2, got move %d", move)
	}
}

// --- terminalValue Tests ---

func TestTerminalValue_Player1Wins(t *testing.T) {
	// Player1 wins, it's Player2's turn (meaning Player1 just moved)
	ttt := playMoves(0, 3, 1, 4, 2) // P1 wins top row
	v := terminalValue(ttt)
	// CurrentPlayer = Player2, playerWhoMovedHere = Player1, result = Player1Wins
	// → -1.0 (defaite pour le joueur courant)
	if math.Abs(v-(-1.0)) > 1e-9 {
		t.Errorf("expected -1.0, got %f", v)
	}
}

func TestTerminalValue_Draw(t *testing.T) {
	ttt := playMoves(4, 0, 2, 6, 3, 5, 1, 7, 8)
	if ttt.Evaluate() != board.DrawResult {
		t.Skipf("sequence didn't produce a draw, got %d", ttt.Evaluate())
	}
	v := terminalValue(ttt)
	if math.Abs(v) > 1e-9 {
		t.Errorf("expected 0.0 for draw, got %f", v)
	}
}

// --- Test helpers ---

// uniformEvaluator retourne une policy uniforme et une value neutre.
// Utile pour tester que le chemin AlphaZero fonctionne sans signal fort.
type uniformEvaluator struct{}

func (u *uniformEvaluator) Evaluate(state board.State) ([]float64, float64) {
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

// rolloutEvaluator retourne une policy uniforme et une value estimee
// par rollout aleatoire. Cela fournit un signal reel pour les tests tactiques.
type rolloutEvaluator struct{}

func (r *rolloutEvaluator) Evaluate(state board.State) ([]float64, float64) {
	moves := state.PossibleMoves()
	n := len(moves)
	if n == 0 {
		return nil, 0.0
	}
	policy := make([]float64, n)
	for i := range policy {
		policy[i] = 1.0 / float64(n)
	}
	// Perform a random rollout to estimate value
	currentState := state
	for currentState.Evaluate() == board.NoPlayer {
		possibleMoves := currentState.PossibleMoves()
		currentState = possibleMoves[rand.Intn(len(possibleMoves))]
	}
	result := currentState.Evaluate()
	current := state.CurrentPlayer()
	if result == current {
		return policy, 1.0
	}
	if result == board.DrawResult {
		return policy, 0.0
	}
	return policy, -1.0
}
