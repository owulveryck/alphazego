package wardley

import (
	"context"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/reasoning"
)

// ProgressInfo contient les informations de progression d'une évaluation MCTS.
type ProgressInfo struct {
	// EvalCount est le nombre d'appels à Evaluate depuis la création de l'Evaluator.
	EvalCount int
	// LLMCalls est le nombre total d'appels au Judge (LLM).
	LLMCalls int
	// CandidateCount est le nombre de candidats générés à cette expansion.
	CandidateCount int
	// Value est le score de l'état courant (dans [-1, 1] après conversion).
	Value float64
}

// Evaluator implémente [mcts.Evaluator] pour l'exploration stratégique Wardley.
// Il utilise les confidences cachées par [State.PossibleMoves] comme policy,
// et un [reasoning.Judge] pour estimer la value de l'état courant.
type Evaluator struct {
	judge     reasoning.Judge
	ctx       context.Context
	evalCount int
	llmCalls  int
	// Progress est appelé après chaque évaluation si non nil.
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
// La policy est construite à partir des scores de confiance des candidats
// générés par le [Proposer] (cachés dans l'état). La value est l'estimation
// de la qualité stratégique de l'état courant, convertie de [0, 1] vers
// [-1, 1] : value_mcts = value_judge * 2 - 1.
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
	e.llmCalls++ // PossibleMoves a déjà fait 1 appel LLM (Proposer)

	policy := ws.CachedConfidences()
	sum := 0.0
	for _, p := range policy {
		sum += p
	}
	if sum > 0 {
		for i := range policy {
			policy[i] /= sum
		}
	}

	valuePrompt := formatValuePrompt(ws.WTG2Text(), ws.Question(), ws.History())
	value, err := e.judge.Score(e.ctx, valuePrompt)
	if err != nil {
		value = 0.0
	}
	value = value*2 - 1

	e.llmCalls++
	if e.Progress != nil {
		e.Progress(ProgressInfo{
			EvalCount:      e.evalCount,
			LLMCalls:       e.llmCalls,
			CandidateCount: n,
			Value:          value,
		})
	}

	return policy, map[decision.ActorID]float64{Player: value}
}
