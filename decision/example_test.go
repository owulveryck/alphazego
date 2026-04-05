package decision_test

import (
	"fmt"

	"github.com/owulveryck/alphazego/decision"
)

// Cet exemple montre les constantes ActorID prédéfinies et leur signification.
func ExampleActorID() {
	fmt.Println("Undecided:", decision.Undecided)
	fmt.Println("Stalemate:", decision.Stalemate)

	// Les acteurs sont des entiers positifs (1, 2, ...).
	actor1 := decision.ActorID(1)
	actor2 := decision.ActorID(2)
	fmt.Println("Actor1:", actor1)
	fmt.Println("Actor2:", actor2)
	// Output:
	// Undecided: 0
	// Stalemate: -1
	// Actor1: 1
	// Actor2: 2
}

// simpleState est un État minimal pour illustrer le contrat de [decision.State].
type simpleState struct {
	value int
	actor decision.ActorID
}

func (s *simpleState) CurrentActor() decision.ActorID  { return s.actor }
func (s *simpleState) PreviousActor() decision.ActorID { return 3 - s.actor }
func (s *simpleState) ID() string                      { return fmt.Sprintf("%d-%d", s.value, s.actor) }
func (s *simpleState) Evaluate() decision.ActorID {
	if s.value >= 3 {
		return decision.ActorID(1) // actor1 gagne
	}
	return decision.Undecided
}
func (s *simpleState) PossibleMoves() []decision.State {
	if s.Evaluate() != decision.Undecided {
		return nil
	}
	return []decision.State{
		&simpleState{value: s.value + 1, actor: 3 - s.actor},
	}
}

// Cet exemple illustre le contrat de [decision.State] : PossibleMoves retourne
// nil pour un état terminal, et des copies indépendantes sinon.
func ExampleState() {
	s := &simpleState{value: 0, actor: 1}

	// État non terminal : PossibleMoves retourne des successeurs.
	moves := s.PossibleMoves()
	fmt.Println("Moves from initial state:", len(moves))
	fmt.Println("Current actor:", s.CurrentActor())

	// Chaque enfant a l'acteur suivant.
	child := moves[0]
	fmt.Println("Child actor:", child.CurrentActor())

	// État terminal : PossibleMoves retourne nil.
	terminal := &simpleState{value: 3, actor: 1}
	fmt.Println("Terminal moves:", terminal.PossibleMoves())
	fmt.Println("Terminal result:", terminal.Evaluate())
	// Output:
	// Moves from initial state: 1
	// Current actor: 1
	// Child actor: 2
	// Terminal moves: []
	// Terminal result: 1
}

// randomState implémente [decision.RandomMover] pour montrer l'optimisation.
type randomState struct {
	value int
}

func (s *randomState) CurrentActor() decision.ActorID  { return 1 }
func (s *randomState) PreviousActor() decision.ActorID { return 1 }
func (s *randomState) ID() string                      { return fmt.Sprintf("r%d", s.value) }
func (s *randomState) Evaluate() decision.ActorID {
	if s.value >= 5 {
		return decision.ActorID(1)
	}
	return decision.Undecided
}
func (s *randomState) PossibleMoves() []decision.State {
	if s.Evaluate() != decision.Undecided {
		return nil
	}
	return []decision.State{
		&randomState{value: s.value + 1},
		&randomState{value: s.value + 2},
	}
}

// RandomMove retourne un unique successeur choisi aléatoirement.
func (s *randomState) RandomMove(rng func(int) int) decision.State {
	delta := rng(2) + 1 // 1 ou 2
	return &randomState{value: s.value + delta}
}

// Cet exemple montre qu'un [decision.State] peut implémenter [decision.RandomMover]
// pour accélérer les rollouts MCTS.
func ExampleRandomMover() {
	s := &randomState{value: 0}

	// Vérifier que l'état implémente RandomMover via une variable d'interface.
	var state decision.State = s
	if rm, ok := state.(decision.RandomMover); ok {
		fmt.Println("State implements RandomMover")

		// RandomMove génère un seul successeur (vs PossibleMoves qui en génère tous).
		deterministic := func(n int) int { return 0 }
		child := rm.RandomMove(deterministic)
		fmt.Println("Child value:", child.(*randomState).value)
	}
	// Output:
	// State implements RandomMover
	// Child value: 1
}
