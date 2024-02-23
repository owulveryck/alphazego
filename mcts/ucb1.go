package mcts

import "math"

// UCB1 calculates the Upper Confidence Bound (UCB) score for the given node
// Formula: Q(s, a) + √(2 * log(N(p)) / N(s, a))
// Where N(p) is the number of visits to the parent node, and N(s, a) is the number of visits to this node
//
// This UCB1 function is designed to balance exploration and exploitation by
// evaluating both the average success (reward) of a node and its relative unexplored state.
// Higher UCB scores indicate nodes that are either highly successful or insufficiently explored.
func (n *MCTSNode) UCB1() float64 {
	if n.visits == 0 {
		// Returns infinity for unvisited nodes to prioritize their exploration.
		// This ensures that every action is tried at least once before further exploitation.
		return math.Inf(1)
	}

	// Exploration constant, typically sqrt(2), to balance exploration and exploitation.
	// Adjusting this constant can make the algorithm prefer exploring new nodes
	// or exploiting nodes with high average rewards.
	C := math.Sqrt(2)

	// avgReward represents the Q value, which is the average reward of taking a specific
	// action from this state, as observed from past simulations.
	// It is calculated as the total reward accumulated from this node divided by the
	// number of visits to this node.
	// Q(s, a) = 1/N(s, a) * Σ(Rᵢ) for i from 1 to N(s, a), where Rᵢ represents the reward
	// received after the i-th visit to the node, and N(s, a) is the total number of visits.
	avgReward := n.wins / float64(n.visits)

	// UCB value calculation using the UCB1 formula:
	// UCB1 = Q(s, a) + C * sqrt(log(N(p)) / N(s, a)),
	// where N(p) is the number of visits to the parent node and N(s, a) is the number of visits
	// to this node. This formula helps in balancing the exploration of less visited nodes
	// and exploitation of nodes with high average rewards.
	ucbValue := avgReward + C*math.Sqrt(math.Log(float64(n.parent.visits))/float64(n.visits))

	return ucbValue
}
