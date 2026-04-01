# AlphaGo et les réseaux de neurones

## Le problème fondamental

Le Go a un facteur de branchement d'environ **250** (nombre de coups légaux moyens) et des parties de **200+ coups**. L'arbre de jeu complet contient environ 10^170 positions -- plus que le nombre d'atomes dans l'univers observable.

Le MCTS pur ne peut pas gérer cette complexité. Ses rollouts aléatoires sont du bruit, et son exploration uniforme gaspille du temps sur des coups absurdes qu'un joueur humain éliminerait instantanément.

La percée d'AlphaGo : utiliser des **réseaux de neurones convolutifs** pour donner au MCTS une "intuition" humaine (voire surhumaine).

## Les deux cerveaux d'AlphaGo

AlphaGo utilise deux fonctions apprises par réseau de neurones :

```
                    Etat du plateau (tenseur)
                         │
                   ┌─────┴─────┐
                   │  Tronc     │
                   │  commun    │  Blocs résiduels (ResNet)
                   │  convolutif│  convolutions 3x3
                   └─────┬─────┘
                    ┌────┴────┐
                    │         │
              ┌─────┴───┐ ┌──┴─────┐
              │ Tete     │ │ Tete   │
              │ Policy   │ │ Value  │
              │ (p)      │ │ (v)    │
              └─────┬───┘ └──┬─────┘
                    │         │
           Distribution    Scalaire
           de probabilités v ∈ [-1, 1]
           sur les coups
```

### Policy network (p) -- "Quels coups sont prometteurs ?"

Le policy network prend un état de plateau et produit une **distribution de probabilités** sur tous les coups légaux. Pour le Go : un vecteur de 362 valeurs (361 intersections + passer).

Exemple simplifié pour le morpion :

```
Position :        Policy :
 X | O | .        p = [0, 0, 0.15,
 . | X | .             0.05, 0, 0.20,
 . | . | O             0.10, 0.05, 0]
                        ↑
                   probabilité de jouer chaque case
```

Ce vecteur remplace l'exploration uniforme dans MCTS. Au lieu d'essayer tous les coups avec la même priorité, l'algorithme explore d'abord les coups à forte probabilité.

**Rôle dans MCTS** : guide la **sélection** via la formule PUCT (voir [de-mcts-a-alphazero.md](de-mcts-a-alphazero.md)).

### Value network (v) -- "Qui va gagner ?"

Le value network prend un état de plateau et produit un **scalaire** entre -1 et +1, estimant la probabilité de victoire du joueur courant.

- `v = +1` : le joueur courant gagne certainement
- `v = 0` : position équilibrée
- `v = -1` : le joueur courant perd certainement

**Rôle dans MCTS** : remplace le **rollout aléatoire**. Au lieu de jouer des coups au hasard jusqu'à la fin, on demande au réseau "qui gagne depuis cette position ?".

## Pourquoi la convolution ?

Le plateau de Go (ou de morpion) est une **grille 2D**, exactement comme une image. Les réseaux convolutifs excellent sur ce type de données grâce à deux propriétés :

### Localité spatiale

Les motifs au Go sont locaux : un "oeil", un "atari", une "echelle" sont des configurations de pierres voisines. Une convolution 3x3 capture exactement ces relations de voisinage.

```
Filtre 3x3 détectant        Application sur le plateau :
un motif d'atari :
                             . . . . .
  0  1  0                   . X O . .
  1 -1  0       ──►         . X ● . .  ← le filtre "voit" l'atari
  0  0  0                   . . . . .
```

### Invariance par translation

Un motif qui fonctionne dans un coin du plateau fonctionne aussi au centre. La convolution partage ses poids sur toute la grille, donc le réseau apprend le motif **une seule fois** et le reconnaît partout.

### Architecture convolutive détaillée

Voir [référence/architecture-réseau.md](../référence/architecture-réseau.md) pour les détails techniques (tenseurs d'entrée, blocs résiduels, dimensions).

## Représentation du plateau en tenseur

Le plateau n'est pas passé au réseau sous forme de grille brute. Il est encodé en **plans de features** (feature planes), chacun de taille H x W :

Pour AlphaGo Zero (Go 19x19) :

| Plan | Contenu | Nombre |
|---|---|---|
| Pierres du joueur courant (t, t-1, ..., t-7) | Positions binaires | 8 |
| Pierres de l'adversaire (t, t-1, ..., t-7) | Positions binaires | 8 |
| Couleur du joueur courant | Plan constant (0 ou 1) | 1 |
| **Total** | | **17** |

Le tenseur d'entrée est donc de dimension `[17][19][19]`.

Pour le morpion, on pourrait utiliser :

| Plan | Contenu | Nombre |
|---|---|---|
| Positions du joueur courant | Binaire 3x3 | 1 |
| Positions de l'adversaire | Binaire 3x3 | 1 |
| Joueur courant | Plan constant | 1 |
| **Total** | | **3** |

Tenseur d'entrée : `[3][3][3]` -- suffisant pour un réseau minimaliste.

## L'évolution : AlphaGo, AlphaGo Zero, AlphaZero

### AlphaGo (2016) -- Silver et al., Nature

- **Deux réseaux séparés** : policy (13 couches CNN) et value (13 couches CNN)
- Policy pré-entraîné par **apprentissage supervisé** sur 30 millions de parties humaines
- Affiné par **reinforcement learning** (self-play)
- Simulation = **mélange** rollout aléatoire + value network : `v_final = λ * v_network + (1-λ) * v_rollout`
- A battu Lee Sedol 4-1

### AlphaGo Zero (2017) -- Silver et al., Nature

- **Un seul réseau à deux têtes** (policy + value partagent le tronc)
- **Aucune donnée humaine** : entraîné uniquement par self-play (tabula rasa)
- **Pas de rollout** : la value network seule remplace la simulation
- Architecture **ResNet** (blocs résiduels) au lieu de CNN simple
- A battu AlphaGo 100-0

### AlphaZero (2018) -- Silver et al., Science

- Même architecture qu'AlphaGo Zero
- **Généralisé** à trois jeux : Go, échecs, shogi
- Aucune connaissance spécifique au jeu sauf les règles
- A battu Stockfish (échecs) et Elmo (shogi) en partant de zéro

### MuZero (2020) -- Schrittwieser et al., Nature

- N'a même plus besoin des **règles du jeu**
- Apprend un **modèle interne** de l'environnement (dynamics network)
- Trois réseaux : représentation, dynamics, prédiction
- Fonctionne aussi sur les jeux Atari (pas seulement les jeux de plateau)

```
AlphaGo          AlphaGo Zero       AlphaZero          MuZero
(2016)           (2017)             (2018)             (2020)
    │                │                  │                  │
    │ Supervisé +    │ Self-play        │ Généralisé       │ Apprend les
    │ rollouts       │ sans rollout     │ à 3 jeux         │ règles aussi
    │ 2 réseaux      │ 1 réseau         │                  │ 3 réseaux
    │                │ ResNet           │                  │
    └──────► amélioré ──────► généralisé ──────► abstrait ──►
```

## Le cycle d'entraînement

AlphaZero apprend par **self-play itératif** :

```
  ┌─────────────────────────────────────────┐
  │                                         │
  v                                         │
Réseau (v0)  ──►  Self-play avec MCTS  ──►  Données d'entraînement
                   (MCTS utilise le            │
                    réseau courant)             │
                                               v
                                         Entraîner le réseau
                                         sur les parties jouées
                                               │
                                               v
                                         Réseau (v1) ──► ...
  │                                         │
  └─────────────────────────────────────────┘
```

À chaque partie de self-play :
1. MCTS (guidé par le réseau courant) choisit les coups
2. On stocke pour chaque position : `(état, politique_MCTS, résultat_final)`
3. On entraîne le réseau pour que :
   - sa **policy** prédise la politique MCTS (cross-entropy loss)
   - sa **value** prédise le résultat final (MSE loss)

La politique MCTS est **meilleure** que la politique brute du réseau (car elle fait de la recherche), donc le réseau apprend à imiter une version améliorée de lui-même. C'est un cercle vertueux.

## Références

- Silver et al., "Mastering the game of Go with deep neural networks and tree search", Nature 529, 2016 -- AlphaGo
- Silver et al., "Mastering the game of Go without human knowledge", Nature 550, 2017 -- AlphaGo Zero
- Silver et al., "A general reinforcement learning algorithm that masters chess, shogi, and Go through self-play", Science 362, 2018 -- AlphaZero
- Schrittwieser et al., "Mastering Atari, Go, Chess and Shogi by Planning with a Learned Model", Nature 588, 2020 -- MuZero
- He et al., "Deep Residual Learning for Image Recognition", CVPR 2016 -- ResNet
