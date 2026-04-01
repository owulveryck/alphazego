package taquin

import (
	"math/rand"
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
	t.Log("\n" + s)
}

func TestTaquin_String_Solved(t *testing.T) {
	taq := NewTaquin(2, 3, 50)
	s := taq.String()
	if len(s) == 0 {
		t.Fatal("String() should not return empty")
	}
	t.Log("\n" + s)
}
