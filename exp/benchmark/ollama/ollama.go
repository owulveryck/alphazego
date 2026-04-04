package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Generator implémente [reasoning.Generator] en utilisant Ollama.
// Il génère des étapes de raisonnement candidates via un modèle local.
type Generator struct {
	BaseURL string
	Model   string
}

// NewGenerator crée un Generator pour le modèle Ollama donné.
func NewGenerator(baseURL, model string) *Generator {
	return &Generator{BaseURL: baseURL, Model: model}
}

// Generate produit n étapes de raisonnement candidates à partir du prompt.
// Les appels sont séquentiels (le GPU local est le goulot d'étranglement).
func (g *Generator) Generate(ctx context.Context, prompt string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}

	candidates := make([]string, 0, n)
	for range n {
		resp, err := doGenerate(ctx, g.BaseURL, g.Model, prompt, 0.8)
		if err != nil {
			continue
		}
		if resp.Response != "" {
			candidates = append(candidates, strings.TrimSpace(resp.Response))
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("ollama: aucune réponse de %s", g.Model)
	}
	return candidates, nil
}

// Judge implémente [reasoning.Judge] en utilisant Ollama.
// Il évalue la qualité d'un raisonnement via un modèle local.
type Judge struct {
	BaseURL string
	Model   string
}

// NewJudge crée un Judge pour le modèle Ollama donné.
func NewJudge(baseURL, model string) *Judge {
	return &Judge{BaseURL: baseURL, Model: model}
}

// Score évalue un raisonnement et retourne un score dans [0, 1].
func (j *Judge) Score(ctx context.Context, prompt string) (float64, error) {
	scoringPrompt := prompt + "\n\nRéponds uniquement par un nombre décimal entre 0 et 1."

	resp, err := doGenerate(ctx, j.BaseURL, j.Model, scoringPrompt, 0.0)
	if err != nil {
		return 0.5, fmt.Errorf("ollama: erreur d'évaluation: %w", err)
	}

	score, err := ParseScore(resp.Response)
	if err != nil {
		return 0.5, fmt.Errorf("ollama: impossible de parser le score %q: %w", resp.Response, err)
	}
	return score, nil
}

// generateRequest est le corps JSON de la requête POST /api/generate.
type generateRequest struct {
	Model   string          `json:"model"`
	Prompt  string          `json:"prompt"`
	Stream  bool            `json:"stream"`
	Options generateOptions `json:"options"`
}

// generateOptions contient les options du modèle.
type generateOptions struct {
	Temperature float64 `json:"temperature"`
}

// GenerateResponse est la réponse JSON de POST /api/generate.
// Exporté pour permettre au benchmark d'accéder aux compteurs de tokens.
type GenerateResponse struct {
	Response        string `json:"response"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

// DoGenerate effectue un appel à l'API Ollama /api/generate.
// Exporté pour permettre au benchmark d'accéder aux compteurs de tokens.
func DoGenerate(ctx context.Context, baseURL, model, prompt string, temperature float64) (*GenerateResponse, error) {
	return doGenerate(ctx, baseURL, model, prompt, temperature)
}

// doGenerate effectue un appel à l'API Ollama /api/generate.
func doGenerate(ctx context.Context, baseURL, model, prompt string, temperature float64) (*GenerateResponse, error) {
	reqBody := generateRequest{
		Model:   model,
		Prompt:  prompt,
		Stream:  false,
		Options: generateOptions{Temperature: temperature},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: requête: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: appel: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama: status %d: %s", resp.StatusCode, string(respBody))
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("ollama: décodage: %w", err)
	}

	return &genResp, nil
}

// ParseScore extrait un score flottant dans [0, 1] depuis une réponse texte.
func ParseScore(text string) (float64, error) {
	text = strings.TrimSpace(text)
	if score, err := strconv.ParseFloat(text, 64); err == nil {
		return clamp(score), nil
	}

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
