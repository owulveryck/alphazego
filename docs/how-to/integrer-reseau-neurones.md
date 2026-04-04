# Intégrer un réseau de neurones dans le MCTS

Guide pratique des modifications à apporter au code d'AlphaZeGo pour passer d'un MCTS pur à un MCTS guidé par réseau de neurones (style AlphaZero).

## Prérequis

Comprendre les concepts derrière ces modifications :

- [L'algorithme MCTS](../explanation/mcts.md)
- [De MCTS à AlphaZero](../explanation/de-mcts-a-alphazero.md)
- [Interfaces Evaluator et Tensorizable](../reference/interfaces-evaluator.md)

## Étape 1 : Les interfaces

Les deux interfaces clés sont déjà définies :

- `Evaluator` dans `mcts/evaluator.go` — fournit policy et value
- `Tensorizable` dans `board/interfaces.go` — convertit un état en tenseur

```go
// mcts/evaluator.go
type Evaluator interface {
    Evaluate(state board.State) (policy []float64, value float64)
}

// board/interfaces.go
type Tensorizable interface {
    Features() []float32
    FeatureShape() [3]int
    ActionSize() int
}
```

Voir [référence/interfaces-evaluator.md](../reference/interfaces-evaluator.md) pour les détails.

## Étape 2 : Implémenter Tensorizable pour le morpion

Fichier : `board/tictactoe/ttt.go`

Ajouter les méthodes `Features()`, `FeatureShape()`, et `ActionSize()` sur `*TicTacToe`. Voir l'exemple dans [référence/interfaces-evaluator.md](../reference/interfaces-evaluator.md).

Vérification :

```go
// Dans un test :
game := tictactoe.NewTicTacToe()
features := game.Features()
assert(len(features) == 27)  // 3 * 3 * 3
assert(game.ActionSize() == 9)
```

## Étape 3 : Ajouter le champ prior au nœud interne

Fichier : `mcts/node.go`

```go
type mctsNode struct {
    // ... champs existants ...
    prior float64 // P(s,a) du policy network (0 si MCTS pur)
}
```

## Étape 4 : Modifier MCTS pour accepter un Evaluator

Fichier : `mcts/mcts.go`

```go
type MCTS struct {
    inventory map[string]*mctsNode
    evaluator Evaluator
    cpuct     float64
}

func NewMCTS() *MCTS {
    return &MCTS{inventory: make(map[string]*mctsNode)}
}

func NewAlphaMCTS(eval Evaluator, cpuct float64) *MCTS {
    return &MCTS{
        inventory: make(map[string]*mctsNode),
        evaluator: eval,
        cpuct:     cpuct,
    }
}
```

## Étape 5 : Implémenter PUCT

Fichier : `mcts/puct.go`

```go
func (n *mctsNode) puct() float64 {
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

## Étape 6 : Modifier selectChildUCB pour utiliser PUCT si disponible

Fichier : `mcts/node.go`

```go
func (n *mctsNode) selectChildUCB() *mctsNode {
    bestScore := math.Inf(-1)
    var bestChild *mctsNode
    for _, child := range n.children {
        var score float64
        if n.mcts != nil && n.mcts.evaluator != nil {
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
```

## Étape 7 : Ajouter expandAll pour l'expansion guidée

Fichier : `mcts/expand.go`

Ajouter une nouvelle méthode `expandAll` qui crée **tous** les enfants et leur attribue leur prior. La méthode `expand()` existante reste inchangée pour le chemin MCTS pur.

```go
// expandAll crée des nœuds enfants pour tous les coups possibles,
// en leur attribuant les priors fournis par le policy network.
func (node *mctsNode) expandAll(policy []float64) {
    possibleMoves := node.state.PossibleMoves()
    for i, move := range possibleMoves {
        child := &mctsNode{
            state:    move,
            parent:   node,
            children: []*mctsNode{},
            prior:    policy[i],
            mcts:     node.mcts,
        }
        node.children = append(node.children, child)
    }
}
```

**Point important** : `expandAll` ne retourne rien et ne fait pas d'appel à l'évaluateur. C'est `RunMCTS` qui appelle l'évaluateur une seule fois et passe la policy à `expandAll` (voir étape 8).

## Étape 8 : Modifier RunMCTS pour utiliser la value au lieu du rollout

Fichier : `mcts/mcts.go`, dans la boucle d'itération.

Le point clé est d'appeler l'évaluateur **une seule fois** par expansion. L'appel retourne à la fois la policy (pour `expandAll`) et la value (pour `backpropagateValue`) :

```go
if !node.isTerminal() && !node.isFullyExpanded() {
    if m.evaluator != nil {
        // AlphaZero : évaluation unique -> expansion + backpropagation
        policy, value := m.evaluator.Evaluate(node.state)
        node.expandAll(policy)
        node.backpropagateValue(value)
    } else {
        // MCTS pur : expansion incrémentale + rollout
        expandedNode := node.expand()
        if expandedNode == nil {
            expandedNode = node
        }
        result := expandedNode.simulate()
        expandedNode.backpropagate(result)
    }
} else if node.isTerminal() {
    // Nœud terminal : résultat connu, pas besoin du réseau
    if m.evaluator != nil {
        value := terminalValue(node.state)
        node.backpropagateValue(value)
    } else {
        result := node.simulate()
        node.backpropagate(result)
    }
}
```

La fonction `terminalValue` convertit un `board.PlayerID` discret en valeur continue du point de vue du joueur courant.

## Étape 9 : Adapter la backpropagation

Fichier : `mcts/backpropagate.go`

Ajouter une variante qui propage une valeur continue. La valeur initiale (du point de vue du joueur courant) est d'abord inversée pour respecter la convention du MCTS pur : `wins` stocke la valeur du point de vue du joueur qui a effectué le coup menant à ce nœud (`PreviousPlayer()`). Ensuite le signe alterne à chaque niveau :

```go
func (node *mctsNode) backpropagateValue(value float64) {
    // Inverser pour passer de la perspective du joueur courant
    // à celle du joueur qui a effectué le coup (convention MCTS)
    value = -value
    for n := node; n != nil; n = n.parent {
        n.visits++
        n.wins += value
        value = -value
    }
}
```

La méthode `backpropagate(result board.PlayerID)` existante reste inchangée pour le chemin MCTS pur.

## Étape 10 : Implémenter l'Evaluator avec ONNX Runtime

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
    // Charger le modèle ONNX
    // Initialiser la session
}

func (e *ONNXEvaluator) Evaluate(state board.State) ([]float64, float64) {
    // 1. Convertir l'état en tenseur via Tensorizable
    t := state.(board.Tensorizable)
    features := t.Features()

    // 2. Appeler le réseau
    // input := ort.NewTensor(features, t.FeatureShape())
    // outputs := e.session.Run(input)

    // 3. Décoder policy (softmax + masquage des coups illégaux)
    // 4. Décoder value (tanh)

    // 5. Retourner policy filtrée sur les coups légaux + value
}
```

## Vérification

1. **Tests de régression** : les tests existants (`TestRunMCTS_TakesWin`, `TestRunMCTS_BlocksWin`) passent avec `NewMCTS()` -- le chemin MCTS pur n'est pas affecté
2. **Tests unitaires** : vérifier PUCT, `expandAll`, `backpropagateValue` avec des valeurs connues
3. **Test d'intégration** : vérifier que `NewAlphaMCTS` avec un évaluateur à rollout (`rolloutEvaluator` dans `puct_test.go`) bloque et prend les victoires
4. **Benchmark** : comparer le temps par itération avec et sans réseau

**État actuel** : les étapes 1-9 sont implémentées et testées. Coverage > 95%.
Pour implémenter un Evaluator, voir [how-to/implementer-evaluator.md](implementer-evaluator.md).

## Prochaines étapes

```
Fait : 1-9 (interfaces, Tensorizable, prior, PUCT, expandAll, backpropagateValue, RunMCTS)
A faire :
  10. Evaluator ONNX (evaluator/onnx.go)     ← nécessite un modèle entraîné
  11. Boucle d'entraînement (Python)          ← hors du scope Go
```
