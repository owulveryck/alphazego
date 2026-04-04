package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/owulveryck/alphazego/decision"
	"github.com/owulveryck/alphazego/decision/reasoning"
	"github.com/owulveryck/alphazego/exp/benchmark/vertexai"
	"github.com/owulveryck/alphazego/mcts"
)

func main() {
	question := flag.String("question", "Quel pays a le plus de fuseaux horaires ?", "Question à résoudre")
	criterion := flag.String("criterion", "Identifier le pays avec le plus grand nombre de fuseaux horaires en considérant tous les territoires, y compris les territoires outre-mer", "Critère de succès")
	maxDepth := flag.Int("depth", 5, "Profondeur maximale du raisonnement")
	branches := flag.Int("branches", 3, "Nombre de candidats par étape")
	iterations := flag.Int("iterations", 20, "Nombre d'itérations MCTS par étape")
	flag.Parse()

	project := os.Getenv("GCP_PROJECT")
	region := os.Getenv("GCP_REGION")
	if project == "" || region == "" {
		log.Fatal("Variables d'environnement GCP_PROJECT et GCP_REGION requises")
	}

	ctx := context.Background()

	client, err := vertexai.NewClient(ctx, project, region)
	if err != nil {
		log.Fatalf("Erreur de connexion à Vertex AI: %v", err)
	}

	gen := vertexai.NewGenerator(client)
	judge := vertexai.NewJudge(client)

	// Test de connexion : vérifier que les modèles fonctionnent
	fmt.Printf("Projet: %s, Région: %s\n", project, region)
	fmt.Printf("Test du Generator (modèle %s)... ", vertexai.GeneratorModel)
	testResult, err := gen.Generate(ctx, "Dis juste 'ok'.", 1)
	if err != nil {
		log.Fatalf("Échec: %v", err)
	}
	fmt.Printf("OK (%q)\n", testResult[0])

	fmt.Printf("Test du Judge (modèle %s)... ", vertexai.JudgeModel)
	score, err := judge.Score(ctx, "Évalue la qualité de cette phrase : 'Le ciel est bleu'.")
	if err != nil {
		log.Fatalf("Échec: %v", err)
	}
	fmt.Printf("OK (score=%.2f)\n\n", score)

	fmt.Printf("Question : %s\n", *question)
	fmt.Printf("Critère  : %s\n\n", *criterion)

	state := reasoning.New(
		ctx,
		*question,
		*criterion,
		gen,
		reasoning.WithMaxDepth(*maxDepth),
		reasoning.WithBranchFactor(*branches),
	)

	eval := reasoning.NewEvaluator(ctx, judge)
	m := mcts.NewAlphaMCTS(eval, 1.5)

	step := 0
	current := state
	for current.Evaluate() == decision.Undecided {
		fmt.Printf("--- Recherche MCTS (étape %d) ---\n", step+1)
		result := m.RunMCTS(current, *iterations)
		next, ok := result.(*reasoning.State)
		if !ok || next == current {
			fmt.Println("MCTS n'a pas trouvé d'étape suivante.")
			break
		}
		step++
		lastStep := next.Steps()[len(next.Steps())-1]
		fmt.Printf("Étape %d : %s\n\n", step, lastStep)
		current = next
	}

	fmt.Println("=== Résultat ===")
	switch current.Evaluate() {
	case reasoning.Player:
		fmt.Println("Raisonnement terminé avec succès !")
	case decision.Stalemate:
		fmt.Println("Profondeur maximale atteinte sans conclusion.")
	default:
		fmt.Println("Fin inattendue.")
	}

	fmt.Printf("\nRésumé du raisonnement (%d étapes) :\n", len(current.Steps()))
	for i, s := range current.Steps() {
		fmt.Printf("  %d. %s\n", i+1, s)
	}
}
