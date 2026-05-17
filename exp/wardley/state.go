package wardley

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/owulveryck/alphazego/decision"
)

// Player est l'unique acteur de l'exploration stratégique.
const Player decision.ActorID = 1

// Phase représente la phase d'évolution d'un composant sur l'axe horizontal.
type Phase int

const (
	// Genesis — composant nouveau, mal compris, forte incertitude.
	Genesis Phase = iota
	// Custom — compris mais bespoke, nécessite de l'expertise.
	Custom
	// Product — de plus en plus standardisé, disponible comme produit.
	Product
	// Commodity — hautement standardisé, pay-per-use, ubiquitaire.
	Commodity
)

// String retourne le nom de la phase.
func (p Phase) String() string {
	switch p {
	case Genesis:
		return "Genesis"
	case Custom:
		return "Custom"
	case Product:
		return "Product"
	case Commodity:
		return "Commodity"
	default:
		return fmt.Sprintf("Phase(%d)", int(p))
	}
}

// MoveType distingue les deux types de moves stratégiques.
type MoveType int

const (
	// Evolve avance un composant d'une phase d'évolution.
	Evolve MoveType = iota
	// ApplyGameplay applique un gameplay stratégique à un composant.
	ApplyGameplay
)

// AvailableGameplays liste les gameplays applicables par le MCTS.
var AvailableGameplays = []string{
	"open-source",
	"ILC",
	"land-grab",
	"embrace-extend",
	"tower-moat",
	"strangler-fig",
}

// Component représente un composant de carte Wardley dans le modèle simplifié.
type Component struct {
	Name       string
	Phase      Phase
	Visibility int
	Type       string
	Inertia    int
	Gameplays  []string
}

// Edge représente une dépendance entre deux composants.
type Edge struct {
	From  string
	To    string
	Label string
}

// Move représente une action stratégique sur la carte.
type Move struct {
	Type      MoveType
	Component string
	Gameplay  string
}

// String retourne une description lisible du move.
func (m Move) String() string {
	switch m.Type {
	case Evolve:
		return fmt.Sprintf("EVOLVE %q", m.Component)
	case ApplyGameplay:
		return fmt.Sprintf("GAMEPLAY %q sur %q", m.Gameplay, m.Component)
	default:
		return fmt.Sprintf("Move(%d)", int(m.Type))
	}
}

// State représente l'état d'une carte Wardley pour l'exploration MCTS.
// Il implémente [decision.State] comme un puzzle mono-acteur.
type State struct {
	title      string
	question   string
	components []Component
	edges      []Edge
	history    []Move
	maxDepth   int
}

// NewState crée un état initial à partir des composants et edges extraits
// d'une carte Wardley. maxDepth limite la profondeur de l'arbre MCTS.
func NewState(title, question string, components []Component, edges []Edge, maxDepth int) *State {
	return &State{
		title:      title,
		question:   question,
		components: components,
		edges:      edges,
		maxDepth:   maxDepth,
	}
}

// CurrentActor retourne [Player]. L'exploration stratégique est mono-acteur.
func (s *State) CurrentActor() decision.ActorID { return Player }

// PreviousActor retourne [Player]. L'exploration stratégique est mono-acteur.
func (s *State) PreviousActor() decision.ActorID { return Player }

// Evaluate retourne l'issue de l'exploration :
//   - [decision.Stalemate] si la profondeur maximale est atteinte
//   - [decision.Undecided] sinon
func (s *State) Evaluate() decision.ActorID {
	if len(s.history) >= s.maxDepth {
		return decision.Stalemate
	}
	return decision.Undecided
}

// PossibleMoves retourne tous les états atteignables par un move stratégique.
// Deux types de moves : évoluer un composant ou appliquer un gameplay.
func (s *State) PossibleMoves() []decision.State {
	if s.Evaluate() != decision.Undecided {
		return nil
	}

	var moves []Move

	for _, c := range s.components {
		if c.Phase < Commodity && c.Inertia == 0 {
			moves = append(moves, Move{Type: Evolve, Component: c.Name})
		}
	}

	for _, c := range s.components {
		for _, gp := range AvailableGameplays {
			if !hasGameplay(c.Gameplays, gp) {
				moves = append(moves, Move{Type: ApplyGameplay, Component: c.Name, Gameplay: gp})
			}
		}
	}

	if len(moves) == 0 {
		return nil
	}

	states := make([]decision.State, len(moves))
	for i, m := range moves {
		states[i] = s.applyMove(m)
	}
	return states
}

// ID retourne un identifiant unique basé sur les composants et leurs gameplays.
func (s *State) ID() string {
	h := sha256.New()
	h.Write([]byte(s.question))

	names := make([]string, len(s.components))
	for i, c := range s.components {
		names[i] = c.Name
	}
	sort.Strings(names)

	for _, name := range names {
		for _, c := range s.components {
			if c.Name == name {
				gps := make([]string, len(c.Gameplays))
				copy(gps, c.Gameplays)
				sort.Strings(gps)
				fmt.Fprintf(h, "\n%s:%d:%s", c.Name, c.Phase, strings.Join(gps, ","))
				break
			}
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Title retourne le titre de la carte.
func (s *State) Title() string { return s.title }

// Question retourne la question stratégique.
func (s *State) Question() string { return s.question }

// Components retourne une copie des composants.
func (s *State) Components() []Component {
	out := make([]Component, len(s.components))
	for i, c := range s.components {
		out[i] = c
		out[i].Gameplays = make([]string, len(c.Gameplays))
		copy(out[i].Gameplays, c.Gameplays)
	}
	return out
}

// Edges retourne une copie des edges.
func (s *State) Edges() []Edge {
	out := make([]Edge, len(s.edges))
	copy(out, s.edges)
	return out
}

// History retourne la séquence de moves appliqués.
func (s *State) History() []Move {
	out := make([]Move, len(s.history))
	copy(out, s.history)
	return out
}

// LastMove retourne le dernier move appliqué, ou un Move vide si aucun.
func (s *State) LastMove() Move {
	if len(s.history) == 0 {
		return Move{}
	}
	return s.history[len(s.history)-1]
}

func (s *State) applyMove(m Move) *State {
	comps := make([]Component, len(s.components))
	for i, c := range s.components {
		comps[i] = c
		comps[i].Gameplays = make([]string, len(c.Gameplays))
		copy(comps[i].Gameplays, c.Gameplays)
	}

	for i, c := range comps {
		if c.Name != m.Component {
			continue
		}
		switch m.Type {
		case Evolve:
			comps[i].Phase++
		case ApplyGameplay:
			comps[i].Gameplays = append(comps[i].Gameplays, m.Gameplay)
		}
		break
	}

	hist := make([]Move, len(s.history), len(s.history)+1)
	copy(hist, s.history)
	hist = append(hist, m)

	edges := make([]Edge, len(s.edges))
	copy(edges, s.edges)

	return &State{
		title:      s.title,
		question:   s.question,
		components: comps,
		edges:      edges,
		history:    hist,
		maxDepth:   s.maxDepth,
	}
}

func hasGameplay(gameplays []string, gp string) bool {
	return slices.Contains(gameplays, gp)
}
