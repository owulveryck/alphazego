package reasoning_test

import (
	"context"
	"fmt"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/reasoning"
	"github.com/owulveryck/alphazego/mcts"
)

// demoGenerator simule un LLM qui génère des étapes de raisonnement.
// Il retourne des candidats différents selon le contenu du prompt reçu.
type demoGenerator struct{}

func (g *demoGenerator) Generate(_ context.Context, prompt string, n int) ([]string, error) {
	// Premier appel (pas d'étapes dans le prompt) : 3 candidates de premier niveau
	// Appels suivants : une conclusion
	// On distingue par la présence de "Étape" dans le prompt
	if len(prompt) > 0 {
		for _, r := range prompt {
			if r == 'É' { // contient "Étape" → on a déjà des étapes
				return []string{
					"CONCLUSION: La France a 12 fuseaux horaires grâce à ses territoires outre-mer, plus que la Russie (11).",
				}, nil
			}
		}
	}
	return []string{
		"Les plus grands pays en superficie sont la Russie, le Canada et la Chine.",
		"Il faut considérer les territoires outre-mer, pas seulement la métropole.",
		"La Russie s'étend sur 11 fuseaux horaires.",
	}[:n], nil
}

// demoJudge simule un évaluateur qui favorise le chemin considérant
// les territoires outre-mer.
type demoJudge struct{}

func (j *demoJudge) Score(_ context.Context, prompt string) (float64, error) {
	// Favoriser les raisonnements qui mentionnent les territoires outre-mer
	for _, r := range prompt {
		if r == 'o' { // heuristique simpliste : "outre-mer" contient 'o'
			// Chercher "outre-mer" plus précisément
			break
		}
	}
	// Score basé sur des mots-clés
	if contains(prompt, "outre-mer") {
		return 0.9, nil
	}
	if contains(prompt, "CONCLUSION") {
		return 0.8, nil
	}
	if contains(prompt, "superficie") {
		return 0.3, nil
	}
	return 0.5, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstr(s, substr)
}

func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Cet exemple montre comment utiliser le package reasoning avec le MCTS
// pour explorer des chemins de raisonnement et trouver la meilleure réponse
// à une question factuelle.
func Example() {
	ctx := context.Background()
	gen := &demoGenerator{}
	judge := &demoJudge{}

	state := reasoning.New(
		ctx,
		"Quel pays a le plus de fuseaux horaires ?",
		"Identifier le pays avec le plus grand nombre de fuseaux horaires en considérant tous les territoires",
		gen,
		reasoning.WithMaxDepth(4),
		reasoning.WithBranchFactor(3),
	)

	eval := reasoning.NewEvaluator(ctx, judge)
	m := mcts.NewAlphaMCTS(eval, 1.5)

	// Boucle : à chaque itération, le MCTS choisit la meilleure prochaine
	// étape de raisonnement, comme pour le taquin.
	current := state
	for current.Evaluate() == decision.Undecided {
		result := m.RunMCTS(current, 50)
		next, ok := result.(*reasoning.State)
		if !ok || next == current {
			break
		}
		current = next
	}

	fmt.Printf("Étapes de raisonnement :\n")
	for i, step := range current.Steps() {
		fmt.Printf("  %d. %s\n", i+1, step)
	}

	// Output:
	// Étapes de raisonnement :
	//   1. Il faut considérer les territoires outre-mer, pas seulement la métropole.
	//   2. CONCLUSION: La France a 12 fuseaux horaires grâce à ses territoires outre-mer, plus que la Russie (11).
}
