package wardley

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Annotator génère des annotations explicatives pour une carte Wardley.
// Les implémentations appellent un LLM et retournent du JSON parseable.
type Annotator interface {
	Annotate(ctx context.Context, prompt string) (string, error)
}

// GenerateAnnotations demande au LLM de produire des notes explicatives
// pour la carte résultante. Les annotations sont ajoutées directement à l'état.
func GenerateAnnotations(ctx context.Context, annotator Annotator, s *State) error {
	wtg2Text := SerializeWTG2(s)
	prompt := formatAnnotationPrompt(wtg2Text, s.Question(), s.History(), s.Components())

	text, err := annotator.Annotate(ctx, prompt)
	if err != nil {
		return fmt.Errorf("annotation: %w", err)
	}

	annotations, err := parseAnnotations(text, s.Components())
	if err != nil {
		return fmt.Errorf("annotation parsing: %w", err)
	}

	s.SetAnnotations(annotations)
	return nil
}

func parseAnnotations(text string, components []Component) ([]Annotation, error) {
	start := strings.Index(text, "[")
	if start == -1 {
		return nil, fmt.Errorf("pas de tableau JSON trouvé")
	}
	end := strings.LastIndex(text, "]")
	if end == -1 || end <= start {
		return nil, fmt.Errorf("tableau JSON mal formé")
	}

	var raw []struct {
		Kind   string `json:"kind"`
		Text   string `json:"text"`
		Target string `json:"target"`
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &raw); err != nil {
		return nil, fmt.Errorf("JSON: %w", err)
	}

	compNames := make(map[string]bool, len(components))
	for _, c := range components {
		compNames[c.Name] = true
	}

	var annotations []Annotation
	for _, r := range raw {
		if r.Kind != "note" && r.Kind != "warning" {
			r.Kind = "note"
		}
		if !compNames[r.Target] {
			continue
		}
		if r.Text == "" {
			continue
		}
		annotations = append(annotations, Annotation{
			Kind:   r.Kind,
			Text:   r.Text,
			Target: r.Target,
		})
	}

	return annotations, nil
}
