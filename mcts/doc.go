// Package mcts implements the Monte Carlo Tree Search algorithm for
// sequential decision problems with one or more actors.
//
// MCTS works by repeatedly running four phases — selection, expansion,
// simulation, and backpropagation — to build a search tree and estimate
// the value of each possible move.
//
// # MCTS pur
//
// Create an [MCTS] instance with [NewMCTS], then call [MCTS.RunMCTS] with
// the current state and a number of iterations:
//
//	m := mcts.NewMCTS()
//	bestState := m.RunMCTS(currentState, 1000)
//
// The returned state represents the result after the best move found by
// the algorithm. If the state implements [board.ActionRecorder], use
// LastAction() to extract the actual move played.
//
// Each iteration performs:
//
//  1. Selection: descend the tree by picking the child with the highest
//     UCB1 score until a leaf node is reached.
//  2. Expansion: add one new child for an untried move.
//  3. Simulation: random rollout until a terminal state.
//  4. Backpropagation: propagate the result back up to the root.
//
// # AlphaZero Mode
//
// Create an [MCTS] instance with [NewAlphaMCTS], providing an
// [Evaluator] (neural network) and an exploration constant cpuct:
//
//	m := mcts.NewAlphaMCTS(evaluator, 1.5)
//	bestState := m.RunMCTS(currentState, 800)
//
// Each iteration performs:
//
//  1. Selection: descend the tree using PUCT (with priors from the
//     policy network) instead of UCB1.
//  2. Expansion + Evaluation: single call to [Evaluator.Evaluate] to
//     obtain policy and value. All children are created at once with
//     their priors.
//  3. No simulation: the network value replaces the rollout.
//  4. Backpropagation: propagates per-actor values (one value per
//     [decision.ActorID]), without zero-sum assumption.
//
// After all iterations, the child of the root with the most visits is
// selected as the best move.
//
// The algorithm is problem-agnostic: it works with any type implementing
// [decision.State].
package mcts
