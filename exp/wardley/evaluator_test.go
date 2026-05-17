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
	proposer := sampleProposer()
	s := wardley.NewState(sampleWTG2Text, "Test", "question?", 5, proposer, context.Background())

	judge := &mockJudge{score: 0.8}
	eval := wardley.NewEvaluator(context.Background(), judge)

	policy, values := eval.Evaluate(s)
	fmt.Printf("policy len: %d\n", len(policy))
	fmt.Printf("value for Player: %.1f\n", values[wardley.Player])
	// Output:
	// policy len: 2
	// value for Player: 0.6
}

func TestEvaluatorPolicyFromConfidences(t *testing.T) {
	proposer := &mockProposer{
		candidates: []wardley.Candidate{
			{Description: "Move A", WTG2: childWTG2Text, Confidence: 0.8},
			{Description: "Move B", WTG2: sampleWTG2Text, Confidence: 0.2},
		},
	}
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 5, proposer, context.Background())

	judge := &mockJudge{score: 0.5}
	eval := wardley.NewEvaluator(context.Background(), judge)

	policy, _ := eval.Evaluate(s)

	sum := 0.0
	for _, p := range policy {
		sum += p
	}
	if diff := sum - 1.0; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("policy sum = %f, want 1.0", sum)
	}

	if policy[0] <= policy[1] {
		t.Errorf("policy[0]=%f should be > policy[1]=%f (higher confidence)", policy[0], policy[1])
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

	proposer := sampleProposer()
	for _, tt := range tests {
		s := wardley.NewState(sampleWTG2Text, "Test", "q", 5, proposer, context.Background())

		judge := &mockJudge{score: tt.judgeScore}
		eval := wardley.NewEvaluator(context.Background(), judge)

		_, values := eval.Evaluate(s)

		got := values[wardley.Player]
		if diff := got - tt.wantValue; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("judgeScore=%.1f: value = %f, want %f", tt.judgeScore, got, tt.wantValue)
		}
	}
}

func TestEvaluatorNoMovesReturnsEmpty(t *testing.T) {
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 0, sampleProposer(), context.Background())

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
