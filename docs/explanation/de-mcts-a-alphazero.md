# De MCTS pur a AlphaZero

Ce document explique les **modifications concretes** que l'approche AlphaZero apporte au MCTS, en les mettant en regard du code actuel d'AlphaZeGo.

## Vue d'ensemble des changements

| Phase | MCTS pur (code actuel) | AlphaZero |
|---|---|---|
| **Selection** | UCB1 | PUCT (avec prior du policy network) |
| **Expansion** | Ajoute 1 enfant, pas de prior | Evalue le noeud avec le NN, stocke les priors |
| **Simulation** | Rollout aleatoire jusqu'au terminal | **Supprimee** -- la value du NN remplace le rollout |
| **Backpropagation** | Propage le resultat (win/loss/draw) | Propage la **value** v ∈ [-1, 1] |

## Changement 1 : PUCT remplace UCB1

### UCB1 actuel

La selection dans l'arbre utilise UCB1 (voir [formules.md](../reference/formules.md)) :

```
UCB1(s, a) = Q(s, a) + C × √(ln N(parent) / N(s, a))
```

Le terme d'exploration `√(ln N(parent) / N(s, a))` traite **tous les coups de maniere egale** : un coup brillant et un coup absurde sont explores avec la meme priorite initiale.

Code actuel (`mcts/ucb1.go`) :

```go
avgReward := n.wins / float64(n.visits)
ucbValue := avgReward + C*math.Sqrt(math.Log(float64(n.parent.visits))/float64(n.visits))
```

### PUCT dans AlphaZero

PUCT (Polynomial Upper Confidence Trees, variante de Rosin 2011, adaptee par Silver et al. 2017) remplace le terme d'exploration par un terme **biaise par le policy network** :

```
PUCT(s, a) = Q(s, a) + C_puct × P(s, a) × √N(parent) / (1 + N(s, a))
                                  ↑
                          prior du policy network
```

Differences cles :

1. **`P(s, a)`** : la probabilite a priori du coup `a` selon le policy network. Un coup juge prometteur par le reseau est explore en priorite.
2. **`√N(parent)` au lieu de `√(ln N(parent))`** : croissance plus rapide, exploration plus agressive.
3. **`1 + N(s, a)` au denominateur** : le `+1` assure un score fini meme pour les noeuds non visites (au lieu de +inf dans UCB1). Le score initial d'un noeud non visite est `C_puct × P(s, a) × √N(parent)`, proportionnel a son prior.

**Consequence** : les noeuds ne sont plus explores dans un ordre arbitraire. Le reseau donne une "intuition" qui concentre la recherche sur les coups pertinents. Pour le Go, cela reduit le facteur de branchement effectif de 250 a quelques dizaines.

### Noeuds non visites

Avec UCB1, un noeud non visite a un score **infini**, forçant l'algorithme a essayer chaque coup au moins une fois. 

Avec PUCT, un noeud non visite a un score **proportionnel a son prior**. Un coup avec un prior de 0.01 peut ne jamais etre explore si d'autres coups avec des priors de 0.3 continuent a etre prometteurs. C'est l'elagage implicite du reseau.

## Changement 2 : Expansion guidee par le reseau

### Expansion actuelle

Le code actuel (`mcts/expand.go`) ajoute un enfant sans aucune information a priori :

```go
child := &mctsNode{
    state:    move,
    parent:   node,
    children: []*mctsNode{},
}
```

Le noeud commence avec `wins = 0`, `visits = 0`, et sera selectionne uniquement grace au score UCB1 infini.

### Expansion AlphaZero

Quand un noeud feuille est atteint, on appelle le reseau de neurones **une seule fois** pour obtenir `(p, v)` :

```
(policy, value) = NeuralNetwork(state)
```

Puis on cree **tous les enfants** en leur attribuant leur prior :

```go
policy, value := evaluator.Evaluate(node.state)
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

On backpropage ensuite `value` directement (pas de rollout).

**Pourquoi creer tous les enfants d'un coup ?** Parce que le reseau est appele une seule fois par expansion et retourne les priors pour **tous** les coups. Il n'y a pas de raison de les ajouter un par un.

## Changement 3 : Plus de rollout

### Rollout actuel

Le code actuel (`mcts/simulate.go`) joue des coups aleatoires jusqu'a la fin :

```go
func (node *mctsNode) simulate() board.PlayerID {
    currentState := node.state
    for currentState.Evaluate() == board.NoPlayer {
        possibleMoves := currentState.PossibleMoves()
        currentState = possibleMoves[rand.Intn(len(possibleMoves))]
    }
    return currentState.Evaluate()
}
```

Pour le morpion (profondeur max 9), c'est rapide et raisonnablement informatif. Pour le Go, un rollout aleatoire de 200 coups ne donne presque aucune information utile.

### Remplacement par le value network

Dans AlphaZero, `simulate()` disparait entierement. La valeur `v` retournee par le reseau lors de l'expansion est directement utilisee pour la backpropagation :

```go
// Avant (MCTS pur) :
result := nodeToSimulate.simulate()       // rollout aleatoire → board.PlayerID (1, 2, ou 3)
nodeToSimulate.backpropagate(result)

// Apres (AlphaZero) :
// Le reseau a deja retourne 'value' lors de l'expansion
nodeToSimulate.backpropagateValue(value)   // value ∈ [-1, 1]
```

**Note sur AlphaGo original** (2016) : il utilisait un **melange** des deux :

```
v_final = λ × v_network + (1 - λ) × v_rollout
```

avec `λ = 0.5`. AlphaGo Zero (2017) a montre que le rollout n'apporte rien quand le reseau est assez bon, et l'a supprime.

## Changement 4 : Backpropagation de valeurs continues

### Backpropagation actuelle

Le code actuel propage un resultat discret (1 = Player1 gagne, 2 = Player2, 3 = nul) :

```go
func (n *mctsNode) backpropagate(result board.PlayerID) {
    n.visits++
    playerWhoMovedHere := n.state.PreviousPlayer()
    if result == playerWhoMovedHere {
        n.wins += 1
    } else if result == board.DrawResult {
        n.wins += 0.5
    }
    if n.parent != nil {
        n.parent.backpropagate(result)
    }
}
```

### Backpropagation AlphaZero

La valeur `v ∈ [-1, 1]` est continue. Elle est exprimee **du point de vue du joueur courant** au noeud evalue. Pour rester coherent avec la convention du MCTS pur (ou `wins` est stocke du point de vue du joueur qui a effectue le coup menant a ce noeud), la valeur est d'abord inversee, puis alternee a chaque niveau :

```go
func (node *mctsNode) backpropagateValue(value float64) {
    // Inverser : passer de la perspective du joueur courant
    // a celle du joueur qui a joue le coup (= PreviousPlayer)
    value = -value
    for n := node; n != nil; n = n.parent {
        n.visits++
        n.wins += value
        value = -value  // ← alternance a chaque niveau
    }
}
```

L'inversion initiale garantit que `Q(child) = wins/visits` represente la valeur du point de vue du **parent**, ce qui est necessaire pour que PUCT selectionne les coups favorables au joueur qui choisit.

## Recapitulatif : une iteration AlphaZero

```
1. SELECTION
   node = root
   while node.is_fully_expanded() and not node.is_terminal():
       node = argmax(child: Q(child) + C × P(child) × √N(parent) / (1 + N(child)))
                                         └── PUCT au lieu d'UCB1

2. EXPANSION + EVALUATION (fusionnees, un seul appel reseau)
   if not node.is_terminal():
       policy, value = neural_network(node.state)    ← appel reseau unique
       for each legal move:
           create child with prior = policy[move]     ← expandAll(policy)

3. PAS DE SIMULATION
   (la value du reseau remplace le rollout)

4. BACKPROPAGATION
   value = -value                                     ← inversion initiale (convention MCTS)
   n = node
   while n != nil:
       n.visits += 1
       n.wins += value
       value = -value                                 ← alternance a chaque niveau
       n = n.parent
```

## Cout de l'appel reseau vs rollout

Un appel au reseau de neurones est **beaucoup plus cher** qu'un rollout aleatoire (millisecondes vs microsecondes). Mais :

1. Le reseau est appele **une fois par expansion**, pas une fois par iteration
2. Sa prediction est **bien plus informative** qu'un rollout aleatoire
3. Sur GPU, on peut **batcher** les evaluations (evaluer plusieurs positions en parallele)

AlphaGo Zero utilisait ~1600 iterations MCTS par coup, avec un reseau evalue sur 4 TPU. Le ratio information/calcul est largement favorable au reseau.

## References

- Rosin, "Multi-armed bandits with episode context", Annals of Mathematics and AI 61, 2011 -- PUCT original
- Silver et al., "Mastering the game of Go without human knowledge", Nature 550, 2017 -- Application de PUCT dans AlphaGo Zero
- Voir aussi [reference/formules.md](../reference/formules.md) pour les derivations mathematiques
