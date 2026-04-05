package reasoning

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/owulveryck/alphazego/decision"
)

// Player est l'unique acteur du raisonnement. Dans un problème à un seul
// acteur, CurrentActor() et PreviousActor() retournent toujours cette valeur.
const Player decision.ActorID = 1

// ConclusionPrefix est le préfixe qu'une étape de raisonnement doit avoir
// pour être considérée comme une conclusion. Quand la dernière étape commence
// par ce préfixe, [State.Evaluate] retourne [Player] (résolu).
const ConclusionPrefix = "CONCLUSION:"

// Generator génère des étapes de raisonnement candidates. Il reçoit un prompt
// formaté contenant la question, les étapes précédentes et le critère de
// succès, et retourne n étapes candidates.
//
// Les implémentations concrètes (VertexAI, Ollama, etc.) vivent dans des
// modules séparés avec leur propre go.mod.
type Generator interface {
	Generate(ctx context.Context, prompt string, n int) ([]string, error)
}

// Judge évalue la qualité d'un raisonnement. Il reçoit un prompt formaté
// contenant le raisonnement à évaluer et retourne un score dans [0, 1].
//
// Les implémentations concrètes (VertexAI, Ollama, etc.) vivent dans des
// modules séparés avec leur propre go.mod.
type Judge interface {
	Score(ctx context.Context, prompt string) (float64, error)
}

// Option configure un [State] lors de sa construction avec [New].
type Option func(*State)

// WithMaxDepth définit la profondeur maximale du raisonnement. Si le nombre
// d'étapes atteint cette limite sans conclusion, [State.Evaluate] retourne
// [decision.Stalemate]. Par défaut : 5.
func WithMaxDepth(n int) Option {
	return func(s *State) {
		s.maxDepth = n
	}
}

// WithBranchFactor définit le nombre d'étapes candidates générées par appel
// à [State.PossibleMoves]. Par défaut : 3.
func WithBranchFactor(n int) Option {
	return func(s *State) {
		s.branchFactor = n
	}
}

// State représente l'état d'un raisonnement par décomposition. Il implémente
// [decision.State] pour être utilisé avec le MCTS.
//
// L'état contient la question initiale, un critère de succès, et les étapes
// de raisonnement accumulées. [State.PossibleMoves] génère des étapes candidates
// via le [Generator] et les cache pour garantir le déterminisme.
//
// En cas d'erreur du [Generator], [State.PossibleMoves] retourne nil (traité
// comme un état terminal par le MCTS). L'erreur est accessible via [State.LastError].
type State struct {
	question     string
	criterion    string
	steps        []string
	generator    Generator
	ctx          context.Context
	maxDepth     int
	branchFactor int
	cachedMoves  []decision.State
	lastErr      error
}

// New crée un état de raisonnement initial avec la question et le critère de
// succès donnés. Le [Generator] est utilisé pour produire les étapes candidates.
//
// Options disponibles : [WithMaxDepth] (défaut 5), [WithBranchFactor] (défaut 3).
func New(ctx context.Context, question, criterion string, gen Generator, opts ...Option) *State {
	s := &State{
		question:     question,
		criterion:    criterion,
		generator:    gen,
		ctx:          ctx,
		maxDepth:     5,
		branchFactor: 3,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// CurrentActor retourne [Player]. Le raisonnement est un problème mono-acteur.
func (s *State) CurrentActor() decision.ActorID {
	return Player
}

// PreviousActor retourne [Player]. Le raisonnement est un problème mono-acteur.
func (s *State) PreviousActor() decision.ActorID {
	return Player
}

// Evaluate retourne l'issue du raisonnement :
//   - [Player] si la dernière étape commence par [ConclusionPrefix]
//   - [decision.Stalemate] si la profondeur maximale est atteinte
//   - [decision.Undecided] sinon
//
// Cette méthode n'effectue aucun appel LLM.
func (s *State) Evaluate() decision.ActorID {
	if len(s.steps) > 0 && strings.HasPrefix(s.steps[len(s.steps)-1], ConclusionPrefix) {
		return Player
	}
	if len(s.steps) >= s.maxDepth {
		return decision.Stalemate
	}
	return decision.Undecided
}

// PossibleMoves retourne les états atteignables en ajoutant une étape de
// raisonnement. Les étapes candidates sont générées par le [Generator] au
// premier appel, puis cachées pour garantir le déterminisme.
//
// Retourne nil si l'état est terminal.
func (s *State) PossibleMoves() []decision.State {
	if s.Evaluate() != decision.Undecided {
		return nil
	}
	if s.cachedMoves != nil {
		return s.cachedMoves
	}

	prompt := formatGeneratePrompt(s.question, s.criterion, s.steps)
	candidates, err := s.generator.Generate(s.ctx, prompt, s.branchFactor)
	if err != nil {
		s.lastErr = fmt.Errorf("reasoning: erreur du Generator: %w", err)
		log.Printf("%v", s.lastErr)
		return nil
	}
	if len(candidates) == 0 {
		s.lastErr = fmt.Errorf("reasoning: le Generator n'a retourné aucun candidat (question=%q, étapes=%d)", s.question, len(s.steps))
		log.Printf("%v", s.lastErr)
		return nil
	}

	moves := make([]decision.State, len(candidates))
	for i, candidate := range candidates {
		childSteps := make([]string, len(s.steps), len(s.steps)+1)
		copy(childSteps, s.steps)
		childSteps = append(childSteps, candidate)

		moves[i] = &State{
			question:     s.question,
			criterion:    s.criterion,
			steps:        childSteps,
			generator:    s.generator,
			ctx:          s.ctx,
			maxDepth:     s.maxDepth,
			branchFactor: s.branchFactor,
		}
	}

	s.cachedMoves = moves
	return s.cachedMoves
}

// ID retourne un identifiant unique pour cet état, basé sur la question
// et toutes les étapes de raisonnement.
func (s *State) ID() string {
	h := sha256.New()
	h.Write([]byte(s.question))
	for _, step := range s.steps {
		h.Write([]byte("\n"))
		h.Write([]byte(step))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Steps retourne les étapes de raisonnement accumulées.
func (s *State) Steps() []string {
	return s.steps
}

// Question retourne la question initiale.
func (s *State) Question() string {
	return s.question
}

// Criterion retourne le critère de succès.
func (s *State) Criterion() string {
	return s.criterion
}

// LastError retourne la dernière erreur survenue lors de l'appel à
// [State.PossibleMoves]. Retourne nil si aucune erreur ne s'est produite.
// Cette méthode permet de diagnostiquer pourquoi PossibleMoves a retourné nil
// (erreur du [Generator] vs état terminal légitime).
func (s *State) LastError() error {
	return s.lastErr
}

// Errors retourne une erreur joignant toutes les erreurs de cet état et de
// ses ancêtres (via [errors.Join]). Utile pour diagnostiquer un raisonnement
// qui s'est terminé prématurément.
func (s *State) Errors() error {
	return errors.Join(s.lastErr)
}
