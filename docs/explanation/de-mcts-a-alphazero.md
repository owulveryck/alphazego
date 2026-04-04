# De MCTS pur à AlphaZero

Ce document explique les **modifications concrètes** que l'approche AlphaZero apporte au MCTS, en les mettant en regard du code actuel d'AlphaZeGo.

## Vue d'ensemble des changements

| Phase | MCTS pur (code actuel) | AlphaZero |
|---|---|---|
| **Selection** | UCB1 | PUCT (avec prior du policy network) |
| **Expansion** | Ajoute 1 enfant, pas de prior | Évalue le nœud avec le NN, stocke les priors |
| **Simulation** | Rollout aléatoire jusqu'au terminal | **Supprimée** -- la value du NN remplace le rollout |
| **Backpropagation** | Propage le résultat (win/loss/draw) | Propage les **values** par acteur (map[ActorID]float64) |

## Changement 1 : PUCT remplace UCB1

### UCB1 actuel

La sélection dans l'arbre utilise UCB1 (voir [formules.md](../reference/formules.md)) :

```
UCB1(s, a) = Q(s, a) + C × √(ln N(parent) / N(s, a))
```

Le terme d'exploration `√(ln N(parent) / N(s, a))` traite **tous les coups de manière égale** : un coup brillant et un coup absurde sont explorés avec la même priorité initiale.

Code actuel (`mcts/ucb1.go`) :

```go
avgReward := n.wins / float64(n.visits)
ucbValue := avgReward + C*math.Sqrt(math.Log(float64(n.parent.visits))/float64(n.visits))
```

### PUCT dans AlphaZero

PUCT (Polynomial Upper Confidence Trees, variante de Rosin 2011, adaptée par Silver et al. 2017) remplace le terme d'exploration par un terme **biaisé par le policy network** :

```
PUCT(s, a) = Q(s, a) + C_puct × P(s, a) × √N(parent) / (1 + N(s, a))
                                  ↑
                          prior du policy network
```

Différences clés :

1. **`P(s, a)`** : la probabilité a priori du coup `a` selon le policy network. Un coup jugé prometteur par le réseau est exploré en priorité.
2. **`√N(parent)` au lieu de `√(ln N(parent))`** : croissance plus rapide, exploration plus agressive.
3. **`1 + N(s, a)` au dénominateur** : le `+1` assure un score fini même pour les nœuds non visités (au lieu de +inf dans UCB1). Le score initial d'un nœud non visité est `C_puct × P(s, a) × √N(parent)`, proportionnel à son prior.

**Conséquence** : les nœuds ne sont plus explorés dans un ordre arbitraire. Le réseau donne une "intuition" qui concentre la recherche sur les coups pertinents. Pour le Go, cela réduit le facteur de branchement effectif de 250 à quelques dizaines.

### Nœuds non visités

Avec UCB1, un nœud non visité a un score **infini**, forçant l'algorithme à essayer chaque coup au moins une fois. 

Avec PUCT, un nœud non visité a un score **proportionnel à son prior**. Un coup avec un prior de 0.01 peut ne jamais être exploré si d'autres coups avec des priors de 0.3 continuent à être prometteurs. C'est l'élagage implicite du réseau.

## Changement 2 : Expansion guidée par le réseau

### Expansion actuelle

Le code actuel (`mcts/expand.go`) ajoute un enfant sans aucune information à priori :

```go
child := &mctsNode{
    state:    move,
    parent:   node,
    children: []*mctsNode{},
}
```

Le nœud commence avec `wins = 0`, `visits = 0`, et sera sélectionné uniquement grâce au score UCB1 infini.

### Expansion AlphaZero

Quand un nœud feuille est atteint, on appelle le réseau de neurones **une seule fois** pour obtenir les priors et les valeurs par acteur :

```
(policy, values) = NeuralNetwork(state)
```

Puis on crée **tous les enfants** en leur attribuant leur prior :

```go
policy, values := evaluator.Evaluate(node.state)
possibleMoves := node.state.PossibleMoves()

for i, move := range possibleMoves {
    child := &mctsNode{
        state: move,
        parent: node,
        prior: policy[i],  // ← nouveau champ
    }
    node.children = append(node.children, child)
}
```

On backpropage ensuite `values` directement (pas de rollout).

**Pourquoi créer tous les enfants d'un coup ?** Parce que le réseau est appelé une seule fois par expansion et retourne les priors pour **tous** les coups. Il n'y a pas de raison de les ajouter un par un.

## Changement 3 : Plus de rollout

### Rollout actuel

Le code actuel (`mcts/simulate.go`) joue des coups aléatoires jusqu'à la fin :

```go
func (node *mctsNode) simulate() decision.ActorID {
    currentState := node.state
    for currentState.Evaluate() == decision.Undecided {
        possibleMoves := currentState.PossibleMoves()
        currentState = possibleMoves[rand.Intn(len(possibleMoves))]
    }
    return currentState.Evaluate()
}
```

Pour le morpion (profondeur max 9), c'est rapide et raisonnablement informatif. Pour le Go, un rollout aléatoire de 200 coups ne donne presque aucune information utile.

### Remplacement par le value network

Dans AlphaZero, `simulate()` disparaît entièrement. La valeur `v` retournée par le réseau lors de l'expansion est directement utilisée pour la backpropagation :

```go
// Avant (MCTS pur) :
result := nodeToSimulate.simulate()       // rollout aléatoire → decision.ActorID
nodeToSimulate.backpropagate(result)

// Après (AlphaZero) :
// Le réseau a déjà retourné 'values' lors de l'expansion
nodeToSimulate.backpropagateValue(values)  // map[ActorID]float64
```

**Note sur AlphaGo original** (2016) : il utilisait un **mélange** des deux :

```
v_final = λ × v_network + (1 - λ) × v_rollout
```

avec `λ = 0.5`. AlphaGo Zero (2017) a montré que le rollout n'apporte rien quand le réseau est assez bon, et l'a supprimé.

## Changement 4 : Backpropagation de valeurs continues

### Backpropagation actuelle (MCTS pur)

Le code actuel propage un résultat discret (`decision.ActorID` : le gagnant, ou `Stalemate`) :

```go
func (n *mctsNode) backpropagate(result decision.ActorID) {
    for n := node; n != nil; n = n.parent {
        n.visits++
        actorWhoMovedHere := n.state.PreviousActor()
        if result == actorWhoMovedHere {
            n.wins += 1
        } else if result == decision.Stalemate {
            n.wins += 0.5
        }
    }
}
```

### Backpropagation AlphaZero

L'`Evaluator` retourne une **map de valeurs par acteur** (`map[ActorID]float64`). Chaque nœud récupère directement la valeur de l'acteur qui a joué le coup menant à ce nœud :

```go
func (node *mctsNode) backpropagateValue(values map[decision.ActorID]float64) {
    for n := node; n != nil; n = n.parent {
        n.visits++
        n.wins += values[n.state.PreviousActor()]
    }
}
```

Cette approche fonctionne pour tout nombre d'acteurs (1, 2, N) sans hypothèse de somme nulle. `Q(child) = wins/visits` représente la valeur du point de vue de l'acteur qui a joué le coup menant à ce nœud, ce qui est nécessaire pour que PUCT sélectionne les coups favorables à l'acteur qui choisit.

Pour les nœuds terminaux (quand le réseau n'est pas appelé), `backpropagateTerminal()` calcule la valeur à la volée pour chaque acteur :

```go
func (node *mctsNode) backpropagateTerminal() {
    result := node.state.Evaluate()
    for n := node; n != nil; n = n.parent {
        n.visits++
        actor := n.state.PreviousActor()
        if result == actor {
            n.wins += 1.0
        } else if result != decision.Stalemate {
            n.wins += -1.0
        }
    }
}
```

## Récapitulatif : une itération AlphaZero

```
1. SELECTION
   node = root
   while node.is_fully_expanded() and not node.is_terminal():
       node = argmax(child: Q(child) + C × P(child) × √N(parent) / (1 + N(child)))
                                         └── PUCT au lieu d'UCB1

2. EXPANSION + EVALUATION (fusionnées, un seul appel réseau)
   if not node.is_terminal():
       policy, values = neural_network(node.state)    ← appel réseau unique
       for each legal move:
           create child with prior = policy[move]      ← expandAll(policy)

3. PAS DE SIMULATION
   (les values du réseau remplacent le rollout)

4. BACKPROPAGATION
   n = node
   while n != nil:
       n.visits += 1
       n.wins += values[n.state.PreviousActor()]       ← lookup par acteur
       n = n.parent
```

## Coût de l'appel réseau vs rollout

Un appel au réseau de neurones est **beaucoup plus cher** qu'un rollout aléatoire (millisecondes vs microsecondes). Mais :

1. Le réseau est appelé **une fois par expansion**, pas une fois par itération
2. Sa prédiction est **bien plus informative** qu'un rollout aléatoire
3. Sur GPU, on peut **batcher** les évaluations (évaluer plusieurs positions en parallèle)

AlphaGo Zero utilisait ~1600 itérations MCTS par coup, avec un réseau évalué sur 4 TPU. Le ratio information/calcul est largement favorable au réseau.

## Références

- Rosin, "Multi-armed bandits with episode context", Annals of Mathematics and AI 61, 2011 -- PUCT original
- Silver et al., "Mastering the game of Go without human knowledge", Nature 550, 2017 -- Application de PUCT dans AlphaGo Zero
- Voir aussi [référence/formules.md](../reference/formules.md) pour les dérivations mathématiques
