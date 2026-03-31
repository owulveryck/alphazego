# Integrer un reseau de neurones dans le MCTS

Guide pratique des modifications a apporter au code d'AlphaZeGo pour passer d'un MCTS pur a un MCTS guide par reseau de neurones (style AlphaZero).

## Prerequis

Comprendre les concepts decriere ces modifications :

- [L'algorithme MCTS](../explanation/mcts.md)
- [De MCTS a AlphaZero](../explanation/de-mcts-a-alphazero.md)
- [Interfaces Evaluator et Tensorizable](../reference/interfaces-evaluator.md)

## Etape 1 : Les interfaces

Les deux interfaces cles sont deja definies :

- `Evaluator` dans `mcts/evaluator.go` — fournit policy et value
- `Tensorizable` dans `board/interfaces.go` — convertit un etat en tenseur

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

Voir [reference/interfaces-evaluator.md](../reference/interfaces-evaluator.md) pour les details.

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

## Etape 3 : Ajouter le champ prior au noeud interne

Fichier : `mcts/node.go`

```go
type mctsNode struct {
    // ... champs existants ...
    prior float64 // P(s,a) du policy network (0 si MCTS pur)
}
```

## Etape 4 : Modifier MCTS pour accepter un Evaluator

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

## Etape 5 : Implementer PUCT

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

## Etape 6 : Modifier selectChildUCB pour utiliser PUCT si disponible

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

## Etape 7 : Ajouter expandAll pour l'expansion guidee

Fichier : `mcts/expand.go`

Ajouter une nouvelle methode `expandAll` qui cree **tous** les enfants et leur attribue leur prior. La methode `expand()` existante reste inchangee pour le chemin MCTS pur.

```go
// expandAll cree des noeuds enfants pour tous les coups possibles,
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

**Point important** : `expandAll` ne retourne rien et ne fait pas d'appel a l'evaluateur. C'est `RunMCTS` qui appelle l'evaluateur une seule fois et passe la policy a `expandAll` (voir etape 8).

## Etape 8 : Modifier RunMCTS pour utiliser la value au lieu du rollout

Fichier : `mcts/mcts.go`, dans la boucle d'iteration.

Le point cle est d'appeler l'evaluateur **une seule fois** par expansion. L'appel retourne a la fois la policy (pour `expandAll`) et la value (pour `backpropagateValue`) :

```go
if !node.isTerminal() && !node.isFullyExpanded() {
    if m.evaluator != nil {
        // AlphaZero : evaluation unique → expansion + backpropagation
        policy, value := m.evaluator.Evaluate(node.state)
        node.expandAll(policy)
        node.backpropagateValue(value)
    } else {
        // MCTS pur : expansion incrementale + rollout
        expandedNode := node.expand()
        if expandedNode == nil {
            expandedNode = node
        }
        result := expandedNode.simulate()
        expandedNode.backpropagate(result)
    }
} else if node.isTerminal() {
    // Noeud terminal : resultat connu, pas besoin du reseau
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

## Etape 9 : Adapter la backpropagation

Fichier : `mcts/backpropagate.go`

Ajouter une variante qui propage une valeur continue. La valeur initiale (du point de vue du joueur courant) est d'abord inversee pour respecter la convention du MCTS pur : `wins` stocke la valeur du point de vue du joueur qui a effectue le coup menant a ce noeud (`PreviousPlayer()`). Ensuite le signe alterne a chaque niveau :

```go
func (node *mctsNode) backpropagateValue(value float64) {
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

La methode `backpropagate(result board.PlayerID)` existante reste inchangee pour le chemin MCTS pur.

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
2. **Tests unitaires** : verifier PUCT, `expandAll`, `backpropagateValue` avec des valeurs connues
3. **Test d'integration** : verifier que `NewAlphaMCTS` avec un evaluateur a rollout (`rolloutEvaluator` dans `puct_test.go`) bloque et prend les victoires
4. **Benchmark** : comparer le temps par iteration avec et sans reseau

**Etat actuel** : les etapes 1-9 sont implementees et testees. Coverage > 95%.
Pour implementer un Evaluator, voir [how-to/implementer-evaluator.md](implementer-evaluator.md).

## Prochaines etapes

```
Fait : 1-9 (interfaces, Tensorizable, prior, PUCT, expandAll, backpropagateValue, RunMCTS)
A faire :
  10. Evaluator ONNX (evaluator/onnx.go)     ← necessite un modele entraine
  11. Boucle d'entrainement (Python)          ← hors du scope Go
```
