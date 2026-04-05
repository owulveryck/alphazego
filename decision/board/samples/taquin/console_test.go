package taquin

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"
)

func TestTaquin_String(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	rng := rand.New(rand.NewSource(42))
	taq.Shuffle(5, rng)
	s := taq.String()
	if len(s) == 0 {
		t.Fatal("String() should not return empty")
	}
	// Vérifier la présence des caractères de grille.
	if !strings.Contains(s, "┌") || !strings.Contains(s, "┘") {
		t.Error("String() should contain grid border characters")
	}
	// Vérifier que le compteur de steps est affiché.
	if !strings.Contains(s, "Steps:") {
		t.Error("String() should contain 'Steps:' header")
	}
}

func TestTaquin_String_Solved(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	s := taq.String()
	if len(s) == 0 {
		t.Fatal("String() should not return empty")
	}
	// Un taquin résolu affiche "Steps: 0".
	if !strings.Contains(s, "Steps: 0") {
		t.Error("solved taquin should display 'Steps: 0'")
	}
}

func TestTaquin_String_LargeGrid(t *testing.T) {
	taq := NewTaquin(4, 3, 50)
	rng := rand.New(rand.NewSource(42))
	taq.Shuffle(10, rng)
	s := taq.String()
	t.Log("\n" + s)

	// Vérifier que toutes les lignes de la grille ont la même longueur visible.
	// On ignore la première ligne (en-tête "Steps: ...").
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatal("expected multiple lines")
	}
	// Supprimer les séquences ANSI pour comparer les largeurs visibles.
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	gridLines := lines[1:]
	refLen := len([]rune(ansi.ReplaceAllString(gridLines[0], "")))
	for i, line := range gridLines {
		l := len([]rune(ansi.ReplaceAllString(line, "")))
		if l != refLen {
			t.Errorf("line %d: visible length %d, want %d\n  got:  %q", i+1, l, refLen, line)
		}
	}
}
