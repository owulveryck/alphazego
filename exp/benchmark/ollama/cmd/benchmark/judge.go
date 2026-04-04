package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/owulveryck/alphazego/exp/benchmark/ollama"
	"github.com/owulveryck/alphazego/exp/benchmark/problems"
)

// Result contient le résultat d'une exécution.
type Result struct {
	Problem  problems.Problem
	Config   Config
	Answer   string
	Score    float64
	Verdict  string
	Error    error
	Duration time.Duration
	Tokens   *TokenStats
}

// Judge évalue une réponse en utilisant le modèle Ollama local.
func Judge(ctx context.Context, baseURL, model string, problem problems.Problem, answer string, tokens *TokenStats) (float64, string, error) {
	prompt := fmt.Sprintf(`Tu es un évaluateur expert en ordonnancement de tâches.

Problème :
%s

Solution de référence : le makespan optimal est de %d jours.

Solution proposée :
%s

Évalue cette solution selon ces critères :
1. Toutes les dépendances sont-elles respectées ? (une tâche ne commence qu'après la fin de ses dépendances)
2. Le makespan proposé est-il correct ?
3. Le makespan est-il optimal ou proche de l'optimal ?

Règles de scoring :
- 1.0 : toutes les dépendances respectées ET makespan optimal (%d jours)
- 0.5 : toutes les dépendances respectées MAIS makespan non optimal
- 0.0 : au moins une dépendance violée OU réponse incompréhensible

Réponds EXACTEMENT au format suivant (2 lignes) :
VERDICT: <explication courte>
SCORE: <0.0 ou 0.5 ou 1.0>`,
		problem.FormatPrompt(), problem.Optimal, answer, problem.Optimal)

	resp, err := ollama.DoGenerate(ctx, baseURL, model, prompt, 0.0)
	if err != nil {
		return 0, "", fmt.Errorf("judge: %w", err)
	}

	tokens.Add(resp.PromptEvalCount, resp.EvalCount)
	text := strings.TrimSpace(resp.Response)

	if verbose {
		log.Printf("[judge] Verdict complet:\n%s", text)
	}

	score, verdict := parseJudgeResponse(text)
	return score, verdict, nil
}

// parseJudgeResponse extrait le score et le verdict de la réponse du juge.
func parseJudgeResponse(text string) (float64, string) {
	lines := strings.Split(text, "\n")
	var verdict string
	var score float64

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "VERDICT:") {
			verdict = strings.TrimSpace(strings.TrimPrefix(line, "VERDICT:"))
		}
		if strings.HasPrefix(line, "SCORE:") {
			scoreStr := strings.TrimSpace(strings.TrimPrefix(line, "SCORE:"))
			if _, err := fmt.Sscanf(scoreStr, "%f", &score); err != nil {
				score = 0
			}
		}
	}

	return score, verdict
}
