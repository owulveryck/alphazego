# alphazego

A repository that may not go far but could be eventually an alphazero implementation in Go from Scratch

## Introduction

I would like to understand AlphaGo and AlphaZero by implementing it from scratch.
I don't know how far it will go.

I will use tic-tac-toe as it is a fairly easy game to understand.

First, I will implement a MTCS from scratch.
Then I will try to add the deep-learning parts

### MCTS

This is written by ChatGPT:

Monte Carlo Tree Search (MCTS) is a decision-making algorithm commonly used in artificial intelligence (AI) applications, especially for playing board games like Go, chess, and many others.
It helps an AI decide on the best move to make in a given situation.
Here's a simplified explanation of how MCTS works:

#### Tree Structure
Imagine the game as a tree where each node represents a game state (or position), and the branches represent the possible moves (or actions) leading to the next states.

#### Four Phases of MCTS

- Selection: Starting from the root node (the current state of the game), the algorithm selects child nodes down the tree based on a strategy that balances between exploring new nodes and exploiting known successful paths.
This strategy often involves some mathematical formulae like the Upper Confidence Bound (UCB1).
- Expansion: Once it reaches a node that has unexplored child nodes (possible moves that haven't been tried yet), the algorithm expands the tree by adding one of these new nodes.
- Simulation: From this new node, the algorithm simulates a random play-out or game until a predefined condition is met (like reaching the end of the game).
This random simulation helps estimate the potential of the move that led to the new node.
- Backpropagation: After the simulation ends, the results (win or lose) are propagated back up the tree, updating the information at each node along the way.
This update improves the accuracy of the selection strategy for future iterations.

#### Repetition
MCTS repeats these four phases many times, building a more and more refined tree where the paths to high-scoring outcomes are explored more thoroughly.

#### Decision
After a set number of iterations or time, the algorithm picks the move associated with the most promising path explored in the tree.

Key Takeaway
The beauty of MCTS is its balance between exploring new, potentially better moves (exploration) and exploiting the moves known to lead to success (exploitation).
It doesn't require an exhaustive search of all possible game states or an extensive database of past games, making it powerful for complex games like Go.
The algorithm "learns" the best moves by simulating many games within the possible moves tree and uses those simulations to make informed decisions, improving as it goes.

### UCB1

Imagine you're at a carnival with lots of different games to play, but you only have a limited amount of time and tokens.
Some games you've tried before and know they're fun, and some games are totally new to you.
You want to make sure you have the best time by playing the most fun games, but you also want to try new games because they might be even more fun.
This is where the UCB1 formula, which stands for "Upper Confidence Bound 1," comes into play, but let's call it the "Fun Finder Formula" to make it easier.

The Fun Finder Formula helps you decide which game to play next by thinking about two things:

#### How much fun you had before
If you played a game before and it was super fun, you'll probably want to play it again.
The Fun Finder Formula remembers how much fun each game was based on your past plays.

#### Giving all games a chance
Even if you haven't tried a game yet, the Fun Finder Formula nudges you to try it because who knows? It might turn out to be the most fun game at the carnival.

Here's how it does it:

- It gives each game a score that goes up both because of how much fun it was before and because of how few times you've tried it.
This means games that were a lot of fun get high scores, but so do games you haven't played much yet.
- Every time you're deciding what to play, you look at all the games' scores and pick the one with the highest score.
This means you're choosing between playing a game you know is fun and trying out a game that could be even more fun.

So, the Fun Finder Formula (UCB1) helps you have the best time at the carnival by making sure you play a mix of games you know you love and new games that might become your new favorites.

#### Maths

Let's add some math to our "Fun Finder Formula" (UCB1) to see how it actually works when you're deciding which carnival game to play next.
Each game at the carnival has two important numbers:

- Average Fun Score: This is how much fun you've had on average each time you played the game.
Let's call this $AverageFun(game)$
- Number of Plays: This is how many times you've played the game.
We'll call this $Plays(game)$.

And there's one more number that's about the whole carnival:

- **Total Number of Plays**: This is how many games you've played in total, not just one game but all the different games.
Let's call this $ \text{TotalPlays} $.

The Fun Finder Formula (UCB1) combines these numbers to give each game a \"Try This Next\" score.
Here's the formula for the score of each game:

$$ \text{Score}(game) = \text{AverageFun}(game) + \sqrt{\frac{2 \times \ln(\text{TotalPlays})}{\text{Plays}(game)}} $$

This formula has two parts:

1. **$\text{AverageFun}(game)$**: This is the first part, which says, \"If a game was fun before, it's probably still fun.\" It's like remembering which games made you laugh the most.
2. **$\sqrt{\frac{2 \times \ln(\text{TotalPlays})}{\text{Plays}(game)}}$**: This is the second part, which pushes you to try games you haven't played much yet.
The more you play in total (making $ \text{TotalPlays} $ bigger) and the fewer times you've tried a game (making $ \text{Plays}(game) $ smaller), the bigger this part gets.
This means games you haven't played much yet get a boost in their score, tempting you to try them.

When you combine these parts, every game at the carnival gets a score that balances \"this is already fun\" with \"this could be exciting because it's new.\" Before you decide what to play next, you calculate these scores for all the games and then go to the one with the highest score.
That way, you're always choosing between the best of what you know and the exciting possibilities of what you don't, making sure your time at the carnival is super fun!"

### Creating the basic MCTS in Go 

Creating a simple Monte Carlo Tree Search (MCTS) for Tic-Tac-Toe in Go (Golang) involves several steps.
Below is a high-level overview of what you need to do, followed by a basic code structure to get you started.

#### Steps to Create MCTS for Tic-Tac-Toe
1. Define the Game State: Represent the Tic-Tac-Toe board and keep track of the current player, the positions of 'X' and 'O', and whether the game is over.
2. Implement Game Logic: Functions for making moves, checking for a win or draw, and generating possible moves from the current state.
3. Node Structure for MCTS: Each node in the MCTS tree will represent a game state, including information like the number of wins, losses, and visits for that state, as well as links to parent and child nodes.
4. Selection: Navigate the tree from the root to a leaf node by choosing nodes that maximize a given selection strategy, such as the Upper Confidence Bound (UCB1).
5. Expansion: If the chosen leaf node represents a non-terminal state, expand the tree by creating new child nodes for each possible move.
6. Simulation: From a new child node, simulate random play (rollout) until the game reaches a terminal state, then evaluate the outcome.
7. Backpropagation: Update the nodes from the selected leaf back to the root with the result of the simulation, adjusting the win/loss statistics.

```go 
package main

import "fmt"

// Define constants for the players and empty cells
const (
    Empty = 0
    PlayerX = 1
    PlayerO = 2
)

// GameState represents the Tic-Tac-Toe board
type GameState struct {
    board [3][3]int
    playerTurn int
}

// MCTSNode represents a node in the MCTS tree
type MCTSNode struct {
    state GameState
    parent *MCTSNode
    children []*MCTSNode
    wins float64
    visits float64
    untriedMoves []GameState // Or any representation of a move
}
```

### Outline for MCTS Implementation
1. Game State Management: Implement functions to make moves, check for game over conditions (win, lose, draw), and list possible moves from the current state.
2. MCTS Node Structure: Design the MCTSNode struct with fields for tracking the game state, parent and child nodes, wins, visits, and possible moves.
3. Selection: Use the UCB1 formula to traverse the tree from the root to a leaf node, selecting the child with the highest value each time.
4. Expansion: At a leaf node, if the game isn't over, create new child nodes for each possible move not yet explored.
5. Simulation: From a new node, simulate random moves to the end of the game and determine the winner.
6. Backpropagation: Update nodes with the simulation result, incrementing visits and adjusting the win count based on the outcome.

```go 
package main

import (
    "fmt"
    "math"
    "math/rand"
    "time"
)

type GameState struct {
    board [3][3]int
    playerTurn int
}

type MCTSNode struct {
    state GameState
    parent *MCTSNode
    children []*MCTSNode
    wins float64
    visits float64
    untriedMoves []GameState
}

// Placeholder for GameState methods
func (gs *GameState) PossibleMoves() []GameState {
    // Return a slice of possible next states
    return nil
}

func (gs *GameState) IsGameOver() bool {
    // Return true if the game is over, false otherwise
    return false
}

func (gs *GameState) MakeMove(move GameState) *GameState {
    // Apply a move to the current game state and return the new state
    return &GameState{}
}

func (gs *GameState) GetWinner() int {
    // Determine the winner of the game; return 0 for draw, 1 for Player X, 2 for Player O
    return 0
}

// MCTSNode methods
func (node *MCTSNode) UCB1() float64 {
    if node.visits == 0 {
        return math.Inf(1) // Return positive infinity to prioritize unvisited nodes
    }
    return node.wins / node.visits + math.Sqrt(2*math.Log(node.parent.visits)/node.visits)
}

func (node *MCTSNode) SelectChild() *MCTSNode {
    // Select the child with the highest UCB1 score
    return nil
}

func (node *MCTSNode) Expand() *MCTSNode {
    // Expand the tree by creating a new child node for one of the untried moves
    return nil
}

func (node *MCTSNode) Simulate() int {
    // Simulate a random playthrough from this node to a terminal state
    return 0
}

func (node *MCTSNode) Backpropagate(result int) {
    // Update this node and its ancestors with the simulation result
}

```

## MCTS Methods

Flesh out the selection, expansion, simulation, and backpropagation methods for the MCTSNode.
This involves implementing the UCB1 calculation, expanding the tree with new moves, simulating games to determine outcomes, and updating nodes based on simulation results.

### Backpropagation

Backpropagation in the context of Monte Carlo Tree Search (MCTS), specifically for a game like Tic-Tac-Toe, is the process of updating the nodes of the search tree after a simulation is completed. The purpose of backpropagation is to incrementally adjust the statistical information stored at each node based on the outcome of the simulation, which helps the algorithm make more informed decisions in future iterations. Here's a detailed explanation:

Goal of Backpropagation
The goal is to update each node from the selected node down to the root with the result of the simulation to reflect how promising each node (or move) is based on the new information. This involves updating two main pieces of information:

Visits (N): The number of times a node has been visited, which includes the current visit.
Wins (W): The total wins recorded from the node's perspective. The definition of a "win" can vary based on the game and the perspective of the player making the simulation.
How Backpropagation Works
Result Evaluation: Once the simulation reaches a terminal state (win, lose, or draw), the outcome is evaluated. In Tic-Tac-Toe, this could be a win for X, a win for O, or a draw.

Backpropagation Loop: Starting from the node that was expanded and simulated, the algorithm moves back up the tree to the root. At each node, it updates the statistics based on the simulation's outcome:

Increment Visits: For every node along the path back to the root, increment the visits count by 1. This reflects that a new simulation has passed through this node.
Update Wins: If the simulation resulted in a win from the perspective of the player corresponding to the node, increment the wins count. For Tic-Tac-Toe, if the simulation ends in a win for X, all nodes representing X's turns along the path would have their wins count incremented.
Perspective Handling: It's important to adjust the wins based on the perspective of the node. In games like Tic-Tac-Toe, where players take turns, the interpretation of a win is inverted at each level of the tree. If the simulated game is won by X, then nodes where it was X's turn would count this as a win, while nodes where it was O's turn would not.

```go
func (node *MCTSNode) Backpropagate(result int) {
    // Loop to update nodes up to the root
    for n := node; n != nil; n = n.parent {
        n.visits += 1
        // If the result matches the playerTurn of this node, it's a win for this node
        if n.state.playerTurn == result {
            n.wins += 1
        }
        // For Tic-Tac-Toe, you might also need to handle draws specifically
        // depending on how you want to treat them in your win/loss statistics
    }
}
```

**Key Considerations**
Draw Handling: How to handle draws depends on your specific implementation. Some strategies might consider a draw as a half-win or a separate statistic.
Perspective Adjustment: The example treats wins from the perspective of the current player at each node. Ensure this aligns with how you've structured your game state and tree.

### Selection
Selection in Monte Carlo Tree Search (MCTS) is a critical step where the algorithm decides which path to take through the game tree, aiming to balance between exploring new possibilities and exploiting known successful paths.
This balance is crucial for efficiently searching the vast space of possible moves in complex games like Tic-Tac-Toe.
Here's a detailed explanation:

### Goal of Selection

The primary goal of the selection phase is to navigate from the root of the tree to a leaf node by making a series of choices that lead to promising areas of the tree.
A \"promising area\" could mean a path that has historically led to wins (exploitation) or a path that hasn't been explored much yet and might lead to new discoveries (exploration).

### The Upper Confidence Bound (UCB1) Algorithm

One popular strategy for selection in MCTS is the Upper Confidence Bound (UCB1) algorithm.
UCB1 balances exploration and exploitation by considering both the win rate of a node and how frequently the node has been visited relative to its siblings.
The formula for UCB1 is:

$$ \text{UCB1} = \frac{W_i}{N_i} + C \sqrt{\frac{\ln N_p}{N_i}} $$

Where:
- $W_i$ is the number of wins after the $i$-th move.
- $N_i$ is the number of simulations that have passed through the $i$-th move.
- $N_p$ is the total number of simulations that have passed through the parent node.
- $C$ is the exploration parameter, which determines the balance between exploration and exploitation.
A higher value of $C$ encourages more exploration.

### How Selection Works

Starting from the root node, the algorithm recursively selects child nodes until it reaches a leaf node (a node with no children or representing a game state that hasn't been fully explored).
At each step, it chooses the child node with the highest UCB1 score.
This process involves:

1. **Calculating UCB1 for Each Child**: For every child of the current node, calculate the UCB1 score using the formula above.
2. **Choosing the Best Child**: Select the child with the highest UCB1 score.
This child is considered the most promising based on current information.
3. **Repeating Until a Leaf Node Is Reached**: Continue this process down the tree until you arrive at a node that either has no children (because it's a terminal state) or hasn't been fully expanded (some moves haven't been explored yet).

### Example Implementation

Here's a simplified version of how the selection process might be implemented in a function in Go.
This function would be part of the `MCTSNode` struct:

```go
func (node *MCTSNode) SelectChild(C float64) *MCTSNode {
    var bestScore float64 = -1
    var bestChild *MCTSNode

    for _, child := range node.children {
        wins := child.wins
        if node.state.playerTurn != child.state.playerTurn {
            // Adjust for the perspective of the player to make a move
            wins = child.visits - child.wins
        }
        ucb1 := wins/child.visits + C*math.Sqrt(math.Log(node.visits)/child.visits)
        if ucb1 > bestScore {
            bestScore = ucb1
            bestChild = child
        }
    }
    return bestChild
}
```

### Key Considerations

- **Exploration Parameter ($C$)**: The choice of $C$ significantly affects the algorithm's behavior.
Common values for $C$ are in the range of $\sqrt{2}$, but the optimal value can depend on the specific game and context.
- **Win Rate Adjustment**: When calculating the UCB1 score, adjust the win rate based on the perspective of the current player, especially in two-player games where the advantage of a move for one player is a disadvantage for the other.

By carefully selecting paths through the game tree, the MCTS algorithm efficiently discovers and focuses on the most promising moves, leading to stronger gameplay performance.

## Expansion
Expansion in Monte Carlo Tree Search (MCTS) is a key phase that occurs after the selection phase.
Once a promising leaf node is selected — that is, a node that represents a game state with unexplored moves — the expansion phase adds one or more new child nodes to the tree, each representing a possible next move from that state.
This phase is crucial for exploring the game's possible outcomes and discovering new strategies.
Here’s how the expansion phase works in detail:

### Goal of Expansion

The primary goal of the expansion phase is to explore new parts of the game tree.
By adding new nodes for unexplored moves, the algorithm incrementally builds a representation of the game's possible states, allowing for a more informed decision-making process in future iterations.

### How Expansion Works

1. **Identify Unexplored Moves**: Starting from the selected leaf node, the algorithm identifies possible moves from the current game state that have not yet been explored.
In games like Tic-Tac-Toe, this would involve looking at the game board and identifying empty spaces where a player can make a move.

2. **Create New Child Nodes**: For each unexplored move identified, the algorithm creates a new child node attached to the current node.
Each child node represents a new game state resulting from making one of the unexplored moves.
The specifics of this process include:
   - Copying the current game state and applying the unexplored move to create the child node's game state.
   - Setting the appropriate player turn for the new game state based on the game's rules.
   - Initializing the win/loss statistics for the new node, typically starting from zero.

3. **Select a Node for Simulation**: Often, immediately after expansion, one of the newly created child nodes is selected for the simulation phase.
The choice of which child node to simulate can vary:
   - The algorithm might randomly select one of the new nodes for simulation.
   - Alternatively, it might use some heuristics or other criteria for selection, depending on the specific implementation and the game's characteristics.

### Example Implementation

Here's a simplified version of what the expansion step might look like in a function as part of the `MCTSNode` struct in Go.
This assumes you have a method to generate possible moves from the current game state and a way to apply those moves to create new game states:

```go
func (node *MCTSNode) Expand() {
    unexploredMoves := node.state.PossibleMoves() // Assume this returns a list of game states
    for _, move := range unexploredMoves {
        newState := node.state.MakeMove(move) // Assume this applies the move and returns a new state
        child := &MCTSNode{
            state: *newState,
            parent: node,
            wins: 0,
            visits: 0,
            children: []*MCTSNode{}, // No children yet
        }
        node.children = append(node.children, child)
    }
}
```

### Key Considerations

- **Single vs.
Multiple Expansions**: Some implementations of MCTS expand only one node per visit to the selected leaf, adding a single new game state based on one possible move.
Others might add multiple children at once.
The choice depends on the specific goals of the algorithm and the computational resources available.
- **Initialization of Node Statistics**: Newly expanded nodes start with no visits and no wins, but depending on the implementation, they might also inherit or simulate some initial values to better integrate into the tree's existing strategy.

By carefully expanding the game tree, MCTS explores new strategies and outcomes, gradually improving its understanding of the game and enhancing its decision-making process.

## Simulation
Simulation, often referred to as the \"playout\" phase in Monte Carlo Tree Search (MCTS), is a crucial step where the algorithm simulates a game from a given node's state to a terminal condition (win, loss, or draw) using a predefined strategy.
This phase follows the expansion of the MCTS tree, where a new node has been added to represent an unexplored move.
The primary goal of the simulation is to gather information about the potential outcome of the game from that point, providing insight into the value of taking a particular path.

### How Simulation Works

1. **Starting Point**: The simulation begins from the state of the newly expanded node.
This state represents a specific configuration of the game (e.g., a Tic-Tac-Toe board layout) after a certain sequence of moves.

2. **Random or Heuristic Play**: The game is simulated forward from this state to a conclusion.
The moves for each player can be selected randomly, which is common in basic MCTS implementations because it's straightforward and unbiased.
Alternatively, lightweight heuristics (simple rules or strategies) can guide move selection to make the simulation more realistic and potentially more informative, though this approach can introduce biases based on the chosen heuristics.

3. **Reaching a Terminal State**: The simulation continues until the game reaches a terminal state, which in the context of Tic-Tac-Toe, occurs when either player wins or all the spaces on the board are filled, resulting in a draw.

4. **Evaluation of Outcome**: Once the game concludes, the outcome is evaluated to determine the winner.
This outcome influences the update process during the backpropagation phase, helping adjust the statistical values of the nodes visited during the selection phase up to the root.

### Purpose of Simulation

The simulation phase aims to estimate the value of reaching the game state represented by the node.
By simulating the game's progression from that state to an end, MCTS gains insights into how promising the path leading to that node might be, without needing to exhaustively search all possible future game states.

### Example Implementation

Here's a conceptual view of how a simulation function could be structured in Go, as part of the MCTSNode struct.
This simplistic version assumes a method to check for terminal states and to randomly select among possible moves:

```go
func (node *MCTSNode) Simulate() int {
    currentState := node.state
    for !currentState.IsGameOver() {
        possibleMoves := currentState.PossibleMoves()
        move := possibleMoves[rand.Intn(len(possibleMoves))] // Randomly select a move
        currentState = currentState.MakeMove(move) // Apply the move
    }
    return currentState.GetWinner() // Return the outcome of the simulation
}
```

### Key Considerations

- **Balance Between Randomness and Strategy**: Purely random simulations provide unbiased estimations of a node's value but might not reflect realistic gameplay, especially in games where certain strategies significantly increase the chances of winning.
Incorporating simple heuristics can improve the quality of simulations but requires careful design to avoid introducing biases that could mislead the search process.

- **Efficiency**: Simulations are the most computationally intensive part of MCTS, as potentially thousands of them may be required to adequately explore the game tree.
Optimizing the simulation process, for example, by limiting the depth of playouts or using efficient methods to select and apply moves, can significantly impact the performance and effectiveness of the MCTS algorithm.

Simulation thus provides a direct way to assess the potential of newly explored moves, informing the algorithm's future selections and expansions by grounding decisions in simulated outcomes.

