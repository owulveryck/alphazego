// Explore effectue une exploration stratégique d'une carte Wardley via MCTS + LLM.
//
// Usage :
//
//	wardley-explore -input carte.wtg2 -project mon-projet -region us-central1
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	wardleyexp "github.com/owulveryck/alphazego/exp/wardley"
	"github.com/owulveryck/alphazego/mcts"
	"google.golang.org/genai"
)

func main() {
	inputFile := flag.String("input", "", "fichier WTG2 d'entrée (requis)")
	depth := flag.Int("depth", 5, "profondeur max de l'arbre MCTS")
	iterations := flag.Int("iterations", 100, "nombre d'itérations MCTS par step")
	cpuct := flag.Float64("cpuct", 1.4, "constante d'exploration PUCT")
	project := flag.String("project", "", "projet GCP pour Vertex AI (requis)")
	region := flag.String("region", "us-central1", "région GCP")
	outputWTG2 := flag.String("output-wtg2", "", "fichier de sortie WTG2 (défaut: stdout)")
	outputSVG := flag.String("output-svg", "", "fichier de sortie SVG (optionnel)")
	model := flag.String("model", "gemini-2.5-flash", "modèle Gemini pour l'évaluation")
	outputDir := flag.String("output-dir", "", "dossier de sortie pour les étapes intermédiaires (WTG2 + URLs)")
	branchFactor := flag.Int("branch-factor", 5, "nombre de candidats par expansion")
	flag.Parse()

	if *inputFile == "" || *project == "" {
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	f, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Erreur ouverture %s: %v", *inputFile, err)
	}
	defer f.Close()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  *project,
		Location: *region,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatalf("Erreur création client Vertex AI: %v", err)
	}

	judge := &geminiJudge{client: client, model: *model}

	state, err := wardleyexp.ParseWTG2(f, *depth, judge, ctx, wardleyexp.WithBranchFactor(*branchFactor))
	if err != nil {
		log.Fatalf("Erreur parsing WTG2: %v", err)
	}

	eval := wardleyexp.NewEvaluator(ctx, judge)
	engine := mcts.NewAlphaMCTS(eval, *cpuct)

	fmt.Println("=== Exploration stratégique ===")
	fmt.Printf("Carte : %s\n", state.Title())
	fmt.Printf("Question : %s\n", state.Question())
	fmt.Printf("Modèle : %s | Candidats/expansion : %d\n", *model, *branchFactor)
	fmt.Printf("Profondeur max : %d | Itérations/step : %d | CPUCT : %.1f\n\n", *depth, *iterations, *cpuct)

	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Fatalf("Erreur création dossier %s: %v", *outputDir, err)
		}
		if err := os.WriteFile(filepath.Join(*outputDir, "step_00.wtg2"), []byte(state.WTG2Text()), 0644); err != nil {
			log.Fatalf("Erreur écriture step_00.wtg2: %v", err)
		}
		fmt.Printf("Step initial écrit dans %s/step_00.wtg2\n", *outputDir)
		if url, err := wardleyexp.PlaygroundURL(state); err == nil {
			fmt.Printf("  %s\n\n", url)
		}
	}

	current := state
	for step := 0; step < *depth; step++ {
		if current.Evaluate() != 0 { // decision.Undecided == 0
			break
		}

		fmt.Printf("--- Step %d/%d ---\n", step+1, *depth)

		eval.ResetCounters()
		eval.Progress = func(info wardleyexp.ProgressInfo) {
			fmt.Fprintf(os.Stderr,
				"\r  iter %d/%d | %d appels LLM | candidats: %d | value: %+.2f",
				info.EvalCount, *iterations, info.LLMCalls, info.CandidateCount, info.Value)
		}

		next := engine.RunMCTS(current, *iterations)
		fmt.Fprintln(os.Stderr)

		ws, ok := next.(*wardleyexp.State)
		if !ok {
			fmt.Println("  Pas de move trouvé.")
			break
		}

		fmt.Printf("  => %s\n\n", ws.LastDescription())

		if *outputDir != "" {
			fmt.Fprintf(os.Stderr, "Annotation de l'étape %d...\n", step+1)
			if err := wardleyexp.GenerateAnnotations(ctx, judge, ws); err != nil {
				fmt.Fprintf(os.Stderr, "Avertissement : annotation step %d échouée : %v\n", step+1, err)
			}
			stepFile := filepath.Join(*outputDir, fmt.Sprintf("step_%02d.wtg2", step+1))
			if err := os.WriteFile(stepFile, []byte(ws.WTG2Text()), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Erreur écriture %s: %v\n", stepFile, err)
			} else {
				fmt.Printf("  Step %d écrit dans %s\n", step+1, stepFile)
			}
			if url, err := wardleyexp.PlaygroundURL(ws); err == nil {
				fmt.Printf("  %s\n\n", url)
			}
		}

		current = ws
	}

	fmt.Println("\n--- Séquence de moves ---")
	for i, desc := range current.History() {
		fmt.Printf("  %d. %s\n", i+1, desc)
	}

	if *outputDir == "" {
		fmt.Fprintf(os.Stderr, "\nAnnotation de la carte par le LLM...\n")
		if err := wardleyexp.GenerateAnnotations(ctx, judge, current); err != nil {
			fmt.Fprintf(os.Stderr, "Avertissement : annotation échouée : %v\n", err)
		}
	}

	fmt.Println("\n--- Carte résultante (WTG2) ---")
	fmt.Println(current.WTG2Text())

	if *outputWTG2 != "" {
		if err := os.WriteFile(*outputWTG2, []byte(current.WTG2Text()), 0644); err != nil {
			log.Fatalf("Erreur écriture WTG2: %v", err)
		}
		fmt.Printf("WTG2 écrit dans %s\n", *outputWTG2)
	}

	if *outputSVG != "" {
		svgFile, err := os.Create(*outputSVG)
		if err != nil {
			log.Fatalf("Erreur création SVG: %v", err)
		}
		defer svgFile.Close()

		if err := wardleyexp.RenderSVG(svgFile, current); err != nil {
			log.Fatalf("Erreur rendu SVG: %v", err)
		}
		fmt.Printf("SVG écrit dans %s\n", *outputSVG)
	}

	if url, err := wardleyexp.PlaygroundURL(current); err == nil {
		fmt.Printf("\n--- Playground ---\n%s\n", url)
	}
}

type geminiJudge struct {
	client *genai.Client
	model  string
}

func (j *geminiJudge) Score(ctx context.Context, prompt string) (float64, error) {
	config := &genai.GenerateContentConfig{
		Temperature:    genai.Ptr(float32(0.0)),
		ThinkingConfig: &genai.ThinkingConfig{ThinkingBudget: genai.Ptr(int32(256))},
	}

	scoringPrompt := prompt + "\n\nRéponds uniquement par un nombre décimal entre 0 et 1."

	resp, err := j.client.Models.GenerateContent(ctx, j.model, genai.Text(scoringPrompt), config)
	if err != nil {
		return 0.5, fmt.Errorf("gemini: erreur d'évaluation: %w", err)
	}

	text := extractText(resp)
	score, err := parseScore(text)
	if err != nil {
		return 0.5, fmt.Errorf("gemini: impossible de parser le score %q: %w", text, err)
	}

	return score, nil
}

func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return ""
	}
	var result string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			result += part.Text
		}
	}
	return result
}

func parseScore(text string) (float64, error) {
	var score float64
	for _, word := range splitWords(text) {
		if _, err := fmt.Sscanf(word, "%f", &score); err == nil {
			return clamp(score), nil
		}
	}
	return 0, fmt.Errorf("aucun nombre trouvé dans %q", text)
}

func splitWords(s string) []string {
	var words []string
	current := ""
	for _, c := range s {
		if c == ' ' || c == '\n' || c == '\t' || c == ',' || c == ';' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (j *geminiJudge) Annotate(ctx context.Context, prompt string) (string, error) {
	config := &genai.GenerateContentConfig{
		Temperature:    genai.Ptr(float32(0.3)),
		ThinkingConfig: &genai.ThinkingConfig{ThinkingBudget: genai.Ptr(int32(1024))},
	}

	resp, err := j.client.Models.GenerateContent(ctx, j.model, genai.Text(prompt), config)
	if err != nil {
		return "", fmt.Errorf("gemini annotate: %w", err)
	}

	return extractText(resp), nil
}

func (j *geminiJudge) Propose(ctx context.Context, prompt string, n int) ([]wardleyexp.Candidate, error) {
	config := &genai.GenerateContentConfig{
		Temperature:    genai.Ptr(float32(0.7)),
		ThinkingConfig: &genai.ThinkingConfig{ThinkingBudget: genai.Ptr(int32(2048))},
	}

	resp, err := j.client.Models.GenerateContent(ctx, j.model, genai.Text(prompt), config)
	if err != nil {
		return nil, fmt.Errorf("gemini propose: %w", err)
	}

	text := extractText(resp)
	return parseCandidates(text)
}

func parseCandidates(text string) ([]wardleyexp.Candidate, error) {
	start := strings.Index(text, "[")
	if start == -1 {
		return nil, fmt.Errorf("pas de tableau JSON trouvé dans la réponse")
	}
	end := strings.LastIndex(text, "]")
	if end == -1 || end <= start {
		return nil, fmt.Errorf("tableau JSON mal formé")
	}

	var raw []struct {
		Description string  `json:"description"`
		WTG2        string  `json:"wtg2"`
		Confidence  float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &raw); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	candidates := make([]wardleyexp.Candidate, 0, len(raw))
	for _, r := range raw {
		if r.WTG2 == "" {
			continue
		}
		candidates = append(candidates, wardleyexp.Candidate{
			Description: r.Description,
			WTG2:        r.WTG2,
			Confidence:  clamp(r.Confidence),
		})
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("aucun candidat valide dans la réponse")
	}

	return candidates, nil
}
