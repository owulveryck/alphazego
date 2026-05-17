package wardley_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/owulveryck/alphazego/exp/wardley"
)

type mockJudge struct {
	score float64
}

func (j *mockJudge) Score(_ context.Context, _ string) (float64, error) {
	return j.score, nil
}

func ExampleEvaluator_Evaluate() {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Custom, Inertia: 0},
	}
	s := wardley.NewState("test", "question?", comps, nil, 5)

	judge := &mockJudge{score: 0.8}
	eval := wardley.NewEvaluator(context.Background(), judge)

	policy, values := eval.Evaluate(s)
	fmt.Printf("policy len: %d\n", len(policy))
	fmt.Printf("value for Player: %.1f\n", values[wardley.Player])
	// Output:
	// policy len: 7
	// value for Player: 0.6
}

func TestEvaluatorPolicyNormalized(t *testing.T) {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Genesis, Inertia: 0},
		{Name: "B", Phase: wardley.Custom, Inertia: 0},
	}
	s := wardley.NewState("t", "q", comps, nil, 5)

	judge := &mockJudge{score: 0.7}
	eval := wardley.NewEvaluator(context.Background(), judge)

	policy, _ := eval.Evaluate(s)

	sum := 0.0
	for _, p := range policy {
		sum += p
	}
	if diff := sum - 1.0; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("policy sum = %f, want 1.0", sum)
	}
}

func TestEvaluatorValueConversion(t *testing.T) {
	tests := []struct {
		judgeScore float64
		wantValue  float64
	}{
		{0.0, -1.0},
		{0.5, 0.0},
		{1.0, 1.0},
		{0.75, 0.5},
	}

	for _, tt := range tests {
		comps := []wardley.Component{
			{Name: "A", Phase: wardley.Custom, Inertia: 0},
		}
		s := wardley.NewState("t", "q", comps, nil, 5)

		judge := &mockJudge{score: tt.judgeScore}
		eval := wardley.NewEvaluator(context.Background(), judge)

		_, values := eval.Evaluate(s)

		got := values[wardley.Player]
		if diff := got - tt.wantValue; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("judgeScore=%.1f: value = %f, want %f", tt.judgeScore, got, tt.wantValue)
		}
	}
}

type mockBatchJudge struct {
	score  float64
	scores []float64
}

func (j *mockBatchJudge) Score(_ context.Context, _ string) (float64, error) {
	return j.score, nil
}

func (j *mockBatchJudge) ScoreBatch(_ context.Context, _ string, count int) ([]float64, error) {
	if j.scores != nil {
		return j.scores, nil
	}
	out := make([]float64, count)
	for i := range out {
		out[i] = j.score
	}
	return out, nil
}

func TestEvaluatorBatchPolicy(t *testing.T) {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Genesis, Inertia: 0},
		{Name: "B", Phase: wardley.Custom, Inertia: 0},
	}
	s := wardley.NewState("t", "q", comps, nil, 5)

	judge := &mockBatchJudge{score: 0.7}
	eval := wardley.NewEvaluator(context.Background(), judge)

	policy, values := eval.Evaluate(s)

	sum := 0.0
	for _, p := range policy {
		sum += p
	}
	if diff := sum - 1.0; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("batch policy sum = %f, want 1.0", sum)
	}

	if _, ok := values[wardley.Player]; !ok {
		t.Error("batch: missing Player value")
	}
}

func TestEvaluatorBatchPolicyVaryingScores(t *testing.T) {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Custom, Inertia: 0},
	}
	s := wardley.NewState("t", "q", comps, nil, 5)

	// 7 moves: 1 evolve + 6 gameplays
	judge := &mockBatchJudge{scores: []float64{0.9, 0.1, 0.3, 0.5, 0.2, 0.4, 0.6}}
	eval := wardley.NewEvaluator(context.Background(), judge)

	policy, _ := eval.Evaluate(s)

	if len(policy) != 7 {
		t.Fatalf("policy len = %d, want 7", len(policy))
	}

	// Le premier move (score 0.9) doit avoir le prior le plus élevé
	maxIdx := 0
	for i, p := range policy {
		if p > policy[maxIdx] {
			maxIdx = i
		}
	}
	if maxIdx != 0 {
		t.Errorf("expected move 0 to have highest prior, got move %d", maxIdx)
	}
}

func TestEvaluatorFallbackWithoutBatchScorer(t *testing.T) {
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Custom, Inertia: 0},
	}
	s := wardley.NewState("t", "q", comps, nil, 5)

	// mockJudge (pas BatchScorer) → fallback individuel
	judge := &mockJudge{score: 0.8}
	eval := wardley.NewEvaluator(context.Background(), judge)

	policy, _ := eval.Evaluate(s)

	sum := 0.0
	for _, p := range policy {
		sum += p
	}
	if diff := sum - 1.0; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("fallback policy sum = %f, want 1.0", sum)
	}
}

func TestEvaluatorNoMovesReturnsEmpty(t *testing.T) {
	s := wardley.NewState("t", "q", nil, nil, 0)

	judge := &mockJudge{score: 0.5}
	eval := wardley.NewEvaluator(context.Background(), judge)

	policy, values := eval.Evaluate(s)
	if policy != nil {
		t.Errorf("policy should be nil for terminal state, got len %d", len(policy))
	}
	if len(values) != 0 {
		t.Errorf("values should be empty for terminal state, got %v", values)
	}
}
