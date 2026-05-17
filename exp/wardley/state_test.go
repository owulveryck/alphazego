package wardley_test

import (
	"fmt"
	"testing"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/exp/wardley"
)

func sampleComponents() []wardley.Component {
	return []wardley.Component{
		{Name: "App", Phase: wardley.Product, Visibility: 80, Type: "build", Inertia: 0},
		{Name: "DB", Phase: wardley.Custom, Visibility: 50, Type: "buy", Inertia: 0},
		{Name: "Cloud", Phase: wardley.Commodity, Visibility: 20, Type: "outsource", Inertia: 0},
	}
}

func sampleEdges() []wardley.Edge {
	return []wardley.Edge{
		{From: "App", To: "DB"},
		{From: "DB", To: "Cloud"},
	}
}

func ExampleState_CurrentActor() {
	s := wardley.NewState("test", "question?", sampleComponents(), sampleEdges(), 3)
	fmt.Println(s.CurrentActor())
	// Output: 1
}

func ExampleState_Evaluate() {
	s := wardley.NewState("test", "question?", sampleComponents(), sampleEdges(), 3)
	fmt.Println(s.Evaluate())
	// Output: 0
}

func ExampleState_PossibleMoves() {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Custom, Visibility: 50, Inertia: 0},
	}
	s := wardley.NewState("test", "q?", comps, nil, 5)
	moves := s.PossibleMoves()
	fmt.Printf("%d moves possibles\n", len(moves))
	// Output: 7 moves possibles
}

func TestStateImplementsDecisionState(t *testing.T) {
	var _ decision.State = (*wardley.State)(nil)
}

func TestMonoActor(t *testing.T) {
	s := wardley.NewState("t", "q", nil, nil, 5)
	if s.CurrentActor() != wardley.Player {
		t.Errorf("CurrentActor = %d, want %d", s.CurrentActor(), wardley.Player)
	}
	if s.PreviousActor() != wardley.Player {
		t.Errorf("PreviousActor = %d, want %d", s.PreviousActor(), wardley.Player)
	}
}

func TestEvaluateTerminalAtMaxDepth(t *testing.T) {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Genesis, Inertia: 0},
	}
	s := wardley.NewState("t", "q", comps, nil, 1)

	if s.Evaluate() != decision.Undecided {
		t.Fatal("état initial devrait être Undecided")
	}

	moves := s.PossibleMoves()
	if len(moves) == 0 {
		t.Fatal("devrait avoir des moves")
	}

	child := moves[0]
	if child.Evaluate() != decision.Stalemate {
		t.Errorf("child.Evaluate = %d, want Stalemate (%d)", child.Evaluate(), decision.Stalemate)
	}

	if childMoves := child.PossibleMoves(); len(childMoves) != 0 {
		t.Errorf("état terminal ne devrait pas avoir de moves, got %d", len(childMoves))
	}
}

func TestPossibleMovesEvolve(t *testing.T) {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Genesis, Inertia: 0},
		{Name: "B", Phase: wardley.Commodity, Inertia: 0},
	}
	s := wardley.NewState("t", "q", comps, nil, 10)
	moves := s.PossibleMoves()

	evolveMoves := 0
	for _, m := range moves {
		st := m.(*wardley.State)
		if st.LastMove().Type == wardley.Evolve {
			evolveMoves++
		}
	}
	if evolveMoves != 1 {
		t.Errorf("got %d evolve moves, want 1 (seul A peut évoluer, B est Commodity)", evolveMoves)
	}
}

func TestPossibleMovesInertiaBlocks(t *testing.T) {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Custom, Inertia: 2},
	}
	s := wardley.NewState("t", "q", comps, nil, 10)
	moves := s.PossibleMoves()

	for _, m := range moves {
		st := m.(*wardley.State)
		if st.LastMove().Type == wardley.Evolve && st.LastMove().Component == "A" {
			t.Error("composant A avec inertie ne devrait pas pouvoir évoluer")
		}
	}
}

func TestPossibleMovesGameplay(t *testing.T) {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Custom, Inertia: 0, Gameplays: []string{"ILC"}},
	}
	s := wardley.NewState("t", "q", comps, nil, 10)
	moves := s.PossibleMoves()

	gameplayCount := 0
	for _, m := range moves {
		st := m.(*wardley.State)
		if st.LastMove().Type == wardley.ApplyGameplay {
			gameplayCount++
			if st.LastMove().Gameplay == "ILC" {
				t.Error("ILC déjà appliqué ne devrait pas être proposé")
			}
		}
	}

	expectedGameplays := len(wardley.AvailableGameplays) - 1
	if gameplayCount != expectedGameplays {
		t.Errorf("got %d gameplay moves, want %d", gameplayCount, expectedGameplays)
	}
}

func TestIDDeterministic(t *testing.T) {
	s1 := wardley.NewState("t", "q", sampleComponents(), sampleEdges(), 5)
	s2 := wardley.NewState("t", "q", sampleComponents(), sampleEdges(), 5)
	if s1.ID() != s2.ID() {
		t.Errorf("IDs should be equal: %s != %s", s1.ID(), s2.ID())
	}
}

func TestIDChangesAfterMove(t *testing.T) {
	s := wardley.NewState("t", "q", sampleComponents(), sampleEdges(), 5)
	id0 := s.ID()

	moves := s.PossibleMoves()
	if len(moves) == 0 {
		t.Fatal("aucun move")
	}
	child := moves[0]
	if child.ID() == id0 {
		t.Error("ID devrait changer après un move")
	}
}

func TestApplyMoveDoesNotMutateParent(t *testing.T) {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Genesis, Inertia: 0, Gameplays: []string{"ILC"}},
	}
	s := wardley.NewState("t", "q", comps, nil, 10)

	parentComps := s.Components()
	parentPhase := parentComps[0].Phase
	parentGPs := len(parentComps[0].Gameplays)

	_ = s.PossibleMoves()

	afterComps := s.Components()
	if afterComps[0].Phase != parentPhase {
		t.Errorf("parent phase mutée: %d -> %d", parentPhase, afterComps[0].Phase)
	}
	if len(afterComps[0].Gameplays) != parentGPs {
		t.Errorf("parent gameplays mutés: %d -> %d", parentGPs, len(afterComps[0].Gameplays))
	}
}
