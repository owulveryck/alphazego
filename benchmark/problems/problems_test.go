package problems

import (
	"strings"
	"testing"
)

func TestAllProblems_Count(t *testing.T) {
	probs := All()
	if got := len(probs); got != 10 {
		t.Errorf("All() retourne %d problèmes, attendu 10", got)
	}
}

func TestAllProblems_OptimalPositive(t *testing.T) {
	for _, p := range All() {
		if p.Optimal <= 0 {
			t.Errorf("%s: optimal=%d, doit être > 0", p.Name, p.Optimal)
		}
	}
}

func TestAllProblems_HasTasks(t *testing.T) {
	for _, p := range All() {
		if len(p.Tasks) == 0 {
			t.Errorf("%s: aucune tâche définie", p.Name)
		}
	}
}

func TestAllProblems_DependenciesExist(t *testing.T) {
	for _, p := range All() {
		names := make(map[string]bool)
		for _, task := range p.Tasks {
			names[task.Name] = true
		}
		for _, task := range p.Tasks {
			for _, dep := range task.Dependencies {
				if !names[dep] {
					t.Errorf("%s: tâche %q dépend de %q qui n'existe pas",
						p.Name, task.Name, dep)
				}
			}
		}
	}
}

func TestAllProblems_NoCyclicSelfDep(t *testing.T) {
	for _, p := range All() {
		for _, task := range p.Tasks {
			for _, dep := range task.Dependencies {
				if dep == task.Name {
					t.Errorf("%s: tâche %q dépend d'elle-même", p.Name, task.Name)
				}
			}
		}
	}
}

func TestAllProblems_DifficultyIncreasing(t *testing.T) {
	probs := All()
	for i := 1; i < len(probs); i++ {
		if len(probs[i].Tasks) < len(probs[i-1].Tasks) {
			// Pas strictement croissant mais ne devrait pas diminuer fortement
			if len(probs[i].Tasks) < len(probs[i-1].Tasks)-1 {
				t.Errorf("problème %d (%s, %d tâches) a moins de tâches que %d (%s, %d tâches)",
					i+1, probs[i].Name, len(probs[i].Tasks),
					i, probs[i-1].Name, len(probs[i-1].Tasks))
			}
		}
	}
}

func TestFormatPrompt_ContainsTaskNames(t *testing.T) {
	p := All()[0] // Chaîne linéaire
	prompt := p.FormatPrompt()

	for _, task := range p.Tasks {
		if !strings.Contains(prompt, task.Name) {
			t.Errorf("FormatPrompt() ne contient pas le nom de tâche %q", task.Name)
		}
	}
}

func TestFormatPrompt_ContainsRules(t *testing.T) {
	p := All()[0]
	prompt := p.FormatPrompt()

	if !strings.Contains(prompt, "dépendances") {
		t.Error("FormatPrompt() ne mentionne pas les dépendances")
	}
	if !strings.Contains(prompt, "makespan") {
		t.Error("FormatPrompt() ne mentionne pas le makespan")
	}
}

func TestAllProblems_OptimalAtLeastSumCriticalPath(t *testing.T) {
	for _, p := range All() {
		// L'optimal doit être au moins la durée de la tâche la plus longue
		maxDuration := 0
		for _, task := range p.Tasks {
			if task.Duration > maxDuration {
				maxDuration = task.Duration
			}
		}
		if p.Optimal < maxDuration {
			t.Errorf("%s: optimal=%d < durée max tâche=%d",
				p.Name, p.Optimal, maxDuration)
		}
	}
}
