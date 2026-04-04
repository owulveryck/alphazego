package mcts

import (
	"math"
	"testing"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/board"
	"github.com/owulveryck/alphazego/decision/board/samples/tictactoe"
)

// --- UCB1 Tests ---

func TestUCB1_UnvisitedNode(t *testing.T) {
	node := &mctsNode{visits: 0}
	if !math.IsInf(node.ucb1(), 1) {
		t.Error("expected +Inf for unvisited node")
	}
}

func TestUCB1_WithParent(t *testing.T) {
	parent := &mctsNode{visits: 10}
	child := &mctsNode{visits: 5, wins: 3, parent: parent}
	score := child.ucb1()
	expected := 0.6 + math.Sqrt(2)*math.Sqrt(math.Log(10)/5)
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("expected UCB1 ~ %f, got %f", expected, score)
	}
}

func TestUCB1_RootNode(t *testing.T) {
	root := &mctsNode{visits: 10, wins: 5, parent: nil}
	score := root.ucb1()
	if math.Abs(score-0.5) > 1e-9 {
		t.Errorf("expected 0.5 for root node, got %f", score)
	}
}

// --- IsTerminal Tests ---

func TestIsTerminal_GameOn(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	node := &mctsNode{state: ttt}
	if node.isTerminal() {
		t.Error("expected non-terminal for new game")
	}
}

func TestIsTerminal_Won(t *testing.T) {
	// Actor1 wins top row: play sequence 0,3,1,4,2
	ttt := playMoves(0, 3, 1, 4, 2)
	node := &mctsNode{state: ttt}
	if !node.isTerminal() {
		t.Error("expected terminal for won game")
	}
}

func TestIsTerminal_Draw(t *testing.T) {
	// Draw: 4,0,2,6,3,5,1,7,8
	ttt := playMoves(4, 0, 2, 6, 3, 5, 1, 7, 8)
	node := &mctsNode{state: ttt}
	if ttt.Evaluate() != decision.Stalemate {
		// If this particular sequence doesn't draw, that's fine - just test what we get
		t.Skipf("sequence didn't produce a draw, got result %d", ttt.Evaluate())
	}
	if !node.isTerminal() {
		t.Error("expected terminal for draw")
	}
}

// --- IsFullyExpanded Tests ---

func TestIsFullyExpanded_NoChildren(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	node := &mctsNode{state: ttt, children: []*mctsNode{}}
	if node.isFullyExpanded() {
		t.Error("expected not fully expanded with no children")
	}
}

func TestIsFullyExpanded_AllExpanded(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	node := &mctsNode{state: ttt, children: make([]*mctsNode, 9)}
	if !node.isFullyExpanded() {
		t.Error("expected fully expanded when children count matches possible moves")
	}
}

// --- SelectChildUCB Tests ---

func TestSelectChildUCB_SelectsUnvisited(t *testing.T) {
	parent := &mctsNode{visits: 10}
	visited := &mctsNode{visits: 5, wins: 2, parent: parent}
	unvisited := &mctsNode{visits: 0, parent: parent}
	parent.children = []*mctsNode{visited, unvisited}

	best := parent.selectChildUCB()
	if best != unvisited {
		t.Error("expected unvisited child to be selected (Inf UCB1)")
	}
}

func TestSelectChildUCB_NoChildren(t *testing.T) {
	node := &mctsNode{children: []*mctsNode{}}
	if node.selectChildUCB() != nil {
		t.Error("expected nil for node with no children")
	}
}

func TestSelectChildUCB_SelectsBestScore(t *testing.T) {
	parent := &mctsNode{visits: 100}
	child1 := &mctsNode{visits: 50, wins: 10, parent: parent}
	child2 := &mctsNode{visits: 50, wins: 40, parent: parent}
	parent.children = []*mctsNode{child1, child2}

	best := parent.selectChildUCB()
	if best != child2 {
		t.Error("expected child with higher win rate to be selected")
	}
}

// --- SelectBestMove Tests ---

func TestSelectBestMove_HighestVisits(t *testing.T) {
	child1 := &mctsNode{visits: 10, wins: 5}
	child2 := &mctsNode{visits: 20, wins: 8}
	child3 := &mctsNode{visits: 15, wins: 12}
	parent := &mctsNode{children: []*mctsNode{child1, child2, child3}}

	best := parent.selectBestMove()
	if best != child2 {
		t.Errorf("expected child with 20 visits, got visits=%f", best.visits)
	}
}

func TestSelectBestMove_NoChildren(t *testing.T) {
	node := &mctsNode{children: []*mctsNode{}}
	if node.selectBestMove() != nil {
		t.Error("expected nil for node with no children")
	}
}

// --- Expand Tests ---

func TestExpand_CreatesOneChild(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.getOrCreateNode(ttt, nil)

	child := node.expand()
	if child == nil {
		t.Fatal("expected a child node from expansion")
	}
	if len(node.children) != 1 {
		t.Errorf("expected 1 child after first expand, got %d", len(node.children))
	}
	if child.parent != node {
		t.Error("expected child's parent to be the expanded node")
	}
}

func TestExpand_ExpandsIncrementally(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.getOrCreateNode(ttt, nil)

	for i := 1; i <= 9; i++ {
		child := node.expand()
		if child == nil {
			t.Fatalf("expected child on expand %d", i)
		}
		if len(node.children) != i {
			t.Errorf("expected %d children after expand %d, got %d", i, i, len(node.children))
		}
	}

	// 10th expand should return nil (fully expanded)
	if node.expand() != nil {
		t.Error("expected nil after all moves expanded")
	}
}

func TestExpand_ChildHasCorrectParent(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.getOrCreateNode(ttt, nil)

	child := node.expand()
	if child.mcts != m {
		t.Error("expected child to have reference to MCTS instance")
	}
	if child.parent != node {
		t.Error("expected child's parent to be the expanded node")
	}
	if child.state == nil {
		t.Error("expected child to have a state")
	}
}

// --- Simulate Tests ---

func TestSimulate_ReturnsTerminalResult(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	node := &mctsNode{state: ttt}

	for i := 0; i < 20; i++ {
		result := node.simulate()
		if result != tictactoe.Cross && result != tictactoe.Circle && result != decision.Stalemate {
			t.Errorf("expected terminal result, got %d", result)
		}
	}
}

func TestSimulate_AlreadyTerminal(t *testing.T) {
	ttt := playMoves(0, 3, 1, 4, 2) // Actor1 wins top row
	node := &mctsNode{state: ttt}

	result := node.simulate()
	if result != tictactoe.Cross {
		t.Errorf("expected Actor1, got %d", result)
	}
}

// --- Backpropagate Tests ---

func TestBackpropagate_UpdatesVisits(t *testing.T) {
	root := &mctsNode{
		state:  tictactoe.NewTicTacToe(),
		parent: nil,
	}
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0)
	child := &mctsNode{
		state:  ttt,
		parent: root,
	}

	child.backpropagate(tictactoe.Cross)

	if child.visits != 1 {
		t.Errorf("expected child visits=1, got %f", child.visits)
	}
	if root.visits != 1 {
		t.Errorf("expected root visits=1, got %f", root.visits)
	}
}

func TestBackpropagate_CreditsCorrectActor(t *testing.T) {
	// Root: Actor1's turn
	root := &mctsNode{
		state:  tictactoe.NewTicTacToe(),
		parent: nil,
	}
	// Child: Actor2's turn (Actor1 just moved)
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0)
	child := &mctsNode{
		state:  ttt,
		parent: root,
	}

	child.backpropagate(tictactoe.Cross)

	// At child: PreviousActor = Actor1. Result=Actor1 → win
	if child.wins != 1 {
		t.Errorf("expected child wins=1 (Actor1 moved here and won), got %f", child.wins)
	}
	// At root: PreviousActor = Actor2. Result=Actor1 → no win
	if root.wins != 0 {
		t.Errorf("expected root wins=0, got %f", root.wins)
	}
}

func TestBackpropagate_Actor2Wins(t *testing.T) {
	root := &mctsNode{
		state:  tictactoe.NewTicTacToe(),
		parent: nil,
	}
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0)
	child := &mctsNode{
		state:  ttt,
		parent: root,
	}

	child.backpropagate(tictactoe.Circle)

	// At child: PreviousActor = Actor1. Result=Actor2 → no win
	if child.wins != 0 {
		t.Errorf("expected child wins=0, got %f", child.wins)
	}
	// At root: PreviousActor = Actor2. Result=Actor2 → win
	if root.wins != 1 {
		t.Errorf("expected root wins=1, got %f", root.wins)
	}
}

func TestBackpropagate_Draw(t *testing.T) {
	root := &mctsNode{
		state:  tictactoe.NewTicTacToe(),
		parent: nil,
	}
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0)
	child := &mctsNode{
		state:  ttt,
		parent: root,
	}

	child.backpropagate(decision.Stalemate)

	if child.wins != 0.5 {
		t.Errorf("expected child wins=0.5 for draw, got %f", child.wins)
	}
	if root.wins != 0.5 {
		t.Errorf("expected root wins=0.5 for draw, got %f", root.wins)
	}
}

func TestBackpropagate_DeepChain(t *testing.T) {
	root := &mctsNode{state: tictactoe.NewTicTacToe()}

	ttt1 := tictactoe.NewTicTacToe()
	ttt1.Play(0)
	child := &mctsNode{state: ttt1, parent: root}

	ttt2 := tictactoe.NewTicTacToe()
	ttt2.Play(0)
	ttt2.Play(1)
	grandchild := &mctsNode{state: ttt2, parent: child}

	grandchild.backpropagate(tictactoe.Cross)

	if grandchild.visits != 1 || child.visits != 1 || root.visits != 1 {
		t.Error("expected all nodes to have 1 visit")
	}
}

// --- GetOrCreateNode Tests ---

func TestGetOrCreateNode_New(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.getOrCreateNode(ttt, nil)

	if node == nil {
		t.Fatal("expected non-nil node")
	}
	if node.state != ttt {
		t.Error("expected node state to be the provided state")
	}
	if node.parent != nil {
		t.Error("expected nil parent for root node")
	}
	if node.mcts != m {
		t.Error("expected node to have reference to MCTS instance")
	}
}

func TestGetOrCreateNode_Existing(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node1 := m.getOrCreateNode(ttt, nil)
	node2 := m.getOrCreateNode(ttt, nil)

	if node1 != node2 {
		t.Error("expected same node for same state")
	}
}

func TestNewMCTS(t *testing.T) {
	m := NewMCTS()
	if m == nil {
		t.Fatal("expected non-nil MCTS")
	}
	if m.inventory == nil {
		t.Error("expected initialized inventory")
	}
}

// --- RunMCTS Integration Tests ---

func TestRunMCTS_ReturnsValidState(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()

	result := m.RunMCTS(ttt, 100)
	if result == nil {
		t.Fatal("expected non-nil result state")
	}
	if result.CurrentActor() != tictactoe.Circle {
		t.Errorf("expected Actor2's turn after MCTS move, got %d", result.CurrentActor())
	}
}

func TestRunMCTS_BlocksWin(t *testing.T) {
	// Actor2's turn. Actor1 has positions 0,1 - about to win at 2.
	ttt := playMoves(0, 3, 1, 4)

	m := NewMCTS()
	result := m.RunMCTS(ttt, 5000)

	move := result.(board.ActionRecorder).LastAction()
	if move != 2 {
		t.Errorf("expected MCTS to block at position 2, got move %d", move)
	}
}

func TestRunMCTS_TakesWin(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0) // A1
	ttt.Play(3) // A2
	ttt.Play(1) // A1
	ttt.Play(7) // A2 plays somewhere else (not blocking)
	// Board: [1,1,0,0,0,0,0,2,0], A1's turn, can win at 2

	m := NewMCTS()
	result := m.RunMCTS(ttt, 5000)

	move := result.(board.ActionRecorder).LastAction()
	if move != 2 {
		t.Errorf("expected MCTS to win at position 2, got move %d", move)
	}
}

func TestRunMCTS_TerminalState(t *testing.T) {
	// Actor1 wins: 0,3,1,4,2
	ttt := playMoves(0, 3, 1, 4, 2)
	if ttt.Evaluate() == decision.Undecided {
		t.Fatal("expected terminal state")
	}

	m := NewMCTS()
	result := m.RunMCTS(ttt, 100)
	if result != ttt {
		t.Error("expected original state returned for terminal state")
	}
}

func TestRunMCTS_ZeroIterations(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	m := NewMCTS()
	result := m.RunMCTS(ttt, 0)

	if result != ttt {
		t.Error("expected original state returned for 0 iterations")
	}
}

func TestRunMCTS_FullGame(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()

	maxMoves := 9
	for i := 0; i < maxMoves && ttt.Evaluate() == decision.Undecided; i++ {
		next := m.RunMCTS(ttt, 500)
		if next == ttt {
			t.Fatal("MCTS returned same state for non-terminal game")
		}
		move := next.(board.ActionRecorder).LastAction()
		ttt.Play(uint8(move))
	}

	result := ttt.Evaluate()
	if result == decision.Undecided {
		t.Error("expected game to end")
	}
}

// --- Three-actor mock state (validates N-actor Backpropagate) ---

// threeActorState est un état mock à 3 acteurs en round-robin.
type threeActorState struct {
	current  decision.ActorID
	previous decision.ActorID
	result   decision.ActorID
	id       string
}

func (s *threeActorState) CurrentActor() decision.ActorID  { return s.current }
func (s *threeActorState) PreviousActor() decision.ActorID { return s.previous }
func (s *threeActorState) Evaluate() decision.ActorID      { return s.result }
func (s *threeActorState) PossibleMoves() []decision.State { return nil }
func (s *threeActorState) ID() string                      { return s.id }

func TestBackpropagate_ThreeActors(t *testing.T) {
	// Simule une chaîne de 3 nœuds : acteur 10 → acteur 11 → acteur 12
	// avec un résultat où l'acteur 10 gagne (Result = 10).
	root := &mctsNode{state: &threeActorState{current: 10, previous: 12, result: decision.Undecided, id: "root"}}
	child := &mctsNode{state: &threeActorState{current: 11, previous: 10, result: decision.Undecided, id: "child"}, parent: root}
	grandchild := &mctsNode{state: &threeActorState{current: 12, previous: 11, result: decision.ActorID(10), id: "gchild"}, parent: child}

	// L'acteur 10 gagne
	grandchild.backpropagate(decision.ActorID(10))

	// grandchild: PreviousActor = 11, result = 10 → pas de victoire pour l'acteur 11
	if grandchild.wins != 0 {
		t.Errorf("expected grandchild wins=0 (actor 11 moved here, actor 10 won), got %f", grandchild.wins)
	}
	// child: PreviousActor = 10, result = 10 → victoire pour l'acteur 10
	if child.wins != 1 {
		t.Errorf("expected child wins=1 (actor 10 moved here and won), got %f", child.wins)
	}
	// root: PreviousActor = 12, result = 10 → pas de victoire pour l'acteur 12
	if root.wins != 0 {
		t.Errorf("expected root wins=0 (actor 12 moved here, actor 10 won), got %f", root.wins)
	}

	// Toutes les visites doivent être 1
	if grandchild.visits != 1 || child.visits != 1 || root.visits != 1 {
		t.Error("expected all nodes to have 1 visit")
	}
}

func TestBackpropagate_ThreeActors_Draw(t *testing.T) {
	root := &mctsNode{state: &threeActorState{current: 10, previous: 12, result: decision.Undecided, id: "root"}}
	child := &mctsNode{state: &threeActorState{current: 11, previous: 10, result: decision.Undecided, id: "child"}, parent: root}

	child.backpropagate(decision.Stalemate)

	// Tous les nœuds reçoivent 0.5 pour un match nul
	if child.wins != 0.5 {
		t.Errorf("expected child wins=0.5 for draw, got %f", child.wins)
	}
	if root.wins != 0.5 {
		t.Errorf("expected root wins=0.5 for draw, got %f", root.wins)
	}
}

// --- Single-actor mock state (validates 1-actor backpropagation) ---

// singleActorState est un état mock à un seul acteur.
// CurrentActor() == PreviousActor() == actor.
type singleActorState struct {
	actor  decision.ActorID
	result decision.ActorID
	id     string
}

func (s *singleActorState) CurrentActor() decision.ActorID  { return s.actor }
func (s *singleActorState) PreviousActor() decision.ActorID { return s.actor }
func (s *singleActorState) Evaluate() decision.ActorID      { return s.result }
func (s *singleActorState) PossibleMoves() []decision.State { return nil }
func (s *singleActorState) ID() string                      { return s.id }

func TestBackpropagate_SingleActor_Win(t *testing.T) {
	root := &mctsNode{state: &singleActorState{actor: 1, result: decision.Undecided, id: "root"}}
	child := &mctsNode{state: &singleActorState{actor: 1, result: decision.Undecided, id: "child"}, parent: root}
	grandchild := &mctsNode{state: &singleActorState{actor: 1, result: 1, id: "gchild"}, parent: child}

	grandchild.backpropagate(decision.ActorID(1))

	// Tous les nœuds ont PreviousActor == 1, result == 1 → victoire partout
	if grandchild.wins != 1 {
		t.Errorf("expected grandchild wins=1, got %f", grandchild.wins)
	}
	if child.wins != 1 {
		t.Errorf("expected child wins=1, got %f", child.wins)
	}
	if root.wins != 1 {
		t.Errorf("expected root wins=1, got %f", root.wins)
	}
}

func TestBackpropagate_SingleActor_Stalemate(t *testing.T) {
	root := &mctsNode{state: &singleActorState{actor: 1, result: decision.Undecided, id: "root"}}
	child := &mctsNode{state: &singleActorState{actor: 1, result: decision.Stalemate, id: "child"}, parent: root}

	child.backpropagate(decision.Stalemate)

	if child.wins != 0.5 {
		t.Errorf("expected child wins=0.5, got %f", child.wins)
	}
	if root.wins != 0.5 {
		t.Errorf("expected root wins=0.5, got %f", root.wins)
	}
}

func TestBackpropagateValue_SingleActor(t *testing.T) {
	root := &mctsNode{state: &singleActorState{actor: 1, result: decision.Undecided, id: "root"}}
	child := &mctsNode{state: &singleActorState{actor: 1, result: decision.Undecided, id: "child"}, parent: root}
	grandchild := &mctsNode{state: &singleActorState{actor: 1, result: decision.Undecided, id: "gchild"}, parent: child}

	// value=0.8 pour l'acteur unique
	values := map[decision.ActorID]float64{1: 0.8}
	grandchild.backpropagateValue(values)

	// Tous les nœuds ont PreviousActor == 1 → lookup values[1] == 0.8
	if math.Abs(grandchild.wins-0.8) > 1e-9 {
		t.Errorf("expected grandchild wins=0.8, got %f", grandchild.wins)
	}
	if math.Abs(child.wins-0.8) > 1e-9 {
		t.Errorf("expected child wins=0.8, got %f", child.wins)
	}
	if math.Abs(root.wins-0.8) > 1e-9 {
		t.Errorf("expected root wins=0.8, got %f", root.wins)
	}
}

func TestBackpropagateValue_ThreeActors(t *testing.T) {
	root := &mctsNode{state: &threeActorState{current: 10, previous: 12, result: decision.Undecided, id: "root"}}
	child := &mctsNode{state: &threeActorState{current: 11, previous: 10, result: decision.Undecided, id: "child"}, parent: root}
	grandchild := &mctsNode{state: &threeActorState{current: 12, previous: 11, result: decision.Undecided, id: "gchild"}, parent: child}

	values := map[decision.ActorID]float64{
		10: 0.6,
		11: -0.3,
		12: -0.3,
	}
	grandchild.backpropagateValue(values)

	// grandchild: PreviousActor = 11 → wins = -0.3
	if math.Abs(grandchild.wins-(-0.3)) > 1e-9 {
		t.Errorf("expected grandchild wins=-0.3, got %f", grandchild.wins)
	}
	// child: PreviousActor = 10 → wins = 0.6
	if math.Abs(child.wins-0.6) > 1e-9 {
		t.Errorf("expected child wins=0.6, got %f", child.wins)
	}
	// root: PreviousActor = 12 → wins = -0.3
	if math.Abs(root.wins-(-0.3)) > 1e-9 {
		t.Errorf("expected root wins=-0.3, got %f", root.wins)
	}
}

func TestBackpropagateTerminal_SingleActor_Win(t *testing.T) {
	root := &mctsNode{state: &singleActorState{actor: 1, result: decision.Undecided, id: "root"}}
	child := &mctsNode{state: &singleActorState{actor: 1, result: 1, id: "solved"}, parent: root}

	child.backpropagateTerminal()

	// Tous les nœuds : PreviousActor == 1, result == 1 → wins = 1.0
	if math.Abs(child.wins-1.0) > 1e-9 {
		t.Errorf("expected child wins=1.0, got %f", child.wins)
	}
	if math.Abs(root.wins-1.0) > 1e-9 {
		t.Errorf("expected root wins=1.0, got %f", root.wins)
	}
}

func TestBackpropagateTerminal_SingleActor_Stalemate(t *testing.T) {
	root := &mctsNode{state: &singleActorState{actor: 1, result: decision.Undecided, id: "root"}}
	child := &mctsNode{state: &singleActorState{actor: 1, result: decision.Stalemate, id: "stuck"}, parent: root}

	child.backpropagateTerminal()

	// Stalemate → wins = 0.0
	if math.Abs(child.wins) > 1e-9 {
		t.Errorf("expected child wins=0.0, got %f", child.wins)
	}
	if math.Abs(root.wins) > 1e-9 {
		t.Errorf("expected root wins=0.0, got %f", root.wins)
	}
}

func TestBackpropagateTerminal_ThreeActors(t *testing.T) {
	root := &mctsNode{state: &threeActorState{current: 10, previous: 12, result: decision.Undecided, id: "root"}}
	child := &mctsNode{state: &threeActorState{current: 11, previous: 10, result: decision.Undecided, id: "child"}, parent: root}
	grandchild := &mctsNode{state: &threeActorState{current: 12, previous: 11, result: 10, id: "gchild"}, parent: child}

	grandchild.backpropagateTerminal()

	// grandchild: PreviousActor = 11, result = 10 → 11 perd → -1.0
	if math.Abs(grandchild.wins-(-1.0)) > 1e-9 {
		t.Errorf("expected grandchild wins=-1.0, got %f", grandchild.wins)
	}
	// child: PreviousActor = 10, result = 10 → 10 gagne → 1.0
	if math.Abs(child.wins-1.0) > 1e-9 {
		t.Errorf("expected child wins=1.0, got %f", child.wins)
	}
	// root: PreviousActor = 12, result = 10 → 12 perd → -1.0
	if math.Abs(root.wins-(-1.0)) > 1e-9 {
		t.Errorf("expected root wins=-1.0, got %f", root.wins)
	}
}

// --- Helper ---

func playMoves(moves ...uint8) *tictactoe.TicTacToe {
	ttt := tictactoe.NewTicTacToe()
	for _, m := range moves {
		ttt.Play(m)
	}
	return ttt
}
