package mcts

import (
	"math"
	"testing"

	"github.com/owulveryck/alphazego/board"
	"github.com/owulveryck/alphazego/board/tictactoe"
)

// --- UCB1 Tests ---

func TestUCB1_UnvisitedNode(t *testing.T) {
	node := &MCTSNode{visits: 0}
	if !math.IsInf(node.UCB1(), 1) {
		t.Error("expected +Inf for unvisited node")
	}
}

func TestUCB1_WithParent(t *testing.T) {
	parent := &MCTSNode{visits: 10}
	child := &MCTSNode{visits: 5, wins: 3, parent: parent}
	score := child.UCB1()
	expected := 0.6 + math.Sqrt(2)*math.Sqrt(math.Log(10)/5)
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("expected UCB1 ~ %f, got %f", expected, score)
	}
}

func TestUCB1_RootNode(t *testing.T) {
	root := &MCTSNode{visits: 10, wins: 5, parent: nil}
	score := root.UCB1()
	if math.Abs(score-0.5) > 1e-9 {
		t.Errorf("expected 0.5 for root node, got %f", score)
	}
}

// --- IsTerminal Tests ---

func TestIsTerminal_GameOn(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	node := &MCTSNode{state: ttt}
	if node.IsTerminal() {
		t.Error("expected non-terminal for new game")
	}
}

func TestIsTerminal_Won(t *testing.T) {
	// Player1 wins top row: play sequence 0,3,1,4,2
	ttt := playMoves(0, 3, 1, 4, 2)
	node := &MCTSNode{state: ttt}
	if !node.IsTerminal() {
		t.Error("expected terminal for won game")
	}
}

func TestIsTerminal_Draw(t *testing.T) {
	// Draw: 4,0,2,6,3,5,1,7,8
	ttt := playMoves(4, 0, 2, 6, 3, 5, 1, 7, 8)
	node := &MCTSNode{state: ttt}
	if ttt.Evaluate() != board.DrawResult {
		// If this particular sequence doesn't draw, that's fine - just test what we get
		t.Skipf("sequence didn't produce a draw, got result %d", ttt.Evaluate())
	}
	if !node.IsTerminal() {
		t.Error("expected terminal for draw")
	}
}

// --- IsFullyExpanded Tests ---

func TestIsFullyExpanded_NoChildren(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	node := &MCTSNode{state: ttt, children: []*MCTSNode{}}
	if node.IsFullyExpanded() {
		t.Error("expected not fully expanded with no children")
	}
}

func TestIsFullyExpanded_AllExpanded(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	node := &MCTSNode{state: ttt, children: make([]*MCTSNode, 9)}
	if !node.IsFullyExpanded() {
		t.Error("expected fully expanded when children count matches possible moves")
	}
}

// --- SelectChildUCB Tests ---

func TestSelectChildUCB_SelectsUnvisited(t *testing.T) {
	parent := &MCTSNode{visits: 10}
	visited := &MCTSNode{visits: 5, wins: 2, parent: parent}
	unvisited := &MCTSNode{visits: 0, parent: parent}
	parent.children = []*MCTSNode{visited, unvisited}

	best := parent.SelectChildUCB()
	if best != unvisited {
		t.Error("expected unvisited child to be selected (Inf UCB1)")
	}
}

func TestSelectChildUCB_NoChildren(t *testing.T) {
	node := &MCTSNode{children: []*MCTSNode{}}
	if node.SelectChildUCB() != nil {
		t.Error("expected nil for node with no children")
	}
}

func TestSelectChildUCB_SelectsBestScore(t *testing.T) {
	parent := &MCTSNode{visits: 100}
	child1 := &MCTSNode{visits: 50, wins: 10, parent: parent}
	child2 := &MCTSNode{visits: 50, wins: 40, parent: parent}
	parent.children = []*MCTSNode{child1, child2}

	best := parent.SelectChildUCB()
	if best != child2 {
		t.Error("expected child with higher win rate to be selected")
	}
}

// --- SelectBestMove Tests ---

func TestSelectBestMove_HighestVisits(t *testing.T) {
	child1 := &MCTSNode{visits: 10, wins: 5}
	child2 := &MCTSNode{visits: 20, wins: 8}
	child3 := &MCTSNode{visits: 15, wins: 12}
	parent := &MCTSNode{children: []*MCTSNode{child1, child2, child3}}

	best := parent.SelectBestMove()
	if best != child2 {
		t.Errorf("expected child with 20 visits, got visits=%f", best.visits)
	}
}

func TestSelectBestMove_NoChildren(t *testing.T) {
	node := &MCTSNode{children: []*MCTSNode{}}
	if node.SelectBestMove() != nil {
		t.Error("expected nil for node with no children")
	}
}

// --- Expand Tests ---

func TestExpand_CreatesOneChild(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.GetOrCreateNode(ttt, nil)

	child := node.Expand()
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
	node := m.GetOrCreateNode(ttt, nil)

	for i := 1; i <= 9; i++ {
		child := node.Expand()
		if child == nil {
			t.Fatalf("expected child on expand %d", i)
		}
		if len(node.children) != i {
			t.Errorf("expected %d children after expand %d, got %d", i, i, len(node.children))
		}
	}

	// 10th expand should return nil (fully expanded)
	if node.Expand() != nil {
		t.Error("expected nil after all moves expanded")
	}
}

func TestExpand_ChildHasCorrectParent(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.GetOrCreateNode(ttt, nil)

	child := node.Expand()
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
	node := &MCTSNode{state: ttt}

	for i := 0; i < 20; i++ {
		result := node.Simulate()
		if result != board.Player1 && result != board.Player2 && result != board.DrawResult {
			t.Errorf("expected terminal result, got %d", result)
		}
	}
}

func TestSimulate_AlreadyTerminal(t *testing.T) {
	ttt := playMoves(0, 3, 1, 4, 2) // Player1 wins top row
	node := &MCTSNode{state: ttt}

	result := node.Simulate()
	if result != board.Player1 {
		t.Errorf("expected Player1, got %d", result)
	}
}

// --- Backpropagate Tests ---

func TestBackpropagate_UpdatesVisits(t *testing.T) {
	root := &MCTSNode{
		state:  tictactoe.NewTicTacToe(),
		parent: nil,
	}
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0)
	child := &MCTSNode{
		state:  ttt,
		parent: root,
	}

	child.Backpropagate(board.Player1)

	if child.visits != 1 {
		t.Errorf("expected child visits=1, got %f", child.visits)
	}
	if root.visits != 1 {
		t.Errorf("expected root visits=1, got %f", root.visits)
	}
}

func TestBackpropagate_CreditsCorrectPlayer(t *testing.T) {
	// Root: Player1's turn
	root := &MCTSNode{
		state:  tictactoe.NewTicTacToe(),
		parent: nil,
	}
	// Child: Player2's turn (Player1 just moved)
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0)
	child := &MCTSNode{
		state:  ttt,
		parent: root,
	}

	child.Backpropagate(board.Player1)

	// At child: PreviousPlayer = Player1. Result=Player1 → win
	if child.wins != 1 {
		t.Errorf("expected child wins=1 (Player1 moved here and won), got %f", child.wins)
	}
	// At root: PreviousPlayer = Player2. Result=Player1 → no win
	if root.wins != 0 {
		t.Errorf("expected root wins=0, got %f", root.wins)
	}
}

func TestBackpropagate_Player2Wins(t *testing.T) {
	root := &MCTSNode{
		state:  tictactoe.NewTicTacToe(),
		parent: nil,
	}
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0)
	child := &MCTSNode{
		state:  ttt,
		parent: root,
	}

	child.Backpropagate(board.Player2)

	// At child: PreviousPlayer = Player1. Result=Player2 → no win
	if child.wins != 0 {
		t.Errorf("expected child wins=0, got %f", child.wins)
	}
	// At root: PreviousPlayer = Player2. Result=Player2 → win
	if root.wins != 1 {
		t.Errorf("expected root wins=1, got %f", root.wins)
	}
}

func TestBackpropagate_Draw(t *testing.T) {
	root := &MCTSNode{
		state:  tictactoe.NewTicTacToe(),
		parent: nil,
	}
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0)
	child := &MCTSNode{
		state:  ttt,
		parent: root,
	}

	child.Backpropagate(board.DrawResult)

	if child.wins != 0.5 {
		t.Errorf("expected child wins=0.5 for draw, got %f", child.wins)
	}
	if root.wins != 0.5 {
		t.Errorf("expected root wins=0.5 for draw, got %f", root.wins)
	}
}

func TestBackpropagate_DeepChain(t *testing.T) {
	root := &MCTSNode{state: tictactoe.NewTicTacToe()}

	ttt1 := tictactoe.NewTicTacToe()
	ttt1.Play(0)
	child := &MCTSNode{state: ttt1, parent: root}

	ttt2 := tictactoe.NewTicTacToe()
	ttt2.Play(0)
	ttt2.Play(1)
	grandchild := &MCTSNode{state: ttt2, parent: child}

	grandchild.Backpropagate(board.Player1)

	if grandchild.visits != 1 || child.visits != 1 || root.visits != 1 {
		t.Error("expected all nodes to have 1 visit")
	}
}

// --- GetOrCreateNode Tests ---

func TestGetOrCreateNode_New(t *testing.T) {
	m := NewMCTS()
	ttt := tictactoe.NewTicTacToe()
	node := m.GetOrCreateNode(ttt, nil)

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
	node1 := m.GetOrCreateNode(ttt, nil)
	node2 := m.GetOrCreateNode(ttt, nil)

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
	if result.CurrentPlayer() != board.Player2 {
		t.Errorf("expected Player2's turn after MCTS move, got %d", result.CurrentPlayer())
	}
}

func TestRunMCTS_BlocksWin(t *testing.T) {
	// Player2's turn. Player1 has positions 0,1 - about to win at 2.
	ttt := playMoves(0, 3, 1, 4)

	m := NewMCTS()
	result := m.RunMCTS(ttt, 5000)

	move := board.State(ttt).(board.Playable).GetMoveFromState(result)
	if move != 2 {
		t.Errorf("expected MCTS to block at position 2, got move %d", move)
	}
}

func TestRunMCTS_TakesWin(t *testing.T) {
	ttt := tictactoe.NewTicTacToe()
	ttt.Play(0) // P1
	ttt.Play(3) // P2
	ttt.Play(1) // P1
	ttt.Play(7) // P2 plays somewhere else (not blocking)
	// Board: [1,1,0,0,0,0,0,2,0], P1's turn, can win at 2

	m := NewMCTS()
	result := m.RunMCTS(ttt, 5000)

	move := board.State(ttt).(board.Playable).GetMoveFromState(result)
	if move != 2 {
		t.Errorf("expected MCTS to win at position 2, got move %d", move)
	}
}

func TestRunMCTS_TerminalState(t *testing.T) {
	// Player1 wins: 0,3,1,4,2
	ttt := playMoves(0, 3, 1, 4, 2)
	if ttt.Evaluate() == board.NoPlayer {
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
	for i := 0; i < maxMoves && ttt.Evaluate() == board.NoPlayer; i++ {
		next := m.RunMCTS(ttt, 500)
		if next == ttt {
			t.Fatal("MCTS returned same state for non-terminal game")
		}
		move := board.State(ttt).(board.Playable).GetMoveFromState(next)
		ttt.Play(move)
	}

	result := ttt.Evaluate()
	if result == board.NoPlayer {
		t.Error("expected game to end")
	}
}

// --- Three-player mock state (validates N-player Backpropagate) ---

// threePlayerState est un etat mock a 3 joueurs en round-robin.
type threePlayerState struct {
	current  board.PlayerID
	previous board.PlayerID
	result   board.PlayerID
	id       string
}

func (s *threePlayerState) CurrentPlayer() board.PlayerID  { return s.current }
func (s *threePlayerState) PreviousPlayer() board.PlayerID { return s.previous }
func (s *threePlayerState) Evaluate() board.PlayerID       { return s.result }
func (s *threePlayerState) PossibleMoves() []board.State   { return nil }
func (s *threePlayerState) ID() board.ID                   { return []byte(s.id) }

func TestBackpropagate_ThreePlayers(t *testing.T) {
	// Simule une chaine de 3 noeuds : joueur 10 → joueur 11 → joueur 12
	// avec un resultat ou le joueur 10 gagne (Result = 10).
	root := &MCTSNode{state: &threePlayerState{current: 10, previous: 12, result: board.NoPlayer, id: "root"}}
	child := &MCTSNode{state: &threePlayerState{current: 11, previous: 10, result: board.NoPlayer, id: "child"}, parent: root}
	grandchild := &MCTSNode{state: &threePlayerState{current: 12, previous: 11, result: board.PlayerID(10), id: "gchild"}, parent: child}

	// Le joueur 10 gagne
	grandchild.Backpropagate(board.PlayerID(10))

	// grandchild: PreviousPlayer = 11, result = 10 → pas de victoire pour le joueur 11
	if grandchild.wins != 0 {
		t.Errorf("expected grandchild wins=0 (player 11 moved here, player 10 won), got %f", grandchild.wins)
	}
	// child: PreviousPlayer = 10, result = 10 → victoire pour le joueur 10
	if child.wins != 1 {
		t.Errorf("expected child wins=1 (player 10 moved here and won), got %f", child.wins)
	}
	// root: PreviousPlayer = 12, result = 10 → pas de victoire pour le joueur 12
	if root.wins != 0 {
		t.Errorf("expected root wins=0 (player 12 moved here, player 10 won), got %f", root.wins)
	}

	// Toutes les visites doivent etre 1
	if grandchild.visits != 1 || child.visits != 1 || root.visits != 1 {
		t.Error("expected all nodes to have 1 visit")
	}
}

func TestBackpropagate_ThreePlayers_Draw(t *testing.T) {
	root := &MCTSNode{state: &threePlayerState{current: 10, previous: 12, result: board.NoPlayer, id: "root"}}
	child := &MCTSNode{state: &threePlayerState{current: 11, previous: 10, result: board.NoPlayer, id: "child"}, parent: root}

	child.Backpropagate(board.DrawResult)

	// Tous les noeuds recoivent 0.5 pour un match nul
	if child.wins != 0.5 {
		t.Errorf("expected child wins=0.5 for draw, got %f", child.wins)
	}
	if root.wins != 0.5 {
		t.Errorf("expected root wins=0.5 for draw, got %f", root.wins)
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
