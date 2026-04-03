// Package vertexai fournit des implémentations de [reasoning.Generator] et
// [reasoning.Judge] utilisant Google Vertex AI (Gemini).
//
// Le [Generator] utilise gemini-3.1-pro-preview pour générer des étapes de
// raisonnement candidates. Le [Judge] utilise gemini-3.1-flash-lite-preview
// (avec thinking level "low") pour évaluer la qualité des raisonnements.
//
// La configuration se fait via les variables d'environnement GCP_PROJECT
// et GCP_REGION, ou directement via les constructeurs.
package vertexai

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/genai"
)

const (
	// GeneratorModel est le modèle utilisé pour générer les étapes de raisonnement.
	GeneratorModel = "gemini-3.1-pro-preview"
	// JudgeModel est le modèle utilisé pour évaluer les raisonnements.
	JudgeModel = "gemini-3.1-flash-lite-preview"
)

// Generator implémente [reasoning.Generator] en utilisant Vertex AI.
// Il génère des étapes de raisonnement candidates via le modèle Gemini Pro.
type Generator struct {
	client *genai.Client
	model  string
}

// NewGenerator crée un Generator utilisant le client Vertex AI donné.
func NewGenerator(client *genai.Client) *Generator {
	return &Generator{
		client: client,
		model:  GeneratorModel,
	}
}

// Generate produit n étapes de raisonnement candidates à partir du prompt.
// Les appels sont parallélisés pour minimiser la latence.
func (g *Generator) Generate(ctx context.Context, prompt string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.8)),
	}

	type result struct {
		idx  int
		text string
		err  error
	}

	results := make(chan result, n)
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resp, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), config)
			if err != nil {
				results <- result{idx: idx, err: err}
				return
			}
			text := extractText(resp)
			results <- result{idx: idx, text: text}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	candidates := make([]string, n)
	var firstErr error
	for r := range results {
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
			}
			continue
		}
		candidates[r.idx] = r.text
	}

	// Filtrer les résultats vides
	filtered := make([]string, 0, n)
	for _, c := range candidates {
		if c != "" {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("vertexai: aucune réponse générée: %w", firstErr)
	}

	return filtered, nil
}

// Judge implémente [reasoning.Judge] en utilisant Vertex AI.
// Il évalue la qualité d'un raisonnement via le modèle Gemini Flash-Lite
// avec thinking level "low".
type Judge struct {
	client *genai.Client
	model  string
}

// NewJudge crée un Judge utilisant le client Vertex AI donné.
func NewJudge(client *genai.Client) *Judge {
	return &Judge{
		client: client,
		model:  JudgeModel,
	}
}

// Score évalue un raisonnement et retourne un score dans [0, 1].
// Le prompt doit demander au modèle de retourner un score numérique.
// Le Judge extrait le premier nombre flottant trouvé dans la réponse.
func (j *Judge) Score(ctx context.Context, prompt string) (float64, error) {
	config := &genai.GenerateContentConfig{
		Temperature:    genai.Ptr(float32(0.0)),
		ThinkingConfig: &genai.ThinkingConfig{ThinkingBudget: genai.Ptr(int32(256))},
	}

	scoringPrompt := prompt + "\n\nRéponds uniquement par un nombre décimal entre 0 et 1."

	resp, err := j.client.Models.GenerateContent(ctx, j.model, genai.Text(scoringPrompt), config)
	if err != nil {
		return 0.5, fmt.Errorf("vertexai: erreur d'évaluation: %w", err)
	}

	text := extractText(resp)
	score, err := parseScore(text)
	if err != nil {
		return 0.5, fmt.Errorf("vertexai: impossible de parser le score %q: %w", text, err)
	}

	return score, nil
}

// NewClient crée un client Vertex AI avec le projet et la région donnés.
func NewClient(ctx context.Context, project, region string) (*genai.Client, error) {
	return genai.NewClient(ctx, &genai.ClientConfig{
		Project:  project,
		Location: region,
		Backend:  genai.BackendVertexAI,
	})
}

// extractText extrait le texte de la première réponse candidate.
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

// parseScore extrait un score flottant dans [0, 1] depuis une réponse texte.
func parseScore(text string) (float64, error) {
	text = strings.TrimSpace(text)
	// Essayer de parser directement
	if score, err := strconv.ParseFloat(text, 64); err == nil {
		return clamp(score), nil
	}

	// Chercher un nombre flottant dans le texte
	fields := strings.Fields(text)
	for _, field := range fields {
		field = strings.Trim(field, ".,;:()[]")
		if score, err := strconv.ParseFloat(field, 64); err == nil {
			return clamp(score), nil
		}
	}

	return 0, fmt.Errorf("aucun nombre trouvé dans %q", text)
}

// clamp limite une valeur dans [0, 1].
func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
