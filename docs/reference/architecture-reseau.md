# Architecture du réseau de neurones

Référence technique de l'architecture utilisée dans AlphaGo Zero / AlphaZero.

## Vue d'ensemble

```
Input [C x H x W]
       │
       v
  Conv 3x3, 256 filtres
  BatchNorm
  ReLU
       │
       v
  ┌─────────────────┐
  │ Bloc résiduel x19│  (ou x39 pour la version large)
  │                  │
  │  Conv 3x3, 256   │
  │  BatchNorm        │
  │  ReLU             │
  │  Conv 3x3, 256   │
  │  BatchNorm        │
  │  + skip connection│
  │  ReLU             │
  └────────┬─────────┘
           │
     ┌─────┴──────┐
     │             │
     v             v
  Policy Head   Value Head
```

## Tenseur d'entrée

### AlphaGo Zero (Go 19x19)

Dimensions : `[17][19][19]`

| Plans | Description | Nombre |
|---|---|---|
| `X_t, X_{t-1}, ..., X_{t-7}` | Pierres du joueur courant aux 8 derniers pas de temps | 8 |
| `Y_t, Y_{t-1}, ..., Y_{t-7}` | Pierres de l'adversaire aux 8 derniers pas de temps | 8 |
| `C` | Couleur du joueur courant (plan constant : 1 si noir, 0 si blanc) | 1 |

Chaque plan est une matrice binaire 19x19.

L'historique (8 pas de temps) permet au réseau de détecter les **répétitions** et les situations de **ko** sans encodage explicite des règles.

### Morpion (3x3) -- proposition minimaliste

Dimensions : `[3][3][3]`

| Plan | Description |
|---|---|
| Plan 0 | Positions du joueur courant (1 si occupé, 0 sinon) |
| Plan 1 | Positions de l'adversaire |
| Plan 2 | Joueur courant (plan constant : 1.0 ou 0.0) |

Pas besoin d'historique pour le morpion (pas de répétition possible).

## Tronc convolutif (shared trunk)

### Couche initiale

```
Conv2D(in=C, out=256, kernel=3x3, padding=1)
BatchNorm2D(256)
ReLU
```

Le padding de 1 preserve les dimensions spatiales : `[256][H][W]`.

### Blocs résiduels

Chaque bloc résiduel (He et al., 2016) :

```
Input x ──────────────────────┐
    │                         │ (skip connection)
    v                         │
  Conv2D(256, 256, 3x3, pad=1)│
  BatchNorm2D(256)             │
  ReLU                        │
    │                         │
    v                         │
  Conv2D(256, 256, 3x3, pad=1)│
  BatchNorm2D(256)             │
    │                         │
    v                         │
  + ◄─────────────────────────┘
    │
    v
  ReLU
    │
    v
  Output
```

Nombre de blocs :
- AlphaGo Zero (version standard) : **19 blocs**, 256 filtres
- AlphaGo Zero (version large) : **39 blocs**, 256 filtres
- AlphaZero : **20 blocs**, 256 filtres

Le skip connection résout le problème du **gradient qui s'évanouit** (vanishing gradient) et permet d'entraîner des réseaux très profonds (40+ couches).

## Tete Policy

```
Trunk output [256][H][W]
       │
       v
  Conv2D(256, 2, 1x1)         ← réduction à 2 canaux
  BatchNorm2D(2)
  ReLU
       │
       v
  Flatten → [2 × H × W]
       │
       v
  Linear(2 × H × W, action_size)   ← action_size = H×W+1 (Go) ou H×W (morpion)
       │
       v
  Softmax
       │
       v
  p ∈ R^action_size            ← distribution de probabilites
```

Pour le Go : `action_size = 362` (361 intersections + 1 passe)
Pour le morpion : `action_size = 9`

Les coups illégaux sont **masqués** avant le softmax : leur logit est mis à -∞, et les probabilités restantes sont renormalisées.

## Tete Value

```
Trunk output [256][H][W]
       │
       v
  Conv2D(256, 1, 1x1)         ← réduction à 1 canal
  BatchNorm2D(1)
  ReLU
       │
       v
  Flatten → [H × W]
       │
       v
  Linear(H × W, 256)
  ReLU
       │
       v
  Linear(256, 1)
       │
       v
  tanh
       │
       v
  v ∈ [-1, 1]                 ← estimation de victoire
```

Le `tanh` borne la sortie entre -1 et +1 :
- `v = +1` : victoire certaine du joueur courant
- `v = 0` : position équilibrée
- `v = -1` : défaite certaine

## Hyperparamètres d'entraînement (AlphaGo Zero)

| Paramètre | Valeur |
|---|---|
| Optimiseur | SGD avec momentum 0.9 |
| Learning rate | 0.01 (divisé par 10 à 400k et 600k steps) |
| Batch size | 2048 |
| L2 régularisation | c = 10⁻⁴ |
| Iterations MCTS par coup | 1600 |
| Bruit Dirichlet (α) | 0.03 |
| Poids du bruit (ε) | 0.25 |
| Température (τ) | 1.0 pour les 30 premiers coups, puis → 0 |
| Parties de self-play | 4.9 millions |
| Durée d'entraînement | 3 jours sur 4 TPU |

## Dimensionnement pour le morpion

Pour un réseau minimaliste adapté au morpion :

| Paramètre | AlphaGo Zero | Morpion (suggestion) |
|---|---|---|
| Taille du plateau | 19x19 | 3x3 |
| Plans d'entrée | 17 | 3 |
| Filtres convolutifs | 256 | 32 |
| Blocs résiduels | 19-39 | 2-4 |
| Taille d'action | 362 | 9 |
| Itérations MCTS | 1600 | 100-400 |

Un réseau aussi petit peut s'entraîner en **quelques minutes** sur un CPU, ce qui est idéal pour expérimenter.

## Références

- He et al., "Deep Residual Learning for Image Recognition", CVPR 2016 -- Architecture ResNet
- Ioffe & Szegedy, "Batch Normalization: Accelerating Deep Network Training", ICML 2015 -- BatchNorm
- Silver et al., "Mastering the game of Go without human knowledge", Nature 550, 2017 -- Architecture AlphaGo Zero (Figure 2, Methods)
- Silver et al., "A general reinforcement learning algorithm that masters chess, shogi, and Go through self-play", Science 362, 2018 -- Paramètres AlphaZero
