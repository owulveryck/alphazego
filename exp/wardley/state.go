package wardley

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"math"

	"github.com/owulveryck/alphazego/decision"
)

// Player est l'unique acteur de l'exploration stratégique.
const Player decision.ActorID = 1

// Candidate représente une modification stratégique candidate proposée par le LLM.
type Candidate struct {
	Description string  // description lisible du move
	WTG2        string  // texte WTG2 complet résultant
	Confidence  float64 // [0,1]
}

// Proposer génère des modifications stratégiques candidates pour une carte Wardley.
// Les implémentations appellent un LLM et retournent des candidats structurés.
type Proposer interface {
	Propose(ctx context.Context, prompt string, n int) ([]Candidate, error)
}

// Annotation représente une note explicative attachée à un composant.
type Annotation struct {
	Kind   string // "note" ou "warning"
	Text   string
	Target string // nom du composant
}

// State représente l'état d'une carte Wardley pour l'exploration MCTS.
// Il implémente [decision.State] comme un puzzle mono-acteur.
//
// L'état stocke le texte WTG2 brut comme représentation canonique,
// préservant les positions exactes, groupes, pipelines, notes et signaux.
// Les candidats (enfants) sont générés par le [Proposer] lors du premier
// appel à [State.PossibleMoves], puis cachés pour le déterminisme.
type State struct {
	wtg2Text string
	title    string
	question string
	history  []string // descriptions des moves appliqués
	maxDepth int

	proposer     Proposer
	ctx          context.Context
	branchFactor int

	cachedMoves       []decision.State
	cachedConfidences []float64
	annotations       []Annotation
	lastErr           error
}

// Option configure un [State] lors de sa construction.
type Option func(*State)

// WithBranchFactor définit le nombre de candidats générés par appel au [Proposer].
// Par défaut : 5.
func WithBranchFactor(n int) Option {
	return func(s *State) {
		s.branchFactor = n
	}
}

// NewState crée un état initial à partir du texte WTG2 brut.
func NewState(wtg2Text, title, question string, maxDepth int, proposer Proposer, ctx context.Context, opts ...Option) *State {
	s := &State{
		wtg2Text:     wtg2Text,
		title:        title,
		question:     question,
		maxDepth:     maxDepth,
		proposer:     proposer,
		ctx:          ctx,
		branchFactor: 5,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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

// PossibleMoves retourne les états atteignables par une modification stratégique.
// Les candidats sont générés par le [Proposer] au premier appel, puis cachés.
// Retourne nil si l'état est terminal ou si le Proposer échoue.
func (s *State) PossibleMoves() []decision.State {
	if s.Evaluate() != decision.Undecided {
		return nil
	}
	if s.cachedMoves != nil {
		return s.cachedMoves
	}

	prompt := formatProposalPrompt(s.wtg2Text, s.question, s.history, s.branchFactor)
	candidates, err := s.proposer.Propose(s.ctx, prompt, s.branchFactor)
	if err != nil {
		s.lastErr = fmt.Errorf("wardley: erreur du Proposer: %w", err)
		log.Printf("%v", s.lastErr)
		return nil
	}
	if len(candidates) == 0 {
		s.lastErr = fmt.Errorf("wardley: le Proposer n'a retourné aucun candidat")
		log.Printf("%v", s.lastErr)
		return nil
	}

	var moves []decision.State
	var confidences []float64
	for _, c := range candidates {
		if err := validateWTG2(c.WTG2); err != nil {
			log.Printf("wardley: candidat rejeté (WTG2 invalide): %v", err)
			continue
		}
		childHistory := make([]string, len(s.history), len(s.history)+1)
		copy(childHistory, s.history)
		childHistory = append(childHistory, c.Description)

		child := &State{
			wtg2Text:     c.WTG2,
			title:        s.title,
			question:     s.question,
			history:      childHistory,
			maxDepth:     s.maxDepth,
			proposer:     s.proposer,
			ctx:          s.ctx,
			branchFactor: s.branchFactor,
		}
		moves = append(moves, child)
		confidences = append(confidences, math.Max(c.Confidence, 1e-8))
	}

	if len(moves) == 0 {
		s.lastErr = fmt.Errorf("wardley: tous les candidats ont un WTG2 invalide")
		log.Printf("%v", s.lastErr)
		return nil
	}

	s.cachedMoves = moves
	s.cachedConfidences = confidences
	return s.cachedMoves
}

// CachedConfidences retourne les scores de confiance des candidats cachés,
// dans le même ordre que [State.PossibleMoves]. Utilisé par l'Evaluator
// pour construire la policy MCTS.
func (s *State) CachedConfidences() []float64 {
	out := make([]float64, len(s.cachedConfidences))
	copy(out, s.cachedConfidences)
	return out
}

// ID retourne un identifiant unique basé sur le texte WTG2.
func (s *State) ID() string {
	h := sha256.New()
	h.Write([]byte(s.wtg2Text))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Title retourne le titre de la carte.
func (s *State) Title() string { return s.title }

// Question retourne la question stratégique.
func (s *State) Question() string { return s.question }

// WTG2Text retourne le texte WTG2 brut de la carte.
func (s *State) WTG2Text() string { return s.wtg2Text }

// History retourne la séquence de descriptions des moves appliqués.
func (s *State) History() []string {
	out := make([]string, len(s.history))
	copy(out, s.history)
	return out
}

// LastDescription retourne la description du dernier move appliqué.
func (s *State) LastDescription() string {
	if len(s.history) == 0 {
		return ""
	}
	return s.history[len(s.history)-1]
}

// Annotations retourne les annotations de la carte.
func (s *State) Annotations() []Annotation {
	out := make([]Annotation, len(s.annotations))
	copy(out, s.annotations)
	return out
}

// SetAnnotations remplace les annotations de la carte.
func (s *State) SetAnnotations(annotations []Annotation) {
	s.annotations = make([]Annotation, len(annotations))
	copy(s.annotations, annotations)
}

// LastError retourne la dernière erreur survenue lors de l'appel à
// [State.PossibleMoves]. Retourne nil si aucune erreur ne s'est produite.
func (s *State) LastError() error {
	return s.lastErr
}
