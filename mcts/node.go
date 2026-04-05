package mcts

import (
	"math"

	"github.com/owulveryck/alphazego/decision"
)

// mctsNode represents a single node in the Monte Carlo Tree Search (MCTS) algorithm.
// Each node corresponds to a specific game state and contains statistical information
// about the outcomes of simulations that have been run through this node. The structure
// of the tree is formed by parent and child relationships between nodes, enabling the
// navigation and expansion of the search tree as the algorithm progresses.
//
// Les champs sont ordonnés par fréquence d'accès pour optimiser la localité cache.
// Les champs « chauds » (backpropagate + sélection) sont regroupés en tête de struct
// pour tenir dans la première ligne de cache (64 octets sur arm64/x86-64).
type mctsNode struct {
	// --- Hot fields (cache line 1, 64B) : backpropagate + selection ---

	// wins records the total number of wins observed in simulations
	// that have passed through this node.
	wins float64

	// visits records the total number of times this node has been visited.
	visits float64

	// logVisits est le logarithme naturel de visits, mis à jour dans
	// backpropagate() pour éviter de recalculer math.Log() dans la boucle
	// de sélection. ucb1() et puct() lisent parent.logVisits directement.
	logVisits float64

	// sqrtVisits est la racine carrée de visits, mise à jour dans
	// backpropagate() pour éviter de recalculer math.Sqrt() pour chaque
	// enfant dans la formule PUCT. puct() lit parent.sqrtVisits directement.
	sqrtVisits float64

	// prior est la probabilité a priori P(s,a) attribuée par le policy network.
	// En MCTS pur, cette valeur est 0 (non utilisée). En mode AlphaZero,
	// elle est fixée lors de l'expansion par l'Evaluator et utilisée dans la
	// formule PUCT pour guider la sélection.
	prior float64

	// parent is a pointer to the parent node in the search tree.
	parent *mctsNode

	// mcts holds a reference back to the MCTS instance.
	mcts *MCTS

	// previousActor est l'acteur qui a effectué l'action menant à cet état,
	// mis en cache à la création du nœud pour éviter l'appel d'interface
	// PreviousActor() dans la boucle de backpropagation.
	previousActor decision.ActorID

	// --- Warm fields (cache line 2) : expand + selection ---

	// state holds the current state that this node represents.
	state decision.State

	// children is a slice of pointers to the child nodes of this node.
	children []*mctsNode

	// expandedIndex est le nombre d'enfants déjà créés par expand().
	// Il sert de curseur dans le slice cachedMoves : le prochain coup
	// non exploré est cachedMoves[expandedIndex].
	expandedIndex int

	// cachedMovesCount est le nombre de coups possibles, mis en cache par
	// getPossibleMoves() pour permettre l'inlining de isFullyExpanded().
	cachedMovesCount int

	// --- Cold fields : accédés rarement ---

	// cachedMoves stocke le résultat de PossibleMoves(), mis en cache pour
	// éviter les allocations répétées dans isFullyExpanded() et expand().
	cachedMoves         []decision.State
	cachedMovesComputed bool

	// cachedEval stocke le résultat de state.Evaluate(), mis en cache par
	// isTerminal() pour éviter un double appel dans la boucle RunMCTS.
	cachedEval         decision.ActorID
	cachedEvalComputed bool
}

// isTerminal returns true if this node represents a terminal state (win, loss, or draw).
// Le résultat d'Evaluate est caché pour éviter les appels redondants.
func (n *mctsNode) isTerminal() bool {
	if !n.cachedEvalComputed {
		n.cachedEval = n.state.Evaluate()
		n.cachedEvalComputed = true
	}
	return n.cachedEval != decision.Undecided
}

// getPossibleMoves retourne les coups possibles, en les cachant au premier appel.
// Met également à jour cachedMovesCount pour permettre l'inlining de isFullyExpanded.
func (n *mctsNode) getPossibleMoves() []decision.State {
	if !n.cachedMovesComputed {
		n.cachedMoves = n.state.PossibleMoves()
		n.cachedMovesCount = len(n.cachedMoves)
		n.cachedMovesComputed = true
	}
	return n.cachedMoves
}

// isFullyExpanded returns true if all possible moves from this state have been expanded as children.
// Cette fonction utilise cachedMovesCount (mis à jour par getPossibleMoves) au lieu
// d'appeler getPossibleMoves+len, ce qui réduit le coût d'inlining sous le budget
// du compilateur (82 → ~15). Si cachedMoves n'a pas encore été calculé,
// retourne false pour forcer l'appel à expand/expandAll qui initialisera le cache.
func (n *mctsNode) isFullyExpanded() bool {
	return n.cachedMovesComputed && n.expandedIndex >= n.cachedMovesCount
}

// selectChildUCB selects the immediate child with the highest score.
// En MCTS pur, la formule UCB1 est utilisée. En mode AlphaZero (evaluator != nil),
// la formule PUCT est utilisée. Les deux formules sont appelées directement
// (pas via pointeur de fonction) pour permettre leur inlining par le compilateur.
func (n *mctsNode) selectChildUCB() *mctsNode {
	if len(n.children) == 0 {
		return nil
	}
	bestScore := math.Inf(-1)
	var bestChild *mctsNode
	useAlphaZero := n.mcts != nil && n.mcts.evaluator != nil
	for _, child := range n.children {
		var score float64
		if useAlphaZero {
			score = child.puct()
		} else {
			score = child.ucb1()
		}
		if score > bestScore {
			bestScore = score
			bestChild = child
		}
	}
	return bestChild
}

// selectBestMove returns the child with the highest visit count (most explored path).
func (n *mctsNode) selectBestMove() *mctsNode {
	var bestChild *mctsNode
	bestVisits := float64(-1)
	for _, child := range n.children {
		if child.visits > bestVisits {
			bestVisits = child.visits
			bestChild = child
		}
	}
	return bestChild
}
