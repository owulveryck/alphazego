package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/reasoning"
	"github.com/owulveryck/alphazego/mcts"
	"github.com/owulveryck/alphazego/vertexai"

	"google.golang.org/genai"
)

// Config décrit une configuration de benchmark.
type Config struct {
	Name       string
	Model      string
	UseMCTS    bool
	Iterations int
}

// AllConfigs retourne les 4 configurations du benchmark.
func AllConfigs() []Config {
	return []Config{
		{Name: "A (flash-lite)", Model: vertexai.JudgeModel, UseMCTS: false},
		{Name: "B (flash-lite+MCTS)", Model: vertexai.JudgeModel, UseMCTS: true, Iterations: 15},
		{Name: "C (pro)", Model: vertexai.GeneratorModel, UseMCTS: false},
		{Name: "D (pro+MCTS)", Model: vertexai.GeneratorModel, UseMCTS: true, Iterations: 15},
	}
}

// RunOneShot résout le problème en un seul appel au modèle.
func RunOneShot(ctx context.Context, client *genai.Client, model string, problem Problem) (string, error) {
	prompt := problem.FormatPrompt()
	prompt += "\nDonne l'ordonnancement optimal sous forme de liste numérotée. "
	prompt += "Indique le makespan total. "
	prompt += "Commence ta réponse par l'ordre d'exécution."

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.0)),
	}

	resp, err := client.Models.GenerateContent(ctx, model, genai.Text(prompt), config)
	if err != nil {
		return "", fmt.Errorf("one-shot %s: %w", model, err)
	}

	return extractText(resp), nil
}

// RunMCTSReasoning résout le problème via le reasoning package + MCTS.
func RunMCTSReasoning(ctx context.Context, client *genai.Client, model string, problem Problem, iterations int) (string, error) {
	gen := &modelGenerator{client: client, model: model}
	judge := &modelJudge{client: client, model: model}

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
	for current.Evaluate() == decision.Undecided {
		result := m.RunMCTS(current, iterations)
		next, ok := result.(*reasoning.State)
		if !ok || next == current {
			break
		}
		current = next
	}

	// Construire la réponse à partir des étapes
	var b strings.Builder
	for _, step := range current.Steps() {
		b.WriteString(step)
		b.WriteString("\n")
	}
	return b.String(), nil
}

// modelGenerator implémente reasoning.Generator avec un modèle configurable.
type modelGenerator struct {
	client *genai.Client
	model  string
}

func (g *modelGenerator) Generate(ctx context.Context, prompt string, n int) ([]string, error) {
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.8)),
	}

	candidates := make([]string, 0, n)
	for range n {
		resp, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), config)
		if err != nil {
			continue
		}
		text := extractText(resp)
		if text != "" {
			candidates = append(candidates, text)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("aucune réponse de %s", g.model)
	}
	return candidates, nil
}

// modelJudge implémente reasoning.Judge avec un modèle configurable.
type modelJudge struct {
	client *genai.Client
	model  string
}

func (j *modelJudge) Score(ctx context.Context, prompt string) (float64, error) {
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.0)),
	}

	scoringPrompt := prompt + "\n\nÉvalue la qualité de ce raisonnement d'ordonnancement sur une échelle de 0 à 1. Réponds uniquement par un nombre décimal."

	resp, err := j.client.Models.GenerateContent(ctx, j.model, genai.Text(scoringPrompt), config)
	if err != nil {
		return 0.5, err
	}

	text := extractText(resp)
	score, err := parseScore(text)
	if err != nil {
		return 0.5, err
	}
	return score, nil
}

// extractText extrait le texte de la réponse genai.
func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return ""
	}
	var b strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			b.WriteString(part.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

// parseScore extrait un score float depuis du texte.
func parseScore(text string) (float64, error) {
	text = strings.TrimSpace(text)
	var score float64
	if _, err := fmt.Sscanf(text, "%f", &score); err == nil {
		return clamp01(score), nil
	}
	for _, field := range strings.Fields(text) {
		field = strings.Trim(field, ".,;:()[]")
		if _, err := fmt.Sscanf(field, "%f", &score); err == nil {
			return clamp01(score), nil
		}
	}
	return 0, fmt.Errorf("pas de nombre dans %q", text)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
