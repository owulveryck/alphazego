package problems_test

import (
	"fmt"
	"strings"

	"github.com/owulveryck/alphazego/benchmark/problems"
)

func ExampleAll() {
	probs := problems.All()
	fmt.Printf("%d problèmes\n", len(probs))
	fmt.Printf("Premier: %s (%d tâches, optimal=%d)\n",
		probs[0].Name, len(probs[0].Tasks), probs[0].Optimal)
	fmt.Printf("Dernier: %s (%d tâches, optimal=%d)\n",
		probs[len(probs)-1].Name, len(probs[len(probs)-1].Tasks), probs[len(probs)-1].Optimal)
	// Output:
	// 10 problèmes
	// Premier: Chaîne linéaire (4 tâches, optimal=8)
	// Dernier: Lancement produit (12 tâches, optimal=25)
}

func ExampleProblem_FormatPrompt() {
	p := problems.All()[0] // Chaîne linéaire
	prompt := p.FormatPrompt()
	// Le prompt contient le nom, les tâches et les règles
	fmt.Println(strings.Contains(prompt, "Chaîne linéaire"))
	fmt.Println(strings.Contains(prompt, "durée: 2 jours"))
	fmt.Println(strings.Contains(prompt, "makespan"))
	// Output:
	// true
	// true
	// true
}
