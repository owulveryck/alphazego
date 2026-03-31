# Integrer un reseau de neurones dans le MCTS

Guide pratique des modifications a apporter au code d'AlphaZeGo pour passer d'un MCTS pur a un MCTS guide par reseau de neurones (style AlphaZero).

## Prerequis

Comprendre les concepts decriere ces modifications :

- [L'algorithme MCTS](../explanation/mcts.md)
- [De MCTS a AlphaZero](../explanation/de-mcts-a-alphazero.md)
- [Interfaces Evaluator et Tensorizable](../reference/interfaces-evaluator.md)

## Etape 1 : Ajouter les interfaces

Fichier : `board/interfaces.go`

Ajouter les interfaces `Evaluator` et `Tensorizable` (voir [reference/interfaces-evaluator.md](../reference/interfaces-evaluator.md) pour le code complet).

```go
type Evaluator interface {
    Evaluate(state State) (policy []float64, value float64)
}

type Tensorizable interface {
    Features() []float32
    FeatureShape() [3]int
    ActionSize() int
}
```

## Etape 2 : Implementer Tensorizable pour le morpion

Fichier : `board/tictactoe/ttt.go`

Ajouter les methodes `Features()`, `FeatureShape()`, et `ActionSize()` sur `*TicTacToe`. Voir l'exemple dans [reference/interfaces-evaluator.md](../reference/interfaces-evaluator.md).

Verification :

```go
// Dans un test :
game := tictactoe.NewTicTacToe()
features := game.Features()
assert(len(features) == 27)  // 3 * 3 * 3
assert(game.ActionSize() == 9)
```

## Etape 3 : Ajouter le champ prior au MCTSNode

Fichier : `mcts/node.go`

```go
type MCTSNode struct {
    // ... champs existants ...
    prior float64 // P(s,a) du policy network (0 si MCTS pur)
}
```

## Etape 4 : Modifier MCTS pour accepter un Evaluator

Fichier : `mcts/mcts.go`

```go
type MCTS struct {
    inventory map[string]*MCTSNode
    evaluator board.Evaluator
    cpuct     float64
}

func NewMCTS() *MCTS {
    return &MCTS{inventory: make(map[string]*MCTSNode)}
}

func NewAlphaMCTS(eval board.Evaluator, cpuct float64) *MCTS {
    return &MCTS{
        inventory: make(map[string]*MCTSNode),
        evaluator: eval,
        cpuct:     cpuct,
    }
}
```

## Etape 5 : Implementer PUCT

Fichier : `mcts/puct.go` (nouveau)

```go
func (n *MCTSNode) PUCT() float64 {
    if n.visits == 0 {
        if n.parent == nil {
            return n.prior
        }
        return n.mcts.cpuct * n.prior * math.Sqrt(n.parent.visits)
    }
    q := n.wins / n.visits
    if n.parent == nil {
        return q
    }
    return q + n.mcts.cpuct*n.prior*math.Sqrt(n.parent.visits)/(1+n.visits)
}
```

## Etape 6 : Modifier SelectChildUCB pour utiliser PUCT si disponible

Fichier : `mcts/node.go`

```go
func (n *MCTSNode) SelectChildUCB() *MCTSNode {
    bestScore := math.Inf(-1)
    var bestChild *MCTSNode
    for _, child := range n.children {
        var score float64
        if n.mcts != nil && n.mcts.evaluator != nil {
            score = child.PUCT()
        } else {
            score = child.UCB1()
        }
        if score > bestScore {
            bestScore = score
            bestChild = child
        }
    }
    return bestChild
}
```

## Etape 7 : Ajouter ExpandAll pour l'expansion guidee

Fichier : `mcts/expand.go`

Ajouter une nouvelle methode `ExpandAll` qui cree **tous** les enfants et leur attribue leur prior. L'`Expand()` existant reste inchange pour le chemin MCTS pur.

```go
// ExpandAll cree des noeuds enfants pour tous les coups possibles,
// en leur attribuant les priors fournis par le policy network.
func (node *MCTSNode) ExpandAll(policy []float64) {
    possibleMoves := node.state.PossibleMoves()
    for i, move := range possibleMoves {
        child := &MCTSNode{
            state:    move,
            parent:   node,
            children: []*MCTSNode{},
            prior:    policy[i],
            mcts:     node.mcts,
        }
        node.children = append(node.children, child)
    }
}
```

**Point important** : `ExpandAll` ne retourne rien et ne fait pas d'appel a l'evaluateur. C'est `RunMCTS` qui appelle l'evaluateur une seule fois et passe la policy a `ExpandAll` (voir etape 8).

## Etape 8 : Modifier RunMCTS pour utiliser la value au lieu du rollout

Fichier : `mcts/mcts.go`, dans la boucle d'iteration.

Le point cle est d'appeler l'evaluateur **une seule fois** par expansion. L'appel retourne a la fois la policy (pour `ExpandAll`) et la value (pour `BackpropagateValue`) :

```go
if !node.IsTerminal() && !node.IsFullyExpanded() {
    if m.evaluator != nil {
        // AlphaZero : evaluation unique → expansion + backpropagation
        policy, value := m.evaluator.Evaluate(node.state)
        node.ExpandAll(policy)
        node.BackpropagateValue(value)
    } else {
        // MCTS pur : expansion incrementale + rollout
        expandedNode := node.Expand()
        if expandedNode == nil {
            expandedNode = node
        }
        result := expandedNode.Simulate()
        expandedNode.Backpropagate(result)
    }
} else if node.IsTerminal() {
    // Noeud terminal : resultat connu, pas besoin du reseau
    if m.evaluator != nil {
        value := terminalValue(node.state)
        node.BackpropagateValue(value)
    } else {
        result := node.Simulate()
        node.Backpropagate(result)
    }
}
```

La fonction `terminalValue` convertit un `board.PlayerID` discret en valeur continue du point de vue du joueur courant.

## Etape 9 : Adapter la backpropagation

Fichier : `mcts/backpropagate.go`

Ajouter une variante qui propage une valeur continue. La valeur initiale (du point de vue du joueur courant) est d'abord inversee pour respecter la convention du MCTS pur : `wins` stocke la valeur du point de vue du joueur qui a effectue le coup menant a ce noeud (`PreviousPlayer()`). Ensuite le signe alterne a chaque niveau :

```go
func (node *MCTSNode) BackpropagateValue(value float64) {
    // Inverser pour passer de la perspective du joueur courant
    // a celle du joueur qui a effectue le coup (convention MCTS)
    value = -value
    for n := node; n != nil; n = n.parent {
        n.visits++
        n.wins += value
        value = -value
    }
}
```

L'`Backpropagate(result board.PlayerID)` existant reste inchange pour le chemin MCTS pur.

## Etape 10 : Implementer l'Evaluator avec ONNX Runtime

Fichier : `evaluator/onnx.go` (nouveau package)

```go
package evaluator

import (
    ort "github.com/yalue/onnxruntime_go"
    "github.com/owulveryck/alphazego/board"
)

type ONNXEvaluator struct {
    session *ort.Session
}

func NewONNXEvaluator(modelPath string) (*ONNXEvaluator, error) {
    // Charger le modele ONNX
    // Initialiser la session
}

func (e *ONNXEvaluator) Evaluate(state board.State) ([]float64, float64) {
    // 1. Convertir l'etat en tenseur via Tensorizable
    t := state.(board.Tensorizable)
    features := t.Features()

    // 2. Appeler le reseau
    // input := ort.NewTensor(features, t.FeatureShape())
    // outputs := e.session.Run(input)

    // 3. Decoder policy (softmax + masquage des coups illegaux)
    // 4. Decoder value (tanh)

    // 5. Retourner policy filtree sur les coups legaux + value
}
```

## Verification

1. **Tests de regression** : les tests existants (`TestRunMCTS_TakesWin`, `TestRunMCTS_BlocksWin`) passent avec `NewMCTS()` -- le chemin MCTS pur n'est pas affecte
2. **Tests unitaires** : verifier PUCT, `ExpandAll`, `BackpropagateValue` avec des valeurs connues
3. **Test d'integration** : verifier que `NewAlphaMCTS` avec un evaluateur a rollout (`rolloutEvaluator` dans `puct_test.go`) bloque et prend les victoires
4. **Benchmark** : comparer le temps par iteration avec et sans reseau

**Etat actuel** : les etapes 1-9 sont implementees et testees. Coverage > 95%.
Pour implementer un Evaluator, voir [how-to/implementer-evaluator.md](implementer-evaluator.md).

## Prochaines etapes

```
Fait : 1-9 (interfaces, Tensorizable, prior, PUCT, ExpandAll, BackpropagateValue, RunMCTS)
A faire :
  10. Evaluator ONNX (evaluator/onnx.go)     ← necessite un modele entraine
  11. Boucle d'entrainement (Python)          ← hors du scope Go
```
