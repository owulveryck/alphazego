# Formules mathematiques

Reference des formules utilisees dans MCTS et AlphaZero.

## UCB1 (Upper Confidence Bound 1)

**Origine** : Auer, Cesa-Bianchi & Fischer, "Finite-time Analysis of the Multiarmed Bandit Problem", Machine Learning 47, 2002.

### Contexte : le probleme du bandit multi-bras

Un joueur fait face a K machines a sous (bandits). A chaque tour, il choisit une machine et observe une recompense. Objectif : maximiser la recompense cumulee. Le dilemme : exploiter la machine qui a le mieux paye jusqu'ici, ou explorer les autres ?

### Formule

```
UCB1(i) = X̄ᵢ + C × √(ln(N) / nᵢ)
```

| Symbole | Signification |
|---|---|
| `X̄ᵢ` | Recompense moyenne du bras `i` (exploitation) |
| `N` | Nombre total de tirages (toutes machines confondues) |
| `nᵢ` | Nombre de tirages du bras `i` |
| `C` | Constante d'exploration (theoriquement `√2`) |
| `ln` | Logarithme naturel |

### Proprietes

- **Regret logarithmique** : UCB1 garantit que le regret (perte par rapport a la strategie optimale) croit au plus en O(ln n). C'est optimal au sens de la borne de Lai-Robbins.
- **Exploration garantie** : le terme `√(ln(N) / nᵢ)` tend vers l'infini pour un bras ignore, donc chaque bras est essaye infiniment souvent.
- **Convergence** : `X̄ᵢ` converge vers la vraie moyenne par la loi des grands nombres.

### Application aux arbres (UCT)

**Origine** : Kocsis & Szepesvari, "Bandit based Monte-Carlo Planning", ECML 2006.

UCT (Upper Confidence bounds applied to Trees) applique UCB1 a chaque noeud de l'arbre MCTS. La formule devient :

```
UCT(s, a) = Q(s, a) + C × √(ln(N(s)) / N(s, a))
```

| Symbole | Signification |
|---|---|
| `Q(s, a)` | Valeur moyenne de l'action `a` dans l'etat `s` : `wins / visits` |
| `N(s)` | Nombre de visites du noeud parent (etat `s`) |
| `N(s, a)` | Nombre de visites du noeud enfant (etat apres action `a`) |

**Implementation** (`mcts/ucb1.go`) :

```go
avgReward := n.wins / float64(n.visits)                    // Q(s, a)
C := math.Sqrt(2)                                          // constante d'exploration
ucbValue := avgReward + C*math.Sqrt(
    math.Log(float64(n.parent.visits)) / float64(n.visits)) // UCT complet
```

**Cas special** : si `N(s, a) = 0`, UCB1 = +∞, ce qui force l'exploration de chaque action au moins une fois.

---

## PUCT (Polynomial Upper Confidence Trees)

**Origine** : Rosin, "Multi-armed bandits with episode context", Annals of Mathematics and AI 61, 2011.
**Application a AlphaGo** : Silver et al., Nature 550, 2017.

### Formule

```
PUCT(s, a) = Q(s, a) + C_puct × P(s, a) × √N(s) / (1 + N(s, a))
```

| Symbole | Signification |
|---|---|
| `Q(s, a)` | Valeur moyenne de l'action `a` dans l'etat `s` |
| `P(s, a)` | Probabilite a priori de l'action `a` selon le policy network |
| `N(s)` | Nombre de visites du noeud parent |
| `N(s, a)` | Nombre de visites du noeud enfant |
| `C_puct` | Constante d'exploration (AlphaGo Zero utilise ~1.5-2.5) |

### Differences avec UCT

| Aspect | UCT | PUCT |
|---|---|---|
| Prior | Aucun (exploration uniforme) | `P(s, a)` du policy network |
| Terme d'exploration | `√(ln N(s) / N(s, a))` | `√N(s) / (1 + N(s, a))` |
| Noeud non visite | Score = +∞ | Score = `C_puct × P(s, a) × √N(s)` |
| Croissance exploration | Logarithmique | Polynomiale (plus rapide) |
| Source de connaissance | Statistique pure | Apprentissage + statistique |

### Comportement du prior

Le prior `P(s, a)` domine quand `N(s, a)` est petit :

```
N(s, a) = 0  →  PUCT = Q + C × P(s,a) × √N(s)           (prior maximum)
N(s, a) → ∞  →  PUCT ≈ Q + C × P(s,a) × √N(s) / N(s,a)  (prior dilue)
```

Avec suffisamment de visites, le terme `Q(s, a)` domine et le prior n'a plus d'influence. Cela signifie que meme si le reseau se trompe sur un coup, MCTS finira par corriger l'erreur grace aux simulations.

### C_puct adaptatif (AlphaZero)

Dans la version finale d'AlphaZero, `C_puct` depend du nombre de visites :

```
C_puct(s) = log((1 + N(s) + C_base) / C_base) + C_init
```

avec `C_base = 19652` et `C_init = 1.25`. Cela augmente l'exploration pour les noeuds tres visites.

---

## Ajout de bruit de Dirichlet a la racine

**Objectif** : garantir l'exploration meme quand le policy network est tres confiant.

A la racine de l'arbre uniquement, les priors sont melanges avec du bruit :

```
P(s, a) = (1 - ε) × p_a + ε × η_a
```

| Symbole | Signification |
|---|---|
| `p_a` | Prior brut du policy network |
| `η_a` | Bruit tire de Dir(α) (distribution de Dirichlet) |
| `ε` | Poids du bruit (AlphaGo Zero : ε = 0.25) |
| `α` | Parametre de Dirichlet (Go : α = 0.03, Echecs : α = 0.3) |

Le parametre `α` est inversement proportionnel au nombre de coups legaux : `α ≈ 10 / nombre_moyen_de_coups`.

---

## Loss function d'entrainement

Le reseau est entraine pour minimiser :

```
L = (z - v)² - π^T × log(p) + c × ||θ||²
```

| Terme | Signification |
|---|---|
| `(z - v)²` | MSE entre le resultat reel `z ∈ {-1, +1}` et la prediction value `v` |
| `-π^T × log(p)` | Cross-entropy entre la politique MCTS `π` et la prediction policy `p` |
| `c × \|\|θ\|\|²` | Regularisation L2 des poids (c = 10⁻⁴ dans AlphaGo Zero) |

La politique MCTS `π` est calculee a partir des compteurs de visite :

```
π(a) = N(s, a)^(1/τ) / Σ_b N(s, b)^(1/τ)
```

ou `τ` est un parametre de **temperature** :
- `τ = 1` : proportionnel aux visites (exploration, utilise en debut de partie)
- `τ → 0` : selectionne le coup le plus visite (exploitation, utilise en fin de partie)

## References

- Auer, Cesa-Bianchi & Fischer, "Finite-time Analysis of the Multiarmed Bandit Problem", Machine Learning 47, 2002 -- UCB1
- Kocsis & Szepesvari, "Bandit based Monte-Carlo Planning", ECML 2006 -- UCT
- Rosin, "Multi-armed bandits with episode context", Annals of Mathematics and AI 61, 2011 -- PUCT
- Silver et al., "Mastering the game of Go without human knowledge", Nature 550, 2017 -- Application PUCT + Dirichlet + loss function
- Silver et al., "A general reinforcement learning algorithm that masters chess, shogi, and Go through self-play", Science 362, 2018 -- C_puct adaptatif
