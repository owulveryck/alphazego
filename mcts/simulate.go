package mcts

import (
	"math/rand"

	"github.com/owulveryck/alphazego/decision"
)

// simulate performs a random playthrough from the current state until a terminal state is reached.
// It selects moves randomly and advances the state until it can be evaluated as a win, lose, or draw.
// Le générateur aléatoire utilisé provient de l'instance MCTS parente (si disponible),
// ce qui permet la reproductibilité via une graine fixée.
func (node *mctsNode) simulate() decision.ActorID {
	rng := rand.Intn // fallback sur le générateur global
	if node.mcts != nil && node.mcts.rng != nil {
		rng = node.mcts.rng.Intn
	}

	// Start from the current state of the node.
	currentState := node.state

	// Continue simulating random moves until the game reaches a terminal state.
	for currentState.Evaluate() == decision.NoActor {
		possibleMoves := currentState.PossibleMoves()
		currentState = possibleMoves[rng(len(possibleMoves))]
	}

	return currentState.Evaluate()
}
