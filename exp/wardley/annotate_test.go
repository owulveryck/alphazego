package wardley

import (
	"context"
	"testing"
)

type mockAnnotator struct {
	response string
}

func (m *mockAnnotator) Annotate(_ context.Context, _ string) (string, error) {
	return m.response, nil
}

const testWTG2 = `title: test
question: "question?"
stages: Genesis, Custom, Product, Commodity

App : III.5
DB : II.5

App -> DB
`

func TestGenerateAnnotations(t *testing.T) {
	proposer := &mockProposerInternal{}
	s := NewState(testWTG2, "test", "question?", 5, proposer, context.Background())

	annotator := &mockAnnotator{
		response: `[{"kind":"note","text":"Evolved to Product","target":"App"},{"kind":"warning","text":"Still in Custom","target":"DB"}]`,
	}

	err := GenerateAnnotations(context.Background(), annotator, s)
	if err != nil {
		t.Fatal(err)
	}

	annotations := s.Annotations()
	if len(annotations) != 2 {
		t.Fatalf("got %d annotations, want 2", len(annotations))
	}

	if annotations[0].Kind != "note" || annotations[0].Target != "App" {
		t.Errorf("annotation[0] = %+v, want note on App", annotations[0])
	}
	if annotations[1].Kind != "warning" || annotations[1].Target != "DB" {
		t.Errorf("annotation[1] = %+v, want warning on DB", annotations[1])
	}
}

func TestGenerateAnnotationsSkipsUnknownTargets(t *testing.T) {
	proposer := &mockProposerInternal{}
	s := NewState(testWTG2, "test", "q", 5, proposer, context.Background())

	annotator := &mockAnnotator{
		response: `[{"kind":"note","text":"Valid","target":"App"},{"kind":"note","text":"Invalid","target":"NonExistent"}]`,
	}

	err := GenerateAnnotations(context.Background(), annotator, s)
	if err != nil {
		t.Fatal(err)
	}

	if len(s.Annotations()) != 1 {
		t.Errorf("got %d annotations, want 1 (unknown target filtered)", len(s.Annotations()))
	}
}

func TestGenerateAnnotationsSkipsEmptyText(t *testing.T) {
	proposer := &mockProposerInternal{}
	s := NewState(testWTG2, "test", "q", 5, proposer, context.Background())

	annotator := &mockAnnotator{
		response: `[{"kind":"note","text":"","target":"App"},{"kind":"note","text":"Valid","target":"App"}]`,
	}

	err := GenerateAnnotations(context.Background(), annotator, s)
	if err != nil {
		t.Fatal(err)
	}

	if len(s.Annotations()) != 1 {
		t.Errorf("got %d annotations, want 1 (empty text filtered)", len(s.Annotations()))
	}
}

func TestParseAnnotationsWithSurroundingText(t *testing.T) {
	text := `Voici les annotations :
[{"kind":"note","text":"Annotation","target":"App"}]
Fin.`
	names := map[string]bool{"App": true}

	annotations, err := parseAnnotations(text, names)
	if err != nil {
		t.Fatal(err)
	}
	if len(annotations) != 1 {
		t.Fatalf("got %d annotations, want 1", len(annotations))
	}
}

type mockProposerInternal struct{}

func (p *mockProposerInternal) Propose(_ context.Context, _ string, _ int) ([]Candidate, error) {
	return nil, nil
}
