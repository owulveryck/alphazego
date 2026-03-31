# Un framework generique de decision, pas un moteur de jeu

## L'intuition

Le MCTS est souvent presente comme un algorithme pour les jeux de plateau. C'est historiquement vrai -- il a ete invente pour le Go. Mais le mecanisme sous-jacent est bien plus general.

Ce que le MCTS fait reellement :

1. **Explorer un arbre de decisions** possibles
2. **Estimer la qualite** de chaque chemin (par simulation ou evaluation)
3. **Equilibrer** entre approfondir les chemins prometteurs et en essayer de nouveaux

Ce mecanisme fonctionne pour **tout probleme** ou :

- On prend des **decisions sequentielles** (une apres l'autre)
- Chaque decision mene a un **etat** identifiable
- On peut **enumerer les choix** possibles a chaque etape
- On peut **evaluer** (meme approximativement) la qualite d'un etat

## L'interface State : un contrat minimal

L'interface `board.State` capture exactement ces quatre proprietes :

```go
type State interface {
    CurrentPlayer() PlayerID    // Quel agent doit agir ?
    PreviousPlayer() PlayerID   // Quel agent a effectue le dernier coup ?
    Evaluate() PlayerID         // L'etat est-il terminal ? Quelle issue ?
    PossibleMoves() []State     // Quels etats sont atteignables ?
    ID() string                 // Identifiant unique de cet etat
    LastMove() uint8            // Quel coup a mene a cet etat ?
}
```

Malgre les noms herites des jeux (`Player`, `Moves`), chaque methode a une signification generique :

| Methode | Jeu de plateau | Negociation | Diagnostic medical | Composition musicale |
|---------|---------------|-------------|-------------------|---------------------|
| `CurrentPlayer()` | Joueur actif | Partie qui propose | Decideur (patient/medecin) | Compositeur/critique |
| `PreviousPlayer()` | Joueur precedent | Partie qui vient de proposer | Dernier decideur | Dernier contributeur |
| `Evaluate()` | Victoire/nul/en cours | Accord/blocage/en cours | Guerison/echec/en cours | Piece terminee/en cours |
| `PossibleMoves()` | Coups legaux | Propositions possibles | Examens/traitements | Notes/accords possibles |
| `ID()` | Position du plateau | Etat de la negociation | Dossier patient | Partition en cours |
| `LastMove()` | Coup joue | Derniere proposition | Dernier examen/traitement | Derniere note ajoutee |

## Le morpion comme implementation de reference

Le morpion (`board/tictactoe`) n'est pas le framework -- c'est **une** implementation de `State` :

```
board/
├── interfaces.go        ← le contrat generique (State, Evaluator, Tensorizable)
└── tictactoe/
    └── ttt.go           ← une implementation concrete pour le morpion
```

Pour resoudre un autre probleme, il suffit d'implementer `State` avec la logique specifique au domaine. Le moteur MCTS (`mcts/`) fonctionne avec n'importe quelle implementation.

## Exemples de problemes modelisables

### Planification de traitement

```
State = etat du patient (symptomes, traitements en cours, resultats d'examens)
CurrentPlayer() = le decideur (alternance medecin/maladie comme adversaire)
PossibleMoves() = examens prescriptibles, traitements disponibles
Evaluate() = guerison, aggravation, ou en cours
ID() = hash du dossier medical courant
```

L'adversaire ici est la maladie : on modelise l'incertitude comme un agent adverse qui "choisit" les reactions du patient. Le MCTS explore les arbres de traitements pour trouver la strategie la plus robuste.

### Negociation

```
State = etat des offres et contre-offres
CurrentPlayer() = quelle partie negocie
PossibleMoves() = propositions possibles (conceder, exiger, bluffer)
Evaluate() = accord trouve, rupture, ou en cours
ID() = hash de l'historique des offres
```

Le MCTS simule des scenarios de negociation pour trouver la strategie qui maximise les chances d'obtenir un bon accord.

### Generation de texte/code

```
State = texte genere jusqu'ici + contexte
CurrentPlayer() = le generateur (ou alternance generateur/critique)
PossibleMoves() = tokens ou blocs de code possibles
Evaluate() = qualite du texte final (coherence, correction)
ID() = hash du texte courant
```

C'est l'idee derriere les modeles de raisonnement actuels : au lieu de generer du texte token par token, on explore un arbre de possibilites et on choisit le chemin le plus prometteur.

## De deux agents a N agents

### L'abstraction : `PreviousPlayer()`

Le moteur MCTS n'a pas besoin de connaitre le nombre de joueurs. Il a besoin de savoir **qui a effectue le coup menant a un etat donne**. C'est le role de `PreviousPlayer()` :

```go
type State interface {
    CurrentPlayer() PlayerID    // Qui doit agir maintenant ?
    PreviousPlayer() PlayerID   // Qui a agi pour arriver ici ?
    Evaluate() PlayerID         // L'etat est-il terminal ?
    PossibleMoves() []State     // Quels etats sont atteignables ?
    ID() string                 // Identifiant unique
    LastMove() uint8            // Quel coup a mene ici ?
}
```

Chaque implementation de `State` encapsule sa propre logique de tour :

| Nombre de joueurs | Logique de `PreviousPlayer()` |
|-------------------|-------------------------------|
| 2 joueurs (morpion) | `3 - CurrentPlayer()` |
| 3 joueurs (round-robin) | `(CurrentPlayer() + 1) % 3 + 1` (le precedent) |
| 1 joueur (planification) | Toujours le meme agent |

Le moteur MCTS utilise `PreviousPlayer()` dans la backpropagation pour crediter les victoires au bon agent, sans aucune arithmetique codee en dur.

### Le chemin MCTS pur : N joueurs immediatement

La backpropagation fonctionne pour N joueurs : a chaque noeud, elle verifie si le resultat terminal correspond a l'agent qui a joue le coup (`result == PreviousPlayer()`). Si oui, elle credite une victoire. Sinon, elle ne credite rien (ou 0.5 pour un match nul).

### Le chemin AlphaZero : deux joueurs pour l'instant

La backpropagation AlphaZero utilise l'alternance de signe (`value = -value`), qui suppose un jeu a **somme nulle a deux joueurs**. Pour generaliser a N joueurs, il faudrait que l'evaluateur retourne un vecteur de N valeurs (une par joueur) au lieu d'un scalaire. C'est une evolution future.

### La contrainte sur les constantes

`Evaluate()` retourne directement un `PlayerID` : le gagnant si la partie est finie, `NoPlayer` (0) si elle est en cours, ou `DrawResult` (-1) en cas de match nul. Il n'y a pas de type `Result` separe. Pour un jeu a 3+ joueurs, il suffit d'utiliser des `PlayerID` distincts (3, 4, ...) ; `DrawResult` (-1) ne peut pas entrer en collision.

### Cas d'usage selon le nombre d'agents

- **1 agent** : planification, composition, optimisation. L'unique agent explore un arbre de decisions sans adversaire.
- **2 agents** (cas de reference) : jeux (morpion, Go, echecs), negociations bilaterales, verification adversariale.
- **N agents** : negociations multilaterales, jeux a plusieurs joueurs, simulations multi-acteurs.

## L'Evaluator comme "oracle du domaine"

L'interface `Evaluator` abstraire la maniere dont on estime la qualite d'un etat :

```go
type Evaluator interface {
    Evaluate(state State) (policy []float64, value float64)
}
```

En MCTS pur, l'estimation se fait par simulation aleatoire (rollout). Avec AlphaZero, un reseau de neurones sert d'oracle. Mais l'evaluateur peut etre **n'importe quoi** :

- Un modele statistique
- Une heuristique experte
- Un modele de langage
- Un ensemble de regles metier
- Meme un humain dans la boucle

L'Evaluator est le point d'injection de la **connaissance du domaine**. Le MCTS fournit le mecanisme de recherche ; l'Evaluator fournit l'intuition.

## Resume

```
Framework generique          Implementation concrete
─────────────────            ─────────────────────
board.State                  tictactoe.TicTacToe
mcts.Evaluator               rolloutEvaluator, ONNXEvaluator, ...
board.Tensorizable           tictactoe.Features(), ...
mcts.MCTS                    (moteur, agnostique au domaine)
```

L'idee cle : **le MCTS ne sait rien du morpion**. Il sait explorer un arbre, equilibrer exploration et exploitation, et propager des resultats. Le domaine est entierement encapsule dans l'implementation de `State` et d'`Evaluator`.
