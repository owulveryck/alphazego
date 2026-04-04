package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
		got, err := ParseScore(tt.input)
		if err != nil {
			t.Errorf("ParseScore(%q) erreur: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseScore(%q) = %f, attendu %f", tt.input, got, tt.want)
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
		got, err := ParseScore(tt.input)
		if err != nil {
			t.Errorf("ParseScore(%q) erreur: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseScore(%q) = %f, attendu %f", tt.input, got, tt.want)
		}
	}
}

func TestParseScore_Clamp(t *testing.T) {
	got, err := ParseScore("1.5")
	if err != nil {
		t.Fatalf("ParseScore(\"1.5\") erreur: %v", err)
	}
	if got != 1.0 {
		t.Errorf("ParseScore(\"1.5\") = %f, attendu 1.0 (clamped)", got)
	}

	got, err = ParseScore("-0.3")
	if err != nil {
		t.Fatalf("ParseScore(\"-0.3\") erreur: %v", err)
	}
	if got != 0.0 {
		t.Errorf("ParseScore(\"-0.3\") = %f, attendu 0.0 (clamped)", got)
	}
}

func TestParseScore_Error(t *testing.T) {
	_, err := ParseScore("pas un nombre")
	if err == nil {
		t.Error("ParseScore(\"pas un nombre\") devrait retourner une erreur")
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

func TestNewGenerator(t *testing.T) {
	g := NewGenerator("http://localhost:11434", "test-model")
	if g.BaseURL != "http://localhost:11434" {
		t.Errorf("BaseURL = %q, attendu http://localhost:11434", g.BaseURL)
	}
	if g.Model != "test-model" {
		t.Errorf("Model = %q, attendu test-model", g.Model)
	}
}

func TestNewJudge(t *testing.T) {
	j := NewJudge("http://localhost:11434", "test-model")
	if j.BaseURL != "http://localhost:11434" {
		t.Errorf("BaseURL = %q, attendu http://localhost:11434", j.BaseURL)
	}
	if j.Model != "test-model" {
		t.Errorf("Model = %q, attendu test-model", j.Model)
	}
}

func TestDoGenerate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("méthode = %s, attendu POST", r.Method)
		}
		if r.URL.Path != "/api/generate" {
			t.Errorf("path = %s, attendu /api/generate", r.URL.Path)
		}

		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("décodage requête: %v", err)
		}
		if req.Model != "test-model" {
			t.Errorf("model = %q, attendu test-model", req.Model)
		}
		if req.Stream {
			t.Error("stream devrait être false")
		}

		resp := GenerateResponse{
			Response:        "Réponse de test",
			PromptEvalCount: 10,
			EvalCount:       20,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	resp, err := DoGenerate(context.Background(), server.URL, "test-model", "hello", 0.8)
	if err != nil {
		t.Fatalf("DoGenerate erreur: %v", err)
	}
	if resp.Response != "Réponse de test" {
		t.Errorf("Response = %q, attendu 'Réponse de test'", resp.Response)
	}
	if resp.PromptEvalCount != 10 {
		t.Errorf("PromptEvalCount = %d, attendu 10", resp.PromptEvalCount)
	}
	if resp.EvalCount != 20 {
		t.Errorf("EvalCount = %d, attendu 20", resp.EvalCount)
	}
}

func TestDoGenerate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	_, err := DoGenerate(context.Background(), server.URL, "model", "prompt", 0.0)
	if err == nil {
		t.Error("DoGenerate devrait retourner une erreur sur status 500")
	}
}

func TestDoGenerate_ConnectionRefused(t *testing.T) {
	_, err := DoGenerate(context.Background(), "http://localhost:1", "model", "prompt", 0.0)
	if err == nil {
		t.Error("DoGenerate devrait retourner une erreur quand le serveur est injoignable")
	}
}

func TestGenerator_Generate_WithMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GenerateResponse{Response: "étape de raisonnement"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	g := NewGenerator(server.URL, "test")
	candidates, err := g.Generate(context.Background(), "question", 3)
	if err != nil {
		t.Fatalf("Generate erreur: %v", err)
	}
	if len(candidates) != 3 {
		t.Errorf("len(candidates) = %d, attendu 3", len(candidates))
	}
	for i, c := range candidates {
		if c != "étape de raisonnement" {
			t.Errorf("candidates[%d] = %q, attendu 'étape de raisonnement'", i, c)
		}
	}
}

func TestGenerator_Generate_ZeroN(t *testing.T) {
	g := NewGenerator("http://localhost:11434", "test")
	candidates, err := g.Generate(context.Background(), "question", 0)
	if err != nil {
		t.Fatalf("Generate(n=0) erreur: %v", err)
	}
	if candidates != nil {
		t.Errorf("Generate(n=0) = %v, attendu nil", candidates)
	}
}

func TestJudge_Score_WithMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := GenerateResponse{Response: "0.85"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	j := NewJudge(server.URL, "test")
	score, err := j.Score(context.Background(), "évaluer ce raisonnement")
	if err != nil {
		t.Fatalf("Score erreur: %v", err)
	}
	if score != 0.85 {
		t.Errorf("Score = %f, attendu 0.85", score)
	}
}
