# L'algorithme MCTS (Monte Carlo Tree Search)

## Intuition

Imagine que tu choisisses un restaurant dans une ville inconnue. Tu pourrais :

1. **Exploiter** : retourner au meilleur restaurant que tu connais déjà
2. **Explorer** : essayer un restaurant inconnu qui pourrait être meilleur

MCTS résout exactement ce dilemme, appliqué aux coups d'un jeu. L'algorithme construit progressivement un **arbre de recherche** en équilibrant exploration et exploitation.

## Les quatre phases

Chaque itération de MCTS exécute quatre phases dans l'ordre :

```
     Selection          Expansion        Simulation      Backpropagation
        │                  │                │                  │
        v                  v                v                  v
   ┌────●────┐        ┌────●────┐     ┌────●────┐        ┌────●────┐
   │  / | \  │        │  / | \  │     │  / | \  │        │  / | \  │
   │ ●  ●  ● │        │ ●  ●  ● │     │ ●  ●  ● │        │ ●  ●  ● │
   │    |    │        │    |    │     │    |    │        │    ↑    │
   │    ●    │        │    ●──► │     │    ●    │        │    ●    │
   │ (choisir│        │ (ajouter│     │    |    │        │    ↑    │
   │  le     │        │  un     │     │    ●    │        │    ●    │
   │  chemin)│        │  enfant)│     │    |    │        │ (remonter│
   │         │        │         │     │    ●    │        │  le      │
   │         │        │         │     │ (jouer  │        │  résultat│
   │         │        │         │     │  au     │        │  )       │
   └─────────┘        └─────────┘     │  hasard)│        └─────────┘
                                      └─────────┘
```

### 1. Sélection

À partir de la racine, on descend dans l'arbre en choisissant à chaque nœud l'enfant avec le meilleur score **UCB1** (voir [référence/formules.md](../reference/formules.md)). On continue tant que le nœud est **complètement expansé** (tous ses coups légaux ont déjà été essayés) et **non terminal**.

Dans le code (`mcts/mcts.go`) :

```go
node := root
for !node.isTerminal() && node.isFullyExpanded() {
    node = node.selectChildUCB()
}
```

### 2. Expansion

Quand on atteint un nœud qui n'est pas complètement expansé, on ajoute **un seul** enfant correspondant à un coup non encore exploré.

Dans le code (`mcts/expand.go`) :

```go
func (node *mctsNode) expand() *mctsNode {
    // Trouve le premier coup pas encore expansé
    // Crée un nouveau nœud enfant
    // L'ajoute à node.children
    return child
}
```

Pourquoi un seul enfant à la fois ? Pour ne pas gaspiller du temps à explorer des branches inutiles. On approfondit là où c'est prometteur.

### 3. Simulation (Rollout)

Depuis le nœud nouvellement créé, on joue des **coups aléatoires** jusqu'à la fin de la partie. C'est la méthode Monte Carlo : on estime la valeur d'une position en simulant des parties aléatoires.

Dans le code (`mcts/simulate.go`) :

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

Le rollout aléatoire est la **principale faiblesse** du MCTS pur. Pour le morpion (9 cases), les rollouts sont courts et informatifs. Pour le Go (361 intersections, parties de 200+ coups), un rollout aléatoire est essentiellement du bruit. C'est là que les réseaux de neurones interviennent (voir [alphago-et-reseaux-de-neurones.md](alphago-et-reseaux-de-neurones.md)).

### 4. Backpropagation

Le résultat de la simulation est **propagé vers le haut** de l'arbre, du nœud simulé jusqu'à la racine. Chaque nœud traversé met à jour :

- `visits += 1`
- `wins += 1` si le joueur qui a joué le coup menant à ce nœud a gagné
- `wins += 0.5` en cas de match nul

**Subtilité critique** : à chaque nœud, `CurrentActor()` désigne l'acteur **à qui c'est le tour d'agir**, pas celui qui a joué le coup. L'acteur qui a joué le coup menant à ce nœud est `PreviousActor()`. Les wins doivent être créditées à ce dernier.

Dans le code (`mcts/backpropagate.go`) :

```go
actorWhoMovedHere := n.state.PreviousActor()
if result == actorWhoMovedHere {
    n.wins += 1
} else if result == decision.Stalemate {
    n.wins += 0.5
}
```

## Pourquoi ça converge

Avec suffisamment d'itérations, MCTS converge vers le **minimax** optimal. La raison :

1. UCB1 garantit que chaque branche est visitée **infiniment souvent** (exploration)
2. Mais les branches prometteuses sont visitées **exponentiellement plus** (exploitation)
3. La loi des grands nombres assure que la moyenne des rollouts converge vers la vraie valeur minimax

En pratique, on n'a pas besoin de convergence parfaite : quelques milliers d'itérations suffisent pour le morpion, quelques millions pour le Go.

## Complexité

- **Temps** : O(itérations x profondeur_rollout) par coup
- **Mémoire** : O(nombre de nœuds dans l'arbre) -- croissance proportionnelle aux itérations
- **Avantage clé** : MCTS est **anytime** -- on peut l'arrêter à tout moment et obtenir une réponse raisonnable. Plus on lui donne de temps, meilleure est la réponse.

## Limites du MCTS pur

| Problème | Cause | Solution AlphaGo |
|---|---|---|
| Rollouts aléatoires = bruit sur grands plateaux | Parties trop longues, espace trop vaste | **Value network** remplace le rollout |
| Exploration uniforme des coups | Tous les coups légaux sont traités également | **Policy network** biaise l'exploration |
| Pas d'apprentissage entre parties | Chaque recherche repart de zéro | **Réseau entraîné par self-play** |

## Références

- Kocsis & Szepesvari, "Bandit based Monte-Carlo Planning", ECML 2006 -- Introduction de UCT (UCB1 appliqué aux arbres)
- Coulom, "Efficient Selectivity and Backup Operators in Monte-Carlo Tree Search", CG 2006 -- Première application de MCTS au Go
- Browne et al., "A Survey of Monte Carlo Tree Search Methods", IEEE 2012 -- Survey de référence
