package mcts

import (
	"math"
	"math/rand"
	"testing"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/board"
	"github.com/owulveryck/alphazego/decision/board/samples/tictactoe"
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

func TestBackpropagateValue_TwoActors(t *testing.T) {
	root := testNode(tictactoe.NewTicTacToe(), nil)
	ttt1 := tictactoe.NewTicTacToe()
	ttt1.Play(0)
	child := testNode(ttt1, root)
	ttt2 := tictactoe.NewTicTacToe()
	ttt2.Play(0)
	ttt2.Play(1)
	grandchild := testNode(ttt2, child)

	// values map: Cross=0.8 (favorable), Circle=-0.8 (défavorable)
	values := map[decision.ActorID]float64{
		tictactoe.Cross:  0.8,
		tictactoe.Circle: -0.8,
	}
	grandchild.backpropagateValue(values)

	// grandchild: PreviousActor = Circle → wins += -0.8
	if math.Abs(grandchild.wins-(-0.8)) > 1e-9 {
		t.Errorf("expected grandchild wins=-0.8, got %f", grandchild.wins)
	}
	// child: PreviousActor = Cross → wins += 0.8
	if math.Abs(child.wins-0.8) > 1e-9 {
		t.Errorf("expected child wins=0.8, got %f", child.wins)
	}
	// root: PreviousActor = Circle → wins += -0.8
	if math.Abs(root.wins-(-0.8)) > 1e-9 {
		t.Errorf("expected root wins=-0.8, got %f", root.wins)
	}
}

func TestBackpropagateValue_UpdatesVisits(t *testing.T) {
	root := testNode(tictactoe.NewTicTacToe(), nil)
	ttt1 := tictactoe.NewTicTacToe()
	ttt1.Play(0)
	child := testNode(ttt1, root)

	values := map[decision.ActorID]float64{
		tictactoe.Cross:  1.0,
		tictactoe.Circle: -1.0,
	}
	child.backpropagateValue(values)

	if child.visits != 1 {
		t.Errorf("expected child visits=1, got %f", child.visits)
	}
	if root.visits != 1 {
		t.Errorf("expected root visits=1, got %f", root.visits)
	}
	// child: PreviousActor = Cross → wins = 1.0
	if math.Abs(child.wins-1.0) > 1e-9 {
		t.Errorf("expected child wins=1.0, got %f", child.wins)
	}
	// root: PreviousActor = Circle → wins = -1.0
	if math.Abs(root.wins-(-1.0)) > 1e-9 {
		t.Errorf("expected root wins=-1.0, got %f", root.wins)
	}
}

// --- ExpandAll Tests ---

func TestExpandAll_CreatesAllChildren(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.newNode(ttt, nil)

	policy := make([]float64, 9)
	for i := range policy {
		policy[i] = 1.0 / 9.0
	}
	if err := node.expandAll(policy); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(node.children) != 9 {
		t.Errorf("expected 9 children, got %d", len(node.children))
	}
}

func TestExpandAll_AssignsPriors(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.newNode(ttt, nil)

	policy := []float64{0.5, 0.1, 0.1, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05}
	if err := node.expandAll(policy); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	node := m.newNode(ttt, nil)

	policy := make([]float64, 9)
	for i := range policy {
		policy[i] = 1.0 / 9.0
	}
	if err := node.expandAll(policy); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !node.isFullyExpanded() {
		t.Error("expected fully expanded after ExpandAll")
	}
}

// --- SelectChildUCB with PUCT dispatch Tests ---

func TestSelectChildUCB_UsesPUCTWithEvaluator(t *testing.T) {
	eval := &uniformEvaluator{}
	m := NewAlphaMCTS(eval, 1.0)
	ttt := tictactoe.NewTicTacToe()
	node := m.newNode(ttt, nil)
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
	if result.CurrentActor() != tictactoe.Circle {
		t.Errorf("expected Actor2's turn after MCTS move, got %d", result.CurrentActor())
	}
}

func TestRunMCTS_WithEvaluator_BlocksWin(t *testing.T) {
	eval := &rolloutEvaluator{}
	m := NewAlphaMCTS(eval, 1.0)
	// Actor2's turn. Actor1 has positions 0,1 - about to win at 2.
	ttt := playMoves(0, 3, 1, 4)

	result := m.RunMCTS(ttt, 5000)

	move := result.(board.ActionRecorder).LastAction()
	if move != 2 {
		t.Errorf("expected AlphaMCTS to block at position 2, got move %d", move)
	}
}

func TestRunMCTS_WithEvaluator_TakesWin(t *testing.T) {
	eval := &rolloutEvaluator{}
	m := NewAlphaMCTS(eval, 1.0)
	// Actor1's turn, can win at position 2
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0) // A1
	ttt.Play(3) // A2
	ttt.Play(1) // A1
	ttt.Play(7) // A2

	result := m.RunMCTS(ttt, 5000)

	move := result.(board.ActionRecorder).LastAction()
	if move != 2 {
		t.Errorf("expected AlphaMCTS to win at position 2, got move %d", move)
	}
}

// --- backpropagateTerminal Tests ---

func TestBackpropagateTerminal_Actor1Wins(t *testing.T) {
	// Actor1 wins, it's Actor2's turn (meaning Actor1 just moved)
	ttt := playMoves(0, 3, 1, 4, 2) // A1 wins top row
	root := testNode(tictactoe.NewTicTacToe(), nil)
	node := testNode(ttt, root)

	node.backpropagateTerminal()

	// node: PreviousActor = Cross (Actor1), result = Cross → wins = 1.0
	if math.Abs(node.wins-1.0) > 1e-9 {
		t.Errorf("expected node wins=1.0 (Cross won, Cross moved here), got %f", node.wins)
	}
	// root: PreviousActor = Circle (Actor2), result = Cross → wins = -1.0
	if math.Abs(root.wins-(-1.0)) > 1e-9 {
		t.Errorf("expected root wins=-1.0 (Cross won, Circle moved here), got %f", root.wins)
	}
}

func TestBackpropagateTerminal_Draw(t *testing.T) {
	ttt := playMoves(4, 0, 2, 6, 3, 5, 1, 7, 8)
	if ttt.Evaluate() != decision.Stalemate {
		t.Skipf("sequence didn't produce a draw, got %d", ttt.Evaluate())
	}
	node := testNode(ttt, nil)
	node.backpropagateTerminal()

	// Stalemate : wins += 0.0
	if math.Abs(node.wins) > 1e-9 {
		t.Errorf("expected 0.0 for draw, got %f", node.wins)
	}
}

// --- Test helpers ---

// uniformEvaluator retourne une policy uniforme et des values neutres.
// Utile pour tester que le chemin AlphaZero fonctionne sans signal fort.
type uniformEvaluator struct{}

func (u *uniformEvaluator) Evaluate(state decision.State) ([]float64, map[decision.ActorID]float64) {
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

// rolloutEvaluator retourne une policy uniforme et des values estimées
// par rollout aléatoire. Cela fournit un signal réel pour les tests tactiques.
type rolloutEvaluator struct{}

func (r *rolloutEvaluator) Evaluate(state decision.State) ([]float64, map[decision.ActorID]float64) {
	moves := state.PossibleMoves()
	n := len(moves)
	if n == 0 {
		return nil, map[decision.ActorID]float64{}
	}
	policy := make([]float64, n)
	for i := range policy {
		policy[i] = 1.0 / float64(n)
	}
	// Perform a random rollout to estimate value
	currentState := state
	for currentState.Evaluate() == decision.Undecided {
		possibleMoves := currentState.PossibleMoves()
		currentState = possibleMoves[rand.Intn(len(possibleMoves))]
	}
	result := currentState.Evaluate()
	current := state.CurrentActor()
	previous := state.PreviousActor()
	values := make(map[decision.ActorID]float64)
	for _, actor := range []decision.ActorID{current, previous} {
		if result == actor {
			values[actor] = 1.0
		} else if result == decision.Stalemate {
			values[actor] = 0.0
		} else {
			values[actor] = -1.0
		}
	}
	return policy, values
}

// --- Edge case Tests ---

func TestSelectChildUCB_NoChildren_AlphaZero(t *testing.T) {
	eval := &uniformEvaluator{}
	m := NewAlphaMCTS(eval, 1.0)
	node := &mctsNode{children: []*mctsNode{}, mcts: m}
	if node.selectChildUCB() != nil {
		t.Error("expected nil for node with no children in AlphaZero mode")
	}
}

func TestExpandAll_EmptyPolicy(t *testing.T) {
	m := NewMCTS()
	// État terminal : PossibleMoves() retourne nil, donc expandAll avec policy vide.
	ttt := playMoves(0, 3, 1, 4, 2) // terminal
	node := m.newNode(ttt, nil)
	err := node.expandAll([]float64{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(node.children) != 0 {
		t.Errorf("expected 0 children, got %d", len(node.children))
	}
}

func TestPUCT_NegativePrior(t *testing.T) {
	m := &MCTS{cpuct: 1.5}
	parent := &mctsNode{visits: 100, mcts: m}
	child := &mctsNode{visits: 0, prior: -0.5, parent: parent, mcts: m}

	score := child.puct()
	// C_puct * P * sqrt(N_parent) = 1.5 * (-0.5) * 10 = -7.5
	expected := 1.5 * (-0.5) * math.Sqrt(100)
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("expected PUCT ~ %f for negative prior, got %f", expected, score)
	}
}

func TestPUCT_NegativePrior_Visited(t *testing.T) {
	m := &MCTS{cpuct: 2.0}
	parent := &mctsNode{visits: 100, mcts: m}
	child := &mctsNode{visits: 10, wins: 6, prior: -0.3, parent: parent, mcts: m}

	score := child.puct()
	q := 6.0 / 10.0
	exploration := 2.0 * (-0.3) * math.Sqrt(100) / (1 + 10)
	expected := q + exploration
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("expected PUCT ~ %f, got %f", expected, score)
	}
}

// TestExpandAll_PolicyWithNegativeValues vérifie que expandAll accepte une
// policy contenant des valeurs négatives (l'évaluateur peut en produire).
func TestExpandAll_PolicyWithNegativeValues(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.newNode(ttt, nil)
	policy := make([]float64, 9)
	for i := range policy {
		policy[i] = -0.5
	}
	// expandAll ne valide pas les valeurs individuelles, seulement la taille
	err := node.expandAll(policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(node.children) != 9 {
		t.Errorf("got %d children, want 9", len(node.children))
	}
}

// TestExpandAll_PolicyNotNormalized vérifie que expandAll log un warning
// quand la policy ne somme pas à ~1.0 mais ne retourne pas d'erreur.
func TestExpandAll_PolicyNotNormalized(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.newNode(ttt, nil)
	// Policy qui somme à 9.0 (chaque élément = 1.0)
	policy := make([]float64, 9)
	for i := range policy {
		policy[i] = 1.0
	}
	err := node.expandAll(policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(node.children) != 9 {
		t.Errorf("got %d children, want 9", len(node.children))
	}
}

// TestExpandAll_PolicyLengthMismatch vérifie que expandAll retourne une erreur
// lorsque la taille de la policy ne correspond pas aux coups possibles.
func TestExpandAll_PolicyLengthMismatch(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.newNode(ttt, nil)
	// Le morpion a 9 coups possibles au départ, on en fournit seulement 3
	err := node.expandAll([]float64{0.3, 0.3, 0.4})
	if err == nil {
		t.Fatal("expected error for policy length mismatch")
	}
}

// TestExpandAll_NilPolicy vérifie que expandAll retourne une erreur
// lorsque la policy est nil.
func TestExpandAll_NilPolicy(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.newNode(ttt, nil)
	err := node.expandAll(nil)
	if err == nil {
		t.Fatal("expected error for nil policy")
	}
}
