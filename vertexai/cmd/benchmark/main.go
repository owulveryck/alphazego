package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/owulveryck/alphazego/vertexai"
)

func main() {
	problemFilter := flag.String("problem", "", "Filtrer par nom de problème (substring)")
	configFilter := flag.String("config", "", "Filtrer par config: A, B, C, D")
	flag.BoolVar(&verbose, "v", false, "Mode verbose : afficher les logs détaillés")
	flag.Parse()

	project := os.Getenv("GCP_PROJECT")
	region := os.Getenv("GCP_REGION")
	if project == "" || region == "" {
		log.Fatal("Variables d'environnement GCP_PROJECT et GCP_REGION requises")
	}

	ctx := context.Background()
	client, err := vertexai.NewClient(ctx, project, region)
	if err != nil {
		log.Fatalf("Erreur de connexion: %v", err)
	}

	problems := AllProblems()
	configs := AllConfigs()

	// Filtrer si demandé
	if *problemFilter != "" {
		var filtered []Problem
		for _, p := range problems {
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(*problemFilter)) {
				filtered = append(filtered, p)
			}
		}
		problems = filtered
	}
	if *configFilter != "" {
		var filtered []Config
		for _, c := range configs {
			if strings.Contains(c.Name, *configFilter) {
				filtered = append(filtered, c)
			}
		}
		configs = filtered
	}

	fmt.Printf("Benchmark : %d problèmes × %d configurations\n", len(problems), len(configs))
	fmt.Printf("Projet: %s, Région: %s\n\n", project, region)

	// Matrice de résultats [problème][config]
	results := make([][]Result, len(problems))
	for i := range results {
		results[i] = make([]Result, len(configs))
	}

	for i, problem := range problems {
		fmt.Printf("━━━ %d/%d : %s (%d tâches, optimal=%d) ━━━\n",
			i+1, len(problems), problem.Name, len(problem.Tasks), problem.Optimal)

		for j, config := range configs {
			fmt.Printf("  %s ... ", config.Name)

			tokens := &TokenStats{}
			start := time.Now()

			var answer string
			var runErr error

			if config.UseMCTS {
				answer, runErr = RunMCTSReasoning(ctx, client, config.Model, problem, config.Iterations, tokens)
			} else {
				answer, runErr = RunOneShot(ctx, client, config.Model, problem, tokens)
			}

			if runErr != nil {
				elapsed := time.Since(start)
				fmt.Printf("ERREUR: %v (%s)\n", runErr, elapsed.Round(time.Millisecond))
				results[i][j] = Result{
					Problem:  problem,
					Config:   config,
					Error:    runErr,
					Duration: elapsed,
					Tokens:   tokens,
				}
				continue
			}

			// Évaluer avec le juge
			score, verdict, judgeErr := Judge(ctx, client, problem, answer, tokens)
			elapsed := time.Since(start)

			if judgeErr != nil {
				fmt.Printf("ERREUR JUGE: %v (%s)\n", judgeErr, elapsed.Round(time.Millisecond))
				results[i][j] = Result{
					Problem:  problem,
					Config:   config,
					Answer:   answer,
					Error:    judgeErr,
					Duration: elapsed,
					Tokens:   tokens,
				}
				continue
			}

			results[i][j] = Result{
				Problem:  problem,
				Config:   config,
				Answer:   answer,
				Score:    score,
				Verdict:  verdict,
				Duration: elapsed,
				Tokens:   tokens,
			}

			fmt.Printf("score=%.1f (%s) [%s, %d tokens]\n", score, verdict, elapsed.Round(time.Millisecond), tokens.Total())
		}
		fmt.Println()
	}

	// Rapport final
	printReport(problems, configs, results)
}

func printReport(problems []Problem, configs []Config, results [][]Result) {
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("RAPPORT FINAL")
	fmt.Println(strings.Repeat("═", 80))

	// En-tête
	fmt.Printf("%-30s", "Problème")
	for _, c := range configs {
		fmt.Printf(" | %-16s", c.Name)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("─", 30+len(configs)*19))

	// Lignes
	totals := make([]float64, len(configs))
	counts := make([]int, len(configs))
	totalTokens := make([]int32, len(configs))
	totalDuration := make([]time.Duration, len(configs))

	for i, problem := range problems {
		name := problem.Name
		if len(name) > 28 {
			name = name[:28]
		}
		fmt.Printf("%-30s", name)
		for j := range configs {
			r := results[i][j]
			if r.Error != nil {
				fmt.Printf(" | %-16s", "ERR")
			} else {
				fmt.Printf(" | %-16s", fmt.Sprintf("%.1f", r.Score))
				totals[j] += r.Score
				counts[j]++
			}
			if r.Tokens != nil {
				totalTokens[j] += r.Tokens.Total()
			}
			totalDuration[j] += r.Duration
		}
		fmt.Println()
	}

	// Totaux
	fmt.Println(strings.Repeat("─", 30+len(configs)*19))
	fmt.Printf("%-30s", "ACCURACY")
	for j := range configs {
		if counts[j] > 0 {
			pct := totals[j] / float64(counts[j]) * 100
			fmt.Printf(" | %-16s", fmt.Sprintf("%.0f%%", pct))
		} else {
			fmt.Printf(" | %-16s", "N/A")
		}
	}
	fmt.Println()

	// Tokens
	fmt.Printf("%-30s", "TOKENS")
	for j := range configs {
		fmt.Printf(" | %-16s", formatTokenCount(totalTokens[j]))
	}
	fmt.Println()

	// Temps
	fmt.Printf("%-30s", "TEMPS")
	for j := range configs {
		fmt.Printf(" | %-16s", totalDuration[j].Round(time.Millisecond).String())
	}
	fmt.Println()
	fmt.Println(strings.Repeat("═", 80))
}

// formatTokenCount formate un nombre de tokens de manière lisible.
func formatTokenCount(n int32) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
