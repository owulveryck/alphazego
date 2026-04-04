package vertexai

import (
	"testing"

	"google.golang.org/genai"
)

func TestParseScore_Direct(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"0.5", 0.5},
		{"0.0", 0.0},
		{"1.0", 1.0},
		{"  0.75  ", 0.75},
		{"0.123", 0.123},
	}
	for _, tt := range tests {
		got, err := parseScore(tt.input)
		if err != nil {
			t.Errorf("parseScore(%q) erreur: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseScore(%q) = %f, attendu %f", tt.input, got, tt.want)
		}
	}
}

func TestParseScore_InText(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"Le score est 0.8 sur 1", 0.8},
		{"Score: 0.5", 0.5},
		{"Je donne 0.9.", 0.9},
	}
	for _, tt := range tests {
		got, err := parseScore(tt.input)
		if err != nil {
			t.Errorf("parseScore(%q) erreur: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseScore(%q) = %f, attendu %f", tt.input, got, tt.want)
		}
	}
}

func TestParseScore_Clamp(t *testing.T) {
	got, err := parseScore("1.5")
	if err != nil {
		t.Fatalf("parseScore(\"1.5\") erreur: %v", err)
	}
	if got != 1.0 {
		t.Errorf("parseScore(\"1.5\") = %f, attendu 1.0 (clamped)", got)
	}

	got, err = parseScore("-0.3")
	if err != nil {
		t.Fatalf("parseScore(\"-0.3\") erreur: %v", err)
	}
	if got != 0.0 {
		t.Errorf("parseScore(\"-0.3\") = %f, attendu 0.0 (clamped)", got)
	}
}

func TestParseScore_Error(t *testing.T) {
	_, err := parseScore("pas un nombre")
	if err == nil {
		t.Error("parseScore(\"pas un nombre\") devrait retourner une erreur")
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		input, want float64
	}{
		{-1.0, 0.0},
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{2.0, 1.0},
	}
	for _, tt := range tests {
		if got := clamp(tt.input); got != tt.want {
			t.Errorf("clamp(%f) = %f, attendu %f", tt.input, got, tt.want)
		}
	}
}

func TestExtractText_Nil(t *testing.T) {
	if got := extractText(nil); got != "" {
		t.Errorf("extractText(nil) = %q, attendu \"\"", got)
	}
}

func TestExtractText_NoCandidates(t *testing.T) {
	resp := &genai.GenerateContentResponse{}
	if got := extractText(resp); got != "" {
		t.Errorf("extractText(pas de candidats) = %q, attendu \"\"", got)
	}
}

func TestExtractText_EmptyContent(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: nil},
		},
	}
	if got := extractText(resp); got != "" {
		t.Errorf("extractText(content nil) = %q, attendu \"\"", got)
	}
}

func TestExtractText_WithText(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "Bonjour "},
						{Text: "le monde"},
					},
				},
			},
		},
	}
	if got := extractText(resp); got != "Bonjour le monde" {
		t.Errorf("extractText = %q, attendu \"Bonjour le monde\"", got)
	}
}

func TestNewGenerator(t *testing.T) {
	// NewGenerator ne peut pas être testé sans client réel,
	// mais on vérifie que le modèle est correctement assigné.
	g := NewGenerator(nil)
	if g.model != GeneratorModel {
		t.Errorf("model = %q, attendu %q", g.model, GeneratorModel)
	}
}

func TestNewJudge(t *testing.T) {
	j := NewJudge(nil)
	if j.model != JudgeModel {
		t.Errorf("model = %q, attendu %q", j.model, JudgeModel)
	}
}

func TestGenerator_Generate_ZeroN(t *testing.T) {
	g := NewGenerator(nil)
	candidates, err := g.Generate(t.Context(), "question", 0)
	if err != nil {
		t.Fatalf("Generate(n=0) erreur: %v", err)
	}
	if candidates != nil {
		t.Errorf("Generate(n=0) = %v, attendu nil", candidates)
	}
}
