package wardley

import (
	"strings"
	"testing"
)

func TestFormatProposalPromptContainsWTG2(t *testing.T) {
	wtg2 := "App : III.5\nDB : II.5\n"
	prompt := formatProposalPrompt(wtg2, "Should we evolve?", nil, 3)

	if !strings.Contains(prompt, "App : III.5") {
		t.Error("prompt should contain WTG2 text")
	}
	if !strings.Contains(prompt, "Should we evolve?") {
		t.Error("prompt should contain question")
	}
	if !strings.Contains(prompt, "3 modifications") {
		t.Error("prompt should contain candidate count")
	}
	if !strings.Contains(prompt, "JSON") {
		t.Error("prompt should ask for JSON output")
	}
}

func TestFormatProposalPromptContainsHistory(t *testing.T) {
	history := []string{"Evolve DB to Product", "Add open-source on App"}
	prompt := formatProposalPrompt("wtg2", "q", history, 5)

	if !strings.Contains(prompt, "Evolve DB to Product") {
		t.Error("prompt should contain history entry")
	}
	if !strings.Contains(prompt, "1.") {
		t.Error("prompt should number history entries")
	}
}

func TestFormatProposalPromptContainsSkillContext(t *testing.T) {
	prompt := formatProposalPrompt("wtg2", "q", nil, 3)

	if !strings.Contains(prompt, "Wardley") {
		t.Error("prompt should contain skill context")
	}
}

func TestFormatValuePromptContainsWTG2(t *testing.T) {
	prompt := formatValuePrompt("App : III.5", "question?", nil)

	if !strings.Contains(prompt, "App : III.5") {
		t.Error("prompt should contain WTG2 text")
	}
	if !strings.Contains(prompt, "question?") {
		t.Error("prompt should contain question")
	}
}

func TestFormatAnnotationPromptContainsHistory(t *testing.T) {
	history := []string{"Move A", "Move B"}
	prompt := formatAnnotationPrompt("wtg2", "q", history)

	if !strings.Contains(prompt, "Move A") {
		t.Error("prompt should contain history")
	}
	if !strings.Contains(prompt, "JSON") {
		t.Error("prompt should ask for JSON")
	}
}
