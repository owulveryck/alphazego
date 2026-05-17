package wardley

import (
	"context"
	"math"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/reasoning"
)

// BatchScorer évalue plusieurs moves en un seul appel LLM.
// Les implémentations retournent un score [0,1] par move, dans l'ordre.
type BatchScorer interface {
	ScoreBatch(ctx context.Context, prompt string, count int) ([]float64, error)
}

// ProgressInfo contient les informations de progression d'une évaluation MCTS.
type ProgressInfo struct {
	// EvalCount est le nombre d'appels à Evaluate depuis la création de l'Evaluator.
	EvalCount int
	// LLMCalls est le nombre total d'appels au Judge (LLM).
	LLMCalls int
	// PolicyScored est le nombre de moves scorés dans l'évaluation courante.
	PolicyScored int
	// PolicyTotal est le nombre total de moves à scorer dans l'évaluation courante.
	PolicyTotal int
	// Value est le score de l'état courant (dans [-1, 1] après conversion).
	Value float64
}

// Evaluator implémente [mcts.Evaluator] pour l'exploration stratégique Wardley.
// Il utilise un [reasoning.Judge] pour estimer la qualité de chaque état de carte
// (policy) et la progression vers une bonne stratégie (value).
type Evaluator struct {
	judge     reasoning.Judge
	ctx       context.Context
	evalCount int
	llmCalls  int
	// Progress est appelé après chaque appel au Judge si non nil.
	Progress func(ProgressInfo)
}

// NewEvaluator crée un Evaluator qui utilise le [reasoning.Judge] donné.
func NewEvaluator(ctx context.Context, judge reasoning.Judge) *Evaluator {
	return &Evaluator{
		judge: judge,
		ctx:   ctx,
	}
}

// ResetCounters remet les compteurs à zéro (utile entre deux steps).
func (e *Evaluator) ResetCounters() {
	e.evalCount = 0
	e.llmCalls = 0
}

// Evaluate retourne la policy et les values pour un état de carte Wardley.
//
// Si le Judge implémente [BatchScorer], la policy est calculée en un seul
// appel LLM (batch). Sinon, chaque état enfant est scoré individuellement.
// Les scores sont normalisés pour sommer à 1.
//
// La value est l'estimation de la qualité stratégique de l'état courant,
// convertie de [0, 1] vers [-1, 1] : value_mcts = value_judge * 2 - 1.
func (e *Evaluator) Evaluate(state decision.State) ([]float64, map[decision.ActorID]float64) {
	moves := state.PossibleMoves()
	n := len(moves)
	if n == 0 {
		return nil, map[decision.ActorID]float64{}
	}

	ws, ok := state.(*State)
	if !ok {
		policy := make([]float64, n)
		for i := range policy {
			policy[i] = 1.0 / float64(n)
		}
		return policy, map[decision.ActorID]float64{Player: 0.0}
	}

	e.evalCount++

	policy := make([]float64, n)

	if batcher, ok := e.judge.(BatchScorer); ok {
		policy = e.evaluatePolicyBatch(batcher, ws, moves)
	} else {
		policy = e.evaluatePolicyIndividual(moves, n)
	}

	sum := 0.0
	for _, p := range policy {
		sum += p
	}
	for i := range policy {
		policy[i] /= sum
	}

	wtg2Text := SerializeWTG2(ws)
	valuePrompt := formatValuePrompt(wtg2Text, ws.Question(), ws.History())
	value, err := e.judge.Score(e.ctx, valuePrompt)
	if err != nil {
		value = 0.0
	}
	value = value*2 - 1

	e.llmCalls++
	if e.Progress != nil {
		e.Progress(ProgressInfo{
			EvalCount:    e.evalCount,
			LLMCalls:     e.llmCalls,
			PolicyScored: n,
			PolicyTotal:  n,
			Value:        value,
		})
	}

	values := map[decision.ActorID]float64{
		Player: value,
	}

	return policy, values
}

func (e *Evaluator) evaluatePolicyBatch(batcher BatchScorer, ws *State, moves []decision.State) []float64 {
	n := len(moves)

	childMoves := make([]Move, n)
	for i, move := range moves {
		if child, ok := move.(*State); ok {
			childMoves[i] = child.LastMove()
		}
	}

	wtg2Text := SerializeWTG2(ws)
	prompt := formatBatchPolicyPrompt(wtg2Text, ws.Question(), childMoves, ws.Components())
	scores, err := batcher.ScoreBatch(e.ctx, prompt, n)

	e.llmCalls++
	if e.Progress != nil {
		e.Progress(ProgressInfo{
			EvalCount:    e.evalCount,
			LLMCalls:     e.llmCalls,
			PolicyScored: n,
			PolicyTotal:  n,
		})
	}

	policy := make([]float64, n)
	if err != nil || len(scores) != n {
		for i := range policy {
			policy[i] = 1.0 / float64(n)
		}
		return policy
	}
	for i, s := range scores {
		policy[i] = math.Max(s, 1e-8)
	}
	return policy
}

func (e *Evaluator) evaluatePolicyIndividual(moves []decision.State, n int) []float64 {
	policy := make([]float64, n)
	for i, move := range moves {
		child, ok := move.(*State)
		if !ok {
			policy[i] = 1.0 / float64(n)
			continue
		}
		wtg2Text := SerializeWTG2(child)
		prompt := formatPolicyPrompt(wtg2Text, child.Question())
		score, err := e.judge.Score(e.ctx, prompt)
		if err != nil {
			score = 1.0 / float64(n)
		}
		policy[i] = math.Max(score, 1e-8)

		e.llmCalls++
		if e.Progress != nil {
			e.Progress(ProgressInfo{
				EvalCount:    e.evalCount,
				LLMCalls:     e.llmCalls,
				PolicyScored: i + 1,
				PolicyTotal:  n,
			})
		}
	}
	return policy
}
