package reasoning

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/owulveryck/alphazego/decision"
)

// mockGenerator retourne des réponses pré-programmées.
type mockGenerator struct {
	responses [][]string
	callIdx   int
}

func (m *mockGenerator) Generate(_ context.Context, _ string, _ int) ([]string, error) {
	if m.callIdx >= len(m.responses) {
		return nil, nil
	}
	r := m.responses[m.callIdx]
	m.callIdx++
	return r, nil
}

// errorGenerator retourne toujours une erreur.
type errorGenerator struct{}

func (e *errorGenerator) Generate(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, fmt.Errorf("connection refused")
}

// mockJudge retourne des scores pré-programmés.
type mockJudge struct {
	scores  []float64
	callIdx int
}

func (m *mockJudge) Score(_ context.Context, _ string) (float64, error) {
	if m.callIdx >= len(m.scores) {
		return 0.5, nil
	}
	s := m.scores[m.callIdx]
	m.callIdx++
	return s, nil
}

func TestNew(t *testing.T) {
	gen := &mockGenerator{}
	ctx := context.Background()

	s := New(ctx, "question", "critère", gen)
	if s.maxDepth != 5 {
		t.Errorf("default maxDepth = %d, want 5", s.maxDepth)
	}
	if s.branchFactor != 3 {
		t.Errorf("default branchFactor = %d, want 3", s.branchFactor)
	}

	s = New(ctx, "question", "critère", gen, WithMaxDepth(10), WithBranchFactor(5))
	if s.maxDepth != 10 {
		t.Errorf("maxDepth = %d, want 10", s.maxDepth)
	}
	if s.branchFactor != 5 {
		t.Errorf("branchFactor = %d, want 5", s.branchFactor)
	}
}

func TestCurrentActor_PreviousActor(t *testing.T) {
	s := New(context.Background(), "q", "c", &mockGenerator{})
	if s.CurrentActor() != Player {
		t.Errorf("CurrentActor() = %d, want %d", s.CurrentActor(), Player)
	}
	if s.PreviousActor() != Player {
		t.Errorf("PreviousActor() = %d, want %d", s.PreviousActor(), Player)
	}
}

func TestEvaluate_Undecided(t *testing.T) {
	s := New(context.Background(), "q", "c", &mockGenerator{})
	if s.Evaluate() != decision.Undecided {
		t.Errorf("Evaluate() = %d, want Undecided", s.Evaluate())
	}
}

func TestEvaluate_Conclusion(t *testing.T) {
	s := &State{
		question:     "q",
		criterion:    "c",
		steps:        []string{"étape 1", "CONCLUSION: la réponse est 42"},
		maxDepth:     5,
		branchFactor: 3,
	}
	if s.Evaluate() != Player {
		t.Errorf("Evaluate() = %d, want Player (%d)", s.Evaluate(), Player)
	}
}

func TestEvaluate_MaxDepth(t *testing.T) {
	s := &State{
		question:     "q",
		criterion:    "c",
		steps:        []string{"a", "b", "c"},
		maxDepth:     3,
		branchFactor: 3,
	}
	if s.Evaluate() != decision.Stalemate {
		t.Errorf("Evaluate() = %d, want Stalemate", s.Evaluate())
	}
}

func TestPossibleMoves_GeneratesCandidates(t *testing.T) {
	gen := &mockGenerator{
		responses: [][]string{
			{"étape A", "étape B", "étape C"},
		},
	}
	s := New(context.Background(), "question", "critère", gen)

	moves := s.PossibleMoves()
	if len(moves) != 3 {
		t.Fatalf("PossibleMoves() returned %d moves, want 3", len(moves))
	}

	for i, move := range moves {
		child := move.(*State)
		if len(child.steps) != 1 {
			t.Errorf("child %d has %d steps, want 1", i, len(child.steps))
		}
		if child.question != "question" {
			t.Errorf("child %d question = %q, want %q", i, child.question, "question")
		}
		if child.criterion != "critère" {
			t.Errorf("child %d criterion = %q, want %q", i, child.criterion, "critère")
		}
	}

	// Vérifier que les étapes sont distinctes
	steps := make(map[string]bool)
	for _, move := range moves {
		child := move.(*State)
		steps[child.steps[0]] = true
	}
	if len(steps) != 3 {
		t.Errorf("expected 3 distinct steps, got %d", len(steps))
	}
}

func TestPossibleMoves_Cached(t *testing.T) {
	gen := &mockGenerator{
		responses: [][]string{
			{"a", "b"},
		},
	}
	s := New(context.Background(), "q", "c", gen, WithBranchFactor(2))

	moves1 := s.PossibleMoves()
	moves2 := s.PossibleMoves()

	if len(moves1) != len(moves2) {
		t.Fatalf("cached moves length mismatch: %d vs %d", len(moves1), len(moves2))
	}
	for i := range moves1 {
		if moves1[i].ID() != moves2[i].ID() {
			t.Errorf("move %d: ID mismatch %q vs %q", i, moves1[i].ID(), moves2[i].ID())
		}
	}
	if gen.callIdx != 1 {
		t.Errorf("Generator called %d times, want 1 (cached)", gen.callIdx)
	}
}

func TestPossibleMoves_Terminal(t *testing.T) {
	s := &State{
		question:     "q",
		criterion:    "c",
		steps:        []string{"CONCLUSION: done"},
		maxDepth:     5,
		branchFactor: 3,
	}
	if moves := s.PossibleMoves(); moves != nil {
		t.Errorf("PossibleMoves() = %v, want nil for terminal state", moves)
	}
}

func TestID_Deterministic(t *testing.T) {
	s1 := &State{question: "q", steps: []string{"a", "b"}}
	s2 := &State{question: "q", steps: []string{"a", "b"}}
	if s1.ID() != s2.ID() {
		t.Errorf("same state, different IDs: %q vs %q", s1.ID(), s2.ID())
	}
}

func TestID_DifferentSteps(t *testing.T) {
	s1 := &State{question: "q", steps: []string{"a"}}
	s2 := &State{question: "q", steps: []string{"b"}}
	if s1.ID() == s2.ID() {
		t.Errorf("different steps should produce different IDs")
	}
}

func TestSteps_Question_Criterion(t *testing.T) {
	s := New(context.Background(), "ma question", "mon critère", &mockGenerator{})
	if s.Question() != "ma question" {
		t.Errorf("Question() = %q, want %q", s.Question(), "ma question")
	}
	if s.Criterion() != "mon critère" {
		t.Errorf("Criterion() = %q, want %q", s.Criterion(), "mon critère")
	}
	if len(s.Steps()) != 0 {
		t.Errorf("Steps() length = %d, want 0", len(s.Steps()))
	}
}

func TestEvaluator_PolicyNormalized(t *testing.T) {
	gen := &mockGenerator{
		responses: [][]string{
			{"a", "b", "c"},
		},
	}
	judge := &mockJudge{
		scores: []float64{0.8, 0.5, 0.3, 0.6}, // 3 pour policy + 1 pour value
	}
	s := New(context.Background(), "q", "c", gen)
	eval := NewEvaluator(context.Background(), judge)

	policy, values := eval.Evaluate(s)

	if len(policy) != 3 {
		t.Fatalf("policy length = %d, want 3", len(policy))
	}

	sum := 0.0
	for _, p := range policy {
		sum += p
	}
	if math.Abs(sum-1.0) > 1e-6 {
		t.Errorf("policy sum = %f, want 1.0", sum)
	}

	if _, ok := values[Player]; !ok {
		t.Error("values should contain Player")
	}
}

func TestEvaluator_Value(t *testing.T) {
	gen := &mockGenerator{
		responses: [][]string{
			{"a"},
		},
	}
	judge := &mockJudge{
		scores: []float64{0.5, 0.8}, // 1 pour policy + 1 pour value (0.8)
	}
	s := New(context.Background(), "q", "c", gen, WithBranchFactor(1))
	eval := NewEvaluator(context.Background(), judge)

	_, values := eval.Evaluate(s)

	// Value 0.8 en [0,1] → 0.6 en [-1,1]
	expected := 0.8*2 - 1
	if math.Abs(values[Player]-expected) > 1e-6 {
		t.Errorf("values[Player] = %f, want %f", values[Player], expected)
	}
}

func TestEvaluator_EmptyMoves(t *testing.T) {
	s := &State{
		question:     "q",
		criterion:    "c",
		steps:        []string{"CONCLUSION: done"},
		maxDepth:     5,
		branchFactor: 3,
	}
	eval := NewEvaluator(context.Background(), &mockJudge{})

	policy, values := eval.Evaluate(s)
	if policy != nil {
		t.Errorf("policy should be nil for terminal state")
	}
	if len(values) != 0 {
		t.Errorf("values should be empty for terminal state")
	}
}

func TestPossibleMoves_GeneratorError(t *testing.T) {
	gen := &errorGenerator{}
	s := New(context.Background(), "question", "critère", gen)

	moves := s.PossibleMoves()
	if moves != nil {
		t.Errorf("PossibleMoves() = %v, want nil on Generator error", moves)
	}
	if s.LastError() == nil {
		t.Fatal("LastError() should be non-nil after Generator error")
	}
	if s.LastError().Error() == "" {
		t.Error("LastError() should contain a message")
	}
}

func TestPossibleMoves_EmptyCandidates(t *testing.T) {
	gen := &mockGenerator{
		responses: [][]string{
			{}, // retourne une liste vide
		},
	}
	s := New(context.Background(), "question", "critère", gen)

	moves := s.PossibleMoves()
	if moves != nil {
		t.Errorf("PossibleMoves() = %v, want nil for empty candidates", moves)
	}
	if s.LastError() == nil {
		t.Fatal("LastError() should be non-nil when Generator returns empty candidates")
	}
}

func TestLastError_NilByDefault(t *testing.T) {
	s := New(context.Background(), "q", "c", &mockGenerator{
		responses: [][]string{{"a"}},
	})
	if s.LastError() != nil {
		t.Errorf("LastError() = %v, want nil", s.LastError())
	}
	// Après un appel réussi, LastError reste nil
	s.PossibleMoves()
	if s.LastError() != nil {
		t.Errorf("LastError() = %v after successful PossibleMoves, want nil", s.LastError())
	}
}
