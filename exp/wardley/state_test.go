package wardley_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/exp/wardley"
)

type mockProposer struct {
	candidates []wardley.Candidate
}

func (p *mockProposer) Propose(_ context.Context, _ string, _ int) ([]wardley.Candidate, error) {
	return p.candidates, nil
}

const sampleWTG2Text = `title: Test
question: "Should we evolve?"
stages: Genesis, Custom, Product, Commodity

App : III.5
DB : II.5

App -> DB
`

const childWTG2Text = `title: Test
question: "Should we evolve?"
stages: Genesis, Custom, Product, Commodity

App : III.5
DB : III.5

App -> DB
`

func sampleProposer() *mockProposer {
	return &mockProposer{
		candidates: []wardley.Candidate{
			{Description: "Evolve DB to Product", WTG2: childWTG2Text, Confidence: 0.8},
			{Description: "Add gameplay on App", WTG2: sampleWTG2Text, Confidence: 0.5},
		},
	}
}

func ExampleState_CurrentActor() {
	s := wardley.NewState(sampleWTG2Text, "Test", "Should we evolve?", 3, sampleProposer(), context.Background())
	fmt.Println(s.CurrentActor())
	// Output: 1
}

func ExampleState_Evaluate() {
	s := wardley.NewState(sampleWTG2Text, "Test", "Should we evolve?", 3, sampleProposer(), context.Background())
	fmt.Println(s.Evaluate())
	// Output: 0
}

func TestStateImplementsDecisionState(t *testing.T) {
	var _ decision.State = (*wardley.State)(nil)
}

func TestMonoActor(t *testing.T) {
	s := wardley.NewState("", "", "", 5, sampleProposer(), context.Background())
	if s.CurrentActor() != wardley.Player {
		t.Errorf("CurrentActor = %d, want %d", s.CurrentActor(), wardley.Player)
	}
	if s.PreviousActor() != wardley.Player {
		t.Errorf("PreviousActor = %d, want %d", s.PreviousActor(), wardley.Player)
	}
}

func TestEvaluateTerminalAtMaxDepth(t *testing.T) {
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 1, sampleProposer(), context.Background())

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

func TestPossibleMovesFromProposer(t *testing.T) {
	proposer := sampleProposer()
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 5, proposer, context.Background())

	moves := s.PossibleMoves()
	if len(moves) != 2 {
		t.Fatalf("got %d moves, want 2", len(moves))
	}

	child := moves[0].(*wardley.State)
	if child.LastDescription() != "Evolve DB to Product" {
		t.Errorf("LastDescription = %q, want %q", child.LastDescription(), "Evolve DB to Product")
	}
}

func TestPossibleMovesCached(t *testing.T) {
	proposer := sampleProposer()
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 5, proposer, context.Background())

	moves1 := s.PossibleMoves()
	moves2 := s.PossibleMoves()

	if len(moves1) != len(moves2) {
		t.Fatalf("cache broken: %d vs %d", len(moves1), len(moves2))
	}
	for i := range moves1 {
		if moves1[i].ID() != moves2[i].ID() {
			t.Errorf("cache broken at index %d", i)
		}
	}
}

func TestCachedConfidences(t *testing.T) {
	proposer := sampleProposer()
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 5, proposer, context.Background())

	_ = s.PossibleMoves()
	conf := s.CachedConfidences()
	if len(conf) != 2 {
		t.Fatalf("got %d confidences, want 2", len(conf))
	}
	if conf[0] < 0.79 || conf[0] > 0.81 {
		t.Errorf("confidence[0] = %f, want ~0.8", conf[0])
	}
}

func TestIDDeterministic(t *testing.T) {
	s1 := wardley.NewState(sampleWTG2Text, "Test", "q", 5, sampleProposer(), context.Background())
	s2 := wardley.NewState(sampleWTG2Text, "Test", "q", 5, sampleProposer(), context.Background())
	if s1.ID() != s2.ID() {
		t.Errorf("IDs should be equal: %s != %s", s1.ID(), s2.ID())
	}
}

func TestIDChangesWithDifferentWTG2(t *testing.T) {
	s1 := wardley.NewState(sampleWTG2Text, "Test", "q", 5, sampleProposer(), context.Background())
	s2 := wardley.NewState(childWTG2Text, "Test", "q", 5, sampleProposer(), context.Background())
	if s1.ID() == s2.ID() {
		t.Error("IDs should differ for different WTG2 text")
	}
}

func TestHistoryAccumulates(t *testing.T) {
	proposer := sampleProposer()
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 5, proposer, context.Background())

	if len(s.History()) != 0 {
		t.Fatalf("initial state should have empty history")
	}

	moves := s.PossibleMoves()
	child := moves[0].(*wardley.State)
	hist := child.History()
	if len(hist) != 1 {
		t.Fatalf("child should have 1 history entry, got %d", len(hist))
	}
	if hist[0] != "Evolve DB to Product" {
		t.Errorf("history[0] = %q, want %q", hist[0], "Evolve DB to Product")
	}
}

func TestWTG2TextPreserved(t *testing.T) {
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 5, sampleProposer(), context.Background())
	if s.WTG2Text() != sampleWTG2Text {
		t.Errorf("WTG2Text not preserved:\ngot:  %q\nwant: %q", s.WTG2Text(), sampleWTG2Text)
	}
}
