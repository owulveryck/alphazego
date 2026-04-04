package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/owulveryck/alphazego/benchmark/problems"
)

func main() {
	model := flag.String("model", "", "Modèle Ollama à utiliser (requis, ex: qwen2.5:7b)")
	baseURL := flag.String("url", "http://localhost:11434", "URL du serveur Ollama")
	problemFilter := flag.String("problem", "", "Filtrer par nom de problème (substring)")
	configFilter := flag.String("config", "", "Filtrer par config: E, F")
	flag.BoolVar(&verbose, "v", false, "Mode verbose : afficher les logs détaillés")
	flag.Parse()

	if *model == "" {
		log.Fatal("Le flag -model est requis (ex: -model qwen2.5:7b)")
	}

	// Vérifier qu'Ollama est accessible
	resp, err := http.Get(*baseURL + "/api/tags")
	if err != nil {
		log.Fatalf("Impossible de contacter Ollama à %s: %v", *baseURL, err)
	}
	resp.Body.Close()

	ctx := context.Background()

	allProblems := problems.All()
	configs := AllConfigs()

	// Filtrer si demandé
	if *problemFilter != "" {
		var filtered []problems.Problem
		for _, p := range allProblems {
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(*problemFilter)) {
				filtered = append(filtered, p)
			}
		}
		allProblems = filtered
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

	fmt.Printf("Benchmark Ollama : %d problèmes × %d configurations\n", len(allProblems), len(configs))
	fmt.Printf("Modèle: %s, URL: %s\n\n", *model, *baseURL)

	// Matrice de résultats [problème][config]
	results := make([][]Result, len(allProblems))
	for i := range results {
		results[i] = make([]Result, len(configs))
	}

	for i, problem := range allProblems {
		fmt.Printf("━━━ %d/%d : %s (%d tâches, optimal=%d) ━━━\n",
			i+1, len(allProblems), problem.Name, len(problem.Tasks), problem.Optimal)

		for j, config := range configs {
			fmt.Printf("  %s ... ", config.Name)

			tokens := &TokenStats{}
			start := time.Now()

			var answer string
			var runErr error

			if config.UseMCTS {
				answer, runErr = RunMCTSReasoning(ctx, *baseURL, *model, problem, config.Iterations, tokens)
			} else {
				answer, runErr = RunOneShot(ctx, *baseURL, *model, problem, tokens)
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

			// Évaluer avec le juge (même modèle local)
			score, verdict, judgeErr := Judge(ctx, *baseURL, *model, problem, answer, tokens)
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
	printReport(allProblems, configs, results)
}

func printReport(allProblems []problems.Problem, configs []Config, results [][]Result) {
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
	totalTokens := make([]int, len(configs))
	totalDuration := make([]time.Duration, len(configs))

	for i, problem := range allProblems {
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
func formatTokenCount(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
