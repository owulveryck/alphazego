// Package mcts implements the Monte Carlo Tree Search algorithm for
// two-player board games.
//
// MCTS works by repeatedly running four phases — selection, expansion,
// simulation, and backpropagation — to build a search tree and estimate
// the value of each possible move.
//
// # Usage
//
// Create an [MCTS] instance, then call [MCTS.RunMCTS] with the current
// game state and a number of iterations:
//
//	m := mcts.NewMCTS()
//	bestState := m.RunMCTS(currentState, 1000)
//
// The returned state represents the board after the best move found by
// the algorithm. Use [board.Playable.GetMoveFromState] to extract the
// actual move played.
//
// # Algorithm overview
//
// Each iteration of MCTS performs the following steps:
//
//  1. Selection: starting from the root, descend the tree by picking the
//     child with the highest UCB1 score until a leaf node is reached.
//  2. Expansion: if the leaf is not terminal, add one new child for an
//     untried move.
//  3. Simulation: play a random game (rollout) from the new node to a
//     terminal state.
//  4. Backpropagation: propagate the result back up to the root, updating
//     visit counts and win statistics.
//
// After all iterations, the child of the root with the most visits is
// selected as the best move.
//
// The algorithm is game-agnostic: it works with any type implementing
// [board.State].
package mcts
