package wardley

import (
	"strings"
	"testing"
)

func TestMoveDescriptionEvolve(t *testing.T) {
	comps := []Component{
		{Name: "DB", Phase: Custom},
	}
	m := Move{Type: Evolve, Component: "DB"}
	got := moveDescription(m, comps)
	want := `EVOLVE "DB" (Custom → Product)`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMoveDescriptionGameplay(t *testing.T) {
	m := Move{Type: ApplyGameplay, Component: "API", Gameplay: "open-source"}
	got := moveDescription(m, nil)
	want := `GAMEPLAY "open-source" sur "API"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatBatchPolicyPromptContainsMoves(t *testing.T) {
	comps := []Component{
		{Name: "App", Phase: Custom},
		{Name: "DB", Phase: Genesis},
	}
	moves := []Move{
		{Type: Evolve, Component: "App"},
		{Type: Evolve, Component: "DB"},
		{Type: ApplyGameplay, Component: "App", Gameplay: "open-source"},
	}

	prompt := formatBatchPolicyPrompt("wtg2 content", "question?", moves, comps)

	if !strings.Contains(prompt, "1. EVOLVE") {
		t.Error("prompt should contain numbered move 1")
	}
	if !strings.Contains(prompt, "Custom → Product") {
		t.Error("prompt should contain phase transition for App")
	}
	if !strings.Contains(prompt, "Genesis → Custom") {
		t.Error("prompt should contain phase transition for DB")
	}
	if !strings.Contains(prompt, `GAMEPLAY "open-source"`) {
		t.Error("prompt should contain gameplay move")
	}
	if !strings.Contains(prompt, "JSON") {
		t.Error("prompt should ask for JSON output")
	}
}

func TestFormatBatchPolicyPromptContainsSkillContext(t *testing.T) {
	prompt := formatBatchPolicyPrompt("wtg2", "q", []Move{{Type: Evolve, Component: "A"}}, []Component{{Name: "A", Phase: Genesis}})

	if !strings.Contains(prompt, "Wardley") {
		t.Error("prompt should contain skill context")
	}
}
