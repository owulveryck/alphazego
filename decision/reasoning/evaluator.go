package reasoning

import (
	"context"
	"math"

	"github.com/owulveryck/alphazego/decision"
)

// Evaluator implémente [mcts.Evaluator] pour le raisonnement par décomposition.
// Il utilise un [Judge] pour estimer la qualité de chaque chemin de raisonnement
// (policy) et la progression vers la solution (value).
type Evaluator struct {
	judge Judge
	ctx   context.Context
}

// NewEvaluator crée un Evaluator qui utilise le [Judge] donné pour évaluer
// les chemins de raisonnement.
func NewEvaluator(ctx context.Context, judge Judge) *Evaluator {
	return &Evaluator{
		judge: judge,
		ctx:   ctx,
	}
}

// Evaluate retourne la policy et les values pour un état de raisonnement.
//
// La policy est construite en demandant au [Judge] de scorer chaque état
// enfant (étape candidate). Les scores sont normalisés pour sommer à 1.
//
// La value est l'estimation de progression de l'état courant vers la solution.
func (e *Evaluator) Evaluate(state decision.State) ([]float64, map[decision.ActorID]float64) {
	moves := state.PossibleMoves()
	n := len(moves)
	if n == 0 {
		return nil, map[decision.ActorID]float64{}
	}

	rs := state.(*State)

	// Policy : scorer chaque candidat
	policy := make([]float64, n)
	for i, move := range moves {
		child := move.(*State)
		prompt := formatJudgePrompt(child.question, child.criterion, child.steps)
		score, err := e.judge.Score(e.ctx, prompt)
		if err != nil {
			score = 1.0 / float64(n) // fallback uniforme
		}
		policy[i] = math.Max(score, 1e-8) // éviter les zéros
	}

	// Normaliser la policy
	sum := 0.0
	for _, p := range policy {
		sum += p
	}
	for i := range policy {
		policy[i] /= sum
	}

	// Value : scorer l'état courant
	valuePrompt := formatValuePrompt(rs.question, rs.criterion, rs.steps)
	value, err := e.judge.Score(e.ctx, valuePrompt)
	if err != nil {
		value = 0.0
	}
	// Convertir de [0, 1] vers [-1, 1]
	value = value*2 - 1

	values := map[decision.ActorID]float64{
		Player: value,
	}

	return policy, values
}
