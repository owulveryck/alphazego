# AlphaGo et les reseaux de neurones

## Le probleme fondamental

Le Go a un facteur de branchement d'environ **250** (nombre de coups legaux moyens) et des parties de **200+ coups**. L'arbre de jeu complet contient environ 10^170 positions -- plus que le nombre d'atomes dans l'univers observable.

Le MCTS pur ne peut pas gerer cette complexite. Ses rollouts aleatoires sont du bruit, et son exploration uniforme gaspille du temps sur des coups absurdes qu'un joueur humain eliminerait instantanement.

La percee d'AlphaGo : utiliser des **reseaux de neurones convolutifs** pour donner au MCTS une "intuition" humaine (voire surhumaine).

## Les deux cerveaux d'AlphaGo

AlphaGo utilise deux fonctions apprise par reseau de neurones :

```
                    Etat du plateau (tenseur)
                         │
                   ┌─────┴─────┐
                   │  Tronc     │
                   │  commun    │  Blocs residuels (ResNet)
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
           de probabilites v ∈ [-1, 1]
           sur les coups
```

### Policy network (p) -- "Quels coups sont prometteurs ?"

Le policy network prend un etat de plateau et produit une **distribution de probabilites** sur tous les coups legaux. Pour le Go : un vecteur de 362 valeurs (361 intersections + passer).

Exemple simplifie pour le morpion :

```
Position :        Policy :
 X | O | .        p = [0, 0, 0.15,
 . | X | .             0.05, 0, 0.20,
 . | . | O             0.10, 0.05, 0]
                        ↑
                   probabilite de jouer chaque case
```

Ce vecteur remplace l'exploration uniforme dans MCTS. Au lieu d'essayer tous les coups avec la meme priorite, l'algorithme explore d'abord les coups a forte probabilite.

**Role dans MCTS** : guide la **selection** via la formule PUCT (voir [de-mcts-a-alphazero.md](de-mcts-a-alphazero.md)).

### Value network (v) -- "Qui va gagner ?"

Le value network prend un etat de plateau et produit un **scalaire** entre -1 et +1, estimant la probabilite de victoire du joueur courant.

- `v = +1` : le joueur courant gagne certainement
- `v = 0` : position equilibree
- `v = -1` : le joueur courant perd certainement

**Role dans MCTS** : remplace le **rollout aleatoire**. Au lieu de jouer des coups au hasard jusqu'a la fin, on demande au reseau "qui gagne depuis cette position ?".

## Pourquoi la convolution ?

Le plateau de Go (ou de morpion) est une **grille 2D**, exactement comme une image. Les reseaux convolutifs excellent sur ce type de donnees grace a deux proprietes :

### Localite spatiale

Les motifs au Go sont locaux : un "oeil", un "atari", une "echelle" sont des configurations de pierres voisines. Une convolution 3x3 capture exactement ces relations de voisinage.

```
Filtre 3x3 detectant        Application sur le plateau :
un motif d'atari :
                             . . . . .
  0  1  0                   . X O . .
  1 -1  0       ──►         . X ● . .  ← le filtre "voit" l'atari
  0  0  0                   . . . . .
```

### Invariance par translation

Un motif qui fonctionne dans un coin du plateau fonctionne aussi au centre. La convolution partage ses poids sur toute la grille, donc le reseau apprend le motif **une seule fois** et le reconnait partout.

### Architecture convolutive detaillee

Voir [reference/architecture-reseau.md](../reference/architecture-reseau.md) pour les details techniques (tenseurs d'entree, blocs residuels, dimensions).

## Representation du plateau en tenseur

Le plateau n'est pas passe au reseau sous forme de grille brute. Il est encode en **plans de features** (feature planes), chacun de taille H x W :

Pour AlphaGo Zero (Go 19x19) :

| Plan | Contenu | Nombre |
|---|---|---|
| Pierres du joueur courant (t, t-1, ..., t-7) | Positions binaires | 8 |
| Pierres de l'adversaire (t, t-1, ..., t-7) | Positions binaires | 8 |
| Couleur du joueur courant | Plan constant (0 ou 1) | 1 |
| **Total** | | **17** |

Le tenseur d'entree est donc de dimension `[17][19][19]`.

Pour le morpion, on pourrait utiliser :

| Plan | Contenu | Nombre |
|---|---|---|
| Positions du joueur courant | Binaire 3x3 | 1 |
| Positions de l'adversaire | Binaire 3x3 | 1 |
| Joueur courant | Plan constant | 1 |
| **Total** | | **3** |

Tenseur d'entree : `[3][3][3]` -- suffisant pour un reseau minimaliste.

## L'evolution : AlphaGo, AlphaGo Zero, AlphaZero

### AlphaGo (2016) -- Silver et al., Nature

- **Deux reseaux separes** : policy (13 couches CNN) et value (13 couches CNN)
- Policy pre-entraine par **apprentissage supervise** sur 30 millions de parties humaines
- Affine par **reinforcement learning** (self-play)
- Simulation = **melange** rollout aleatoire + value network : `v_final = λ * v_network + (1-λ) * v_rollout`
- A battu Lee Sedol 4-1

### AlphaGo Zero (2017) -- Silver et al., Nature

- **Un seul reseau a deux tetes** (policy + value partagent le tronc)
- **Aucune donnee humaine** : entraine uniquement par self-play (tabula rasa)
- **Pas de rollout** : la value network seule remplace la simulation
- Architecture **ResNet** (blocs residuels) au lieu de CNN simple
- A battu AlphaGo 100-0

### AlphaZero (2018) -- Silver et al., Science

- Meme architecture qu'AlphaGo Zero
- **Generalise** a trois jeux : Go, echecs, shogi
- Aucune connaissance specifique au jeu sauf les regles
- A battu Stockfish (echecs) et Elmo (shogi) en partant de zero

### MuZero (2020) -- Schrittwieser et al., Nature

- N'a meme plus besoin des **regles du jeu**
- Apprend un **modele interne** de l'environnement (dynamics network)
- Trois reseaux : representation, dynamics, prediction
- Fonctionne aussi sur les jeux Atari (pas seulement les jeux de plateau)

```
AlphaGo          AlphaGo Zero       AlphaZero          MuZero
(2016)           (2017)             (2018)             (2020)
    │                │                  │                  │
    │ Supervise +    │ Self-play        │ Generalise       │ Apprend les
    │ rollouts       │ sans rollout     │ a 3 jeux         │ regles aussi
    │ 2 reseaux      │ 1 reseau         │                  │ 3 reseaux
    │                │ ResNet           │                  │
    └──────► ameliore ──────► generalise ──────► abstrait ──►
```

## Le cycle d'entrainement

AlphaZero apprend par **self-play iteratif** :

```
  ┌─────────────────────────────────────────┐
  │                                         │
  v                                         │
Reseau (v0)  ──►  Self-play avec MCTS  ──►  Donnees d'entrainement
                   (MCTS utilise le            │
                    reseau courant)             │
                                               v
                                         Entrainer le reseau
                                         sur les parties jouees
                                               │
                                               v
                                         Reseau (v1) ──► ...
  │                                         │
  └─────────────────────────────────────────┘
```

A chaque partie de self-play :
1. MCTS (guide par le reseau courant) choisit les coups
2. On stocke pour chaque position : `(etat, politique_MCTS, resultat_final)`
3. On entraine le reseau pour que :
   - sa **policy** predise la politique MCTS (cross-entropy loss)
   - sa **value** predise le resultat final (MSE loss)

La politique MCTS est **meilleure** que la politique brute du reseau (car elle fait de la recherche), donc le reseau apprend a imiter une version amelioree de lui-meme. C'est un cercle vertueux.

## References

- Silver et al., "Mastering the game of Go with deep neural networks and tree search", Nature 529, 2016 -- AlphaGo
- Silver et al., "Mastering the game of Go without human knowledge", Nature 550, 2017 -- AlphaGo Zero
- Silver et al., "A general reinforcement learning algorithm that masters chess, shogi, and Go through self-play", Science 362, 2018 -- AlphaZero
- Schrittwieser et al., "Mastering Atari, Go, Chess and Shogi by Planning with a Learned Model", Nature 588, 2020 -- MuZero
- He et al., "Deep Residual Learning for Image Recognition", CVPR 2016 -- ResNet
