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
// # Mode AlphaZero
//
// Create an [MCTS] instance with [NewAlphaMCTS] en fournissant un
// [Evaluator] (réseau de neurones) et une constante d'exploration cpuct :
//
//	m := mcts.NewAlphaMCTS(evaluator, 1.5)
//	bestState := m.RunMCTS(currentState, 800)
//
// Each iteration performs:
//
//  1. Selection: descend the tree using PUCT (avec priors du
//     policy network) au lieu de UCB1.
//  2. Expansion + Évaluation : appel unique à [Evaluator.Evaluate] pour
//     obtenir policy et value. Tous les enfants sont créés d'un coup avec
//     leurs priors.
//  3. Pas de simulation : la value du réseau remplace le rollout.
//  4. Backpropagation : propage la value continue avec inversion de signe.
//
// After all iterations, the child of the root with the most visits is
// selected as the best move.
//
// The algorithm is problem-agnostic: it works with any type implementing
// [decision.State].
package mcts
