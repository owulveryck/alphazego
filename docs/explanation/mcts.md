# L'algorithme MCTS (Monte Carlo Tree Search)

## Intuition

Imagine que tu choisisses un restaurant dans une ville inconnue. Tu pourrais :

1. **Exploiter** : retourner au meilleur restaurant que tu connais deja
2. **Explorer** : essayer un restaurant inconnu qui pourrait etre meilleur

MCTS resout exactement ce dilemme, applique aux coups d'un jeu. L'algorithme construit progressivement un **arbre de recherche** en equilibrant exploration et exploitation.

## Les quatre phases

Chaque iteration de MCTS execute quatre phases dans l'ordre :

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
   │         │        │         │     │ (jouer  │        │  resultat│
   │         │        │         │     │  au     │        │  )       │
   └─────────┘        └─────────┘     │  hasard)│        └─────────┘
                                      └─────────┘
```

### 1. Selection

A partir de la racine, on descend dans l'arbre en choisissant a chaque noeud l'enfant avec le meilleur score **UCB1** (voir [reference/formules.md](../reference/formules.md)). On continue tant que le noeud est **completement expanse** (tous ses coups legaux ont deja ete essayes) et **non terminal**.

Dans le code (`mcts/mcts.go`) :

```go
node := root
for !node.IsTerminal() && node.IsFullyExpanded() {
    node = node.SelectChildUCB()
}
```

### 2. Expansion

Quand on atteint un noeud qui n'est pas completement expanse, on ajoute **un seul** enfant correspondant a un coup non encore explore.

Dans le code (`mcts/expand.go`) :

```go
func (node *MCTSNode) Expand() *MCTSNode {
    // Trouve le premier coup pas encore expanse
    // Cree un nouveau noeud enfant
    // L'ajoute a node.children
    return child
}
```

Pourquoi un seul enfant a la fois ? Pour ne pas gaspiller du temps a explorer des branches inutiles. On approfondit la ou c'est prometteur.

### 3. Simulation (Rollout)

Depuis le noeud nouvellement cree, on joue des **coups aleatoires** jusqu'a la fin de la partie. C'est la methode Monte Carlo : on estime la valeur d'une position en simulant des parties aleatoires.

Dans le code (`mcts/simulate.go`) :

```go
func (node *MCTSNode) Simulate() board.Result {
    currentState := node.state
    for currentState.Evaluate() == board.GameOn {
        possibleMoves := currentState.PossibleMoves()
        currentState = possibleMoves[rand.Intn(len(possibleMoves))]
    }
    return currentState.Evaluate()
}
```

Le rollout aleatoire est la **principale faiblesse** du MCTS pur. Pour le morpion (9 cases), les rollouts sont courts et informatifs. Pour le Go (361 intersections, parties de 200+ coups), un rollout aleatoire est essentiellement du bruit. C'est la que les reseaux de neurones interviennent (voir [alphago-et-reseaux-de-neurones.md](alphago-et-reseaux-de-neurones.md)).

### 4. Backpropagation

Le resultat de la simulation est **propage vers le haut** de l'arbre, du noeud simule jusqu'a la racine. Chaque noeud traverse met a jour :

- `visits += 1`
- `wins += 1` si le joueur qui a joue le coup menant a ce noeud a gagne
- `wins += 0.5` en cas de match nul

**Subtilite critique** : a chaque noeud, `CurrentPlayer()` designe le joueur **a qui c'est le tour de jouer**, pas celui qui a joue le coup. Le joueur qui a joue le coup menant a ce noeud est `3 - CurrentPlayer()`. Les wins doivent etre creditees a ce dernier.

Dans le code (`mcts/backpropagate.go`) :

```go
playerWhoMovedHere := 3 - n.state.CurrentPlayer()
if result == playerWhoMovedHere {
    n.wins += 1
} else if result == board.Draw {
    n.wins += 0.5
}
```

## Pourquoi ca converge

Avec suffisamment d'iterations, MCTS converge vers le **minimax** optimal. La raison :

1. UCB1 garantit que chaque branche est visitee **infiniment souvent** (exploration)
2. Mais les branches prometteuses sont visitees **exponentiellement plus** (exploitation)
3. La loi des grands nombres assure que la moyenne des rollouts converge vers la vraie valeur minimax

En pratique, on n'a pas besoin de convergence parfaite : quelques milliers d'iterations suffisent pour le morpion, quelques millions pour le Go.

## Complexite

- **Temps** : O(iterations x profondeur_rollout) par coup
- **Memoire** : O(nombre de noeuds dans l'arbre) -- croissance proportionnelle aux iterations
- **Avantage cle** : MCTS est **anytime** -- on peut l'arreter a tout moment et obtenir une reponse raisonnable. Plus on lui donne de temps, meilleure est la reponse.

## Limites du MCTS pur

| Probleme | Cause | Solution AlphaGo |
|---|---|---|
| Rollouts aleatoires = bruit sur grands plateaux | Parties trop longues, espace trop vaste | **Value network** remplace le rollout |
| Exploration uniforme des coups | Tous les coups legaux sont traites egalement | **Policy network** biaise l'exploration |
| Pas d'apprentissage entre parties | Chaque recherche repart de zero | **Reseau entraine par self-play** |

## References

- Kocsis & Szepesvari, "Bandit based Monte-Carlo Planning", ECML 2006 -- Introduction de UCT (UCB1 applique aux arbres)
- Coulom, "Efficient Selectivity and Backup Operators in Monte-Carlo Tree Search", CG 2006 -- Premiere application de MCTS au Go
- Browne et al., "A Survey of Monte Carlo Tree Search Methods", IEEE 2012 -- Survey de reference
