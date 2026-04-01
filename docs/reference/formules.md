# Formules mathématiques

Référence des formules utilisées dans MCTS et AlphaZero.

## UCB1 (Upper Confidence Bound 1)

**Origine** : Auer, Cesa-Bianchi & Fischer, "Finite-time Analysis of the Multiarmed Bandit Problem", Machine Learning 47, 2002.

### Contexte : le problème du bandit multi-bras

Un joueur fait face à K machines à sous (bandits). À chaque tour, il choisit une machine et observe une récompense. Objectif : maximiser la récompense cumulée. Le dilemme : exploiter la machine qui a le mieux payé jusqu'ici, ou explorer les autres ?

### Formule

```
UCB1(i) = X̄ᵢ + C × √(ln(N) / nᵢ)
```

| Symbole | Signification |
|---|---|
| `X̄ᵢ` | Récompense moyenne du bras `i` (exploitation) |
| `N` | Nombre total de tirages (toutes machines confondues) |
| `nᵢ` | Nombre de tirages du bras `i` |
| `C` | Constante d'exploration (théoriquement `√2`) |
| `ln` | Logarithme naturel |

### Propriétés

- **Regret logarithmique** : UCB1 garantit que le regret (perte par rapport à la stratégie optimale) croît au plus en O(ln n). C'est optimal au sens de la borne de Lai-Robbins.
- **Exploration garantie** : le terme `√(ln(N) / nᵢ)` tend vers l'infini pour un bras ignoré, donc chaque bras est essayé infiniment souvent.
- **Convergence** : `X̄ᵢ` converge vers la vraie moyenne par la loi des grands nombres.

### Application aux arbres (UCT)

**Origine** : Kocsis & Szepesvari, "Bandit based Monte-Carlo Planning", ECML 2006.

UCT (Upper Confidence bounds applied to Trees) applique UCB1 à chaque nœud de l'arbre MCTS. La formule devient :

```
UCT(s, a) = Q(s, a) + C × √(ln(N(s)) / N(s, a))
```

| Symbole | Signification |
|---|---|
| `Q(s, a)` | Valeur moyenne de l'action `a` dans l'état `s` : `wins / visits` |
| `N(s)` | Nombre de visites du nœud parent (état `s`) |
| `N(s, a)` | Nombre de visites du nœud enfant (état après action `a`) |

**Implémentation** (`mcts/ucb1.go`) :

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
**Application à AlphaGo** : Silver et al., Nature 550, 2017.

### Formule

```
PUCT(s, a) = Q(s, a) + C_puct × P(s, a) × √N(s) / (1 + N(s, a))
```

| Symbole | Signification |
|---|---|
| `Q(s, a)` | Valeur moyenne de l'action `a` dans l'état `s` |
| `P(s, a)` | Probabilité a priori de l'action `a` selon le policy network |
| `N(s)` | Nombre de visites du nœud parent |
| `N(s, a)` | Nombre de visites du nœud enfant |
| `C_puct` | Constante d'exploration (AlphaGo Zero utilise ~1.5-2.5) |

### Differences avec UCT

| Aspect | UCT | PUCT |
|---|---|---|
| Prior | Aucun (exploration uniforme) | `P(s, a)` du policy network |
| Terme d'exploration | `√(ln N(s) / N(s, a))` | `√N(s) / (1 + N(s, a))` |
| Nœud non visité | Score = +∞ | Score = `C_puct × P(s, a) × √N(s)` |
| Croissance exploration | Logarithmique | Polynomiale (plus rapide) |
| Source de connaissance | Statistique pure | Apprentissage + statistique |

### Comportement du prior

Le prior `P(s, a)` domine quand `N(s, a)` est petit :

```
N(s, a) = 0  →  PUCT = Q + C × P(s,a) × √N(s)           (prior maximum)
N(s, a) → ∞  →  PUCT ≈ Q + C × P(s,a) × √N(s) / N(s,a)  (prior dilué)
```

Avec suffisamment de visites, le terme `Q(s, a)` domine et le prior n'a plus d'influence. Cela signifie que même si le réseau se trompe sur un coup, MCTS finira par corriger l'erreur grâce aux simulations.

### C_puct adaptatif (AlphaZero)

Dans la version finale d'AlphaZero, `C_puct` dépend du nombre de visites :

```
C_puct(s) = log((1 + N(s) + C_base) / C_base) + C_init
```

avec `C_base = 19652` et `C_init = 1.25`. Cela augmente l'exploration pour les nœuds très visités.

---

## Ajout de bruit de Dirichlet à la racine

**Objectif** : garantir l'exploration même quand le policy network est très confiant.

À la racine de l'arbre uniquement, les priors sont mélangés avec du bruit :

```
P(s, a) = (1 - ε) × p_a + ε × η_a
```

| Symbole | Signification |
|---|---|
| `p_a` | Prior brut du policy network |
| `η_a` | Bruit tiré de Dir(α) (distribution de Dirichlet) |
| `ε` | Poids du bruit (AlphaGo Zero : ε = 0.25) |
| `α` | Paramètre de Dirichlet (Go : α = 0.03, Échecs : α = 0.3) |

Le paramètre `α` est inversement proportionnel au nombre de coups légaux : `α ≈ 10 / nombre_moyen_de_coups`.

---

## Loss function d'entrainement

Le réseau est entraîné pour minimiser :

```
L = (z - v)² - π^T × log(p) + c × ||θ||²
```

| Terme | Signification |
|---|---|
| `(z - v)²` | MSE entre le résultat réel `z ∈ {-1, +1}` et la prédiction value `v` |
| `-π^T × log(p)` | Cross-entropy entre la politique MCTS `π` et la prédiction policy `p` |
| `c × \|\|θ\|\|²` | Régularisation L2 des poids (c = 10⁻⁴ dans AlphaGo Zero) |

La politique MCTS `π` est calculée à partir des compteurs de visite :

```
π(a) = N(s, a)^(1/τ) / Σ_b N(s, b)^(1/τ)
```

où `τ` est un paramètre de **température** :
- `τ = 1` : proportionnel aux visites (exploration, utilisé en début de partie)
- `τ → 0` : sélectionne le coup le plus visité (exploitation, utilisé en fin de partie)

## Références

- Auer, Cesa-Bianchi & Fischer, "Finite-time Analysis of the Multiarmed Bandit Problem", Machine Learning 47, 2002 -- UCB1
- Kocsis & Szepesvari, "Bandit based Monte-Carlo Planning", ECML 2006 -- UCT
- Rosin, "Multi-armed bandits with episode context", Annals of Mathematics and AI 61, 2011 -- PUCT
- Silver et al., "Mastering the game of Go without human knowledge", Nature 550, 2017 -- Application PUCT + Dirichlet + loss function
- Silver et al., "A general reinforcement learning algorithm that masters chess, shogi, and Go through self-play", Science 362, 2018 -- C_puct adaptatif
