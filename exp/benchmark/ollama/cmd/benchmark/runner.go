package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/reasoning"
	"github.com/owulveryck/alphazego/exp/benchmark/ollama"
	"github.com/owulveryck/alphazego/exp/benchmark/problems"
	"github.com/owulveryck/alphazego/mcts"
)

// TokenStats accumule les compteurs de tokens pour une exécution.
type TokenStats struct {
	PromptTokens int
	OutputTokens int
	mu           sync.Mutex
}

// Add ajoute les tokens d'une réponse.
func (ts *TokenStats) Add(promptTokens, outputTokens int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.PromptTokens += promptTokens
	ts.OutputTokens += outputTokens
}

// Total retourne le nombre total de tokens.
func (ts *TokenStats) Total() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.PromptTokens + ts.OutputTokens
}

// verbose contrôle l'affichage des logs détaillés.
var verbose bool

// Config décrit une configuration de benchmark.
type Config struct {
	Name       string
	UseMCTS    bool
	Iterations int
}

// AllConfigs retourne les 2 configurations du benchmark Ollama.
func AllConfigs() []Config {
	return []Config{
		{Name: "E (ollama)", UseMCTS: false},
		{Name: "F (ollama+MCTS)", UseMCTS: true, Iterations: 15},
	}
}

// RunOneShot résout le problème en un seul appel au modèle.
func RunOneShot(ctx context.Context, baseURL, model string, problem problems.Problem, tokens *TokenStats) (string, error) {
	prompt := problem.FormatPrompt()
	prompt += "\nDonne l'ordonnancement optimal sous forme de liste numérotée. "
	prompt += "Indique le makespan total. "
	prompt += "Commence ta réponse par l'ordre d'exécution."

	resp, err := ollama.DoGenerate(ctx, baseURL, model, prompt, 0.0)
	if err != nil {
		return "", fmt.Errorf("one-shot: %w", err)
	}

	tokens.Add(resp.PromptEvalCount, resp.EvalCount)
	text := strings.TrimSpace(resp.Response)

	if verbose {
		log.Printf("[one-shot] Réponse:\n%s\n", text)
	}

	return text, nil
}

// RunMCTSReasoning résout le problème via le reasoning package + MCTS.
func RunMCTSReasoning(ctx context.Context, baseURL, model string, problem problems.Problem, iterations int, tokens *TokenStats) (string, error) {
	gen := &modelGenerator{baseURL: baseURL, model: model, tokens: tokens}
	judge := &modelJudge{baseURL: baseURL, model: model, tokens: tokens}

	question := problem.FormatPrompt()
	criterion := "Proposer un ordonnancement complet respectant toutes les dépendances et minimisant le temps total (makespan). Conclure avec CONCLUSION: suivi de l'ordre et du makespan."

	state := reasoning.New(
		ctx,
		question,
		criterion,
		gen,
		reasoning.WithMaxDepth(6),
		reasoning.WithBranchFactor(3),
	)

	eval := reasoning.NewEvaluator(ctx, judge)
	m := mcts.NewAlphaMCTS(eval, 1.5)

	current := state
	step := 0
	for current.Evaluate() == decision.Undecided {
		step++
		result := m.RunMCTS(current, iterations)
		next, ok := result.(*reasoning.State)
		if !ok || next == current {
			break
		}
		current = next

		if verbose {
			steps := current.Steps()
			if len(steps) > 0 {
				log.Printf("[MCTS] Étape %d choisie: %s", step, steps[len(steps)-1])
			}
		}
	}

	// Construire la réponse à partir des étapes
	var b strings.Builder
	for _, step := range current.Steps() {
		b.WriteString(step)
		b.WriteString("\n")
	}
	return b.String(), nil
}

// modelGenerator implémente reasoning.Generator via Ollama.
type modelGenerator struct {
	baseURL string
	model   string
	tokens  *TokenStats
}

func (g *modelGenerator) Generate(ctx context.Context, prompt string, n int) ([]string, error) {
	candidates := make([]string, 0, n)
	for range n {
		resp, err := ollama.DoGenerate(ctx, g.baseURL, g.model, prompt, 0.8)
		if err != nil {
			continue
		}
		g.tokens.Add(resp.PromptEvalCount, resp.EvalCount)
		text := strings.TrimSpace(resp.Response)
		if text != "" {
			candidates = append(candidates, text)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("aucune réponse de %s", g.model)
	}
	return candidates, nil
}

// modelJudge implémente reasoning.Judge via Ollama.
type modelJudge struct {
	baseURL string
	model   string
	tokens  *TokenStats
}

func (j *modelJudge) Score(ctx context.Context, prompt string) (float64, error) {
	scoringPrompt := prompt + "\n\nÉvalue la qualité de ce raisonnement d'ordonnancement sur une échelle de 0 à 1. Réponds uniquement par un nombre décimal."

	resp, err := ollama.DoGenerate(ctx, j.baseURL, j.model, scoringPrompt, 0.0)
	if err != nil {
		return 0.5, err
	}

	j.tokens.Add(resp.PromptEvalCount, resp.EvalCount)
	text := strings.TrimSpace(resp.Response)

	if verbose {
		log.Printf("[judge-mcts] Score brut: %s", text)
	}

	score, err := ollama.ParseScore(text)
	if err != nil {
		return 0.5, err
	}
	return score, nil
}
