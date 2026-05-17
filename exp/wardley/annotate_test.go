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

func TestGenerateAnnotations(t *testing.T) {
	comps := []Component{
		{Name: "App", Phase: Product},
		{Name: "DB", Phase: Custom},
	}
	s := NewState("test", "question?", comps, nil, 5)

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
	comps := []Component{
		{Name: "App", Phase: Product},
	}
	s := NewState("test", "q", comps, nil, 5)

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
	comps := []Component{
		{Name: "App", Phase: Product},
	}
	s := NewState("test", "q", comps, nil, 5)

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
	comps := []Component{{Name: "App", Phase: Product}}

	annotations, err := parseAnnotations(text, comps)
	if err != nil {
		t.Fatal(err)
	}
	if len(annotations) != 1 {
		t.Fatalf("got %d annotations, want 1", len(annotations))
	}
}

func TestSerializeWithAnnotations(t *testing.T) {
	comps := []Component{
		{Name: "App", Phase: Product},
	}
	s := NewState("test", "q", comps, nil, 5)
	s.SetAnnotations([]Annotation{
		{Kind: "note", Text: "Important note", Target: "App"},
		{Kind: "warning", Text: "Risk here", Target: "App"},
	})

	output := SerializeWTG2(s)

	if !contains(output, `note "Important note" on App`) {
		t.Errorf("output missing note annotation:\n%s", output)
	}
	if !contains(output, `warning "Risk here" on App`) {
		t.Errorf("output missing warning annotation:\n%s", output)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
