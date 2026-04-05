package mcts

import (
	"math/rand"

	"github.com/owulveryck/alphazego/decision"
)

// simulate performs a random playthrough from the current state until a terminal state is reached.
// It selects moves randomly and advances the state until it can be evaluated as a win, lose, or draw.
// The random generator comes from the parent MCTS instance (if available),
// enabling reproducibility via a fixed seed.
//
// If the state implements [decision.RandomMover], RandomMove is used instead
// of PossibleMoves to reduce allocations (N → 1 per rollout step).
func (node *mctsNode) simulate() decision.ActorID {
	// Fallback sur le générateur global pour les tests unitaires qui
	// construisent des nœuds sans instance MCTS. En production, node.mcts
	// est toujours non-nil (créé par newNode).
	rng := rand.Intn
	if node.mcts != nil && node.mcts.rng != nil {
		rng = node.mcts.rng.Intn
	}

	currentState := node.state

	// Vérifier une seule fois si l'état implémente RandomMover.
	// En pratique, tous les états d'un même jeu ont le même type concret :
	// si le premier l'implémente, les suivants aussi.
	if _, ok := currentState.(decision.RandomMover); ok {
		for {
			result := currentState.Evaluate()
			if result != decision.Undecided {
				return result
			}
			currentState = currentState.(decision.RandomMover).RandomMove(rng)
		}
	}

	for {
		result := currentState.Evaluate()
		if result != decision.Undecided {
			return result
		}
		possibleMoves := currentState.PossibleMoves()
		if len(possibleMoves) == 0 {
			// Malformed State: Evaluate() says Undecided but no moves available.
			// Treat as stalemate to avoid panic.
			return decision.Stalemate
		}
		currentState = possibleMoves[rng(len(possibleMoves))]
	}
}
