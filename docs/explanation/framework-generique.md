# Un framework générique de décision, pas un moteur de jeu

## L'intuition

Le MCTS est souvent présenté comme un algorithme pour les jeux de plateau. C'est historiquement vrai -- il a été inventé pour le Go. Mais le mécanisme sous-jacent est bien plus général.

Ce que le MCTS fait réellement :

1. **Explorer un arbre de décisions** possibles
2. **Estimer la qualité** de chaque chemin (par simulation ou évaluation)
3. **Équilibrer** entre approfondir les chemins prometteurs et en essayer de nouveaux

Ce mécanisme fonctionne pour **tout problème** où :

- On prend des **décisions séquentielles** (une après l'autre)
- Chaque décision mène à un **état** identifiable
- On peut **énumérer les choix** possibles à chaque étape
- On peut **évaluer** (même approximativement) la qualité d'un état

## L'interface State : un contrat minimal

L'interface `decision.State` capture exactement ces quatre propriétés :

```go
type State interface {
    CurrentActor() ActorID    // Quel agent doit agir ?
    PreviousActor() ActorID   // Quel agent a effectué la dernière action ?
    Evaluate() ActorID        // L'état est-il terminal ? Quelle issue ?
    PossibleMoves() []State   // Quels états sont atteignables ?
    ID() string               // Identifiant unique de cet état
}
```

Les noms sont génériques (`Actor`, `Moves`) et chaque méthode a une signification adaptable à tout domaine :

| Méthode | Jeu de plateau | Négociation | Diagnostic médical | Composition musicale |
|---------|---------------|-------------|-------------------|---------------------|
| `CurrentActor()` | Joueur actif | Partie qui propose | Décideur (patient/médecin) | Compositeur/critique |
| `PreviousActor()` | Joueur précédent | Partie qui vient de proposer | Dernier décideur | Dernier contributeur |
| `Evaluate()` | Victoire/nul/en cours | Accord/blocage/en cours | Guérison/échec/en cours | Pièce terminée/en cours |
| `PossibleMoves()` | Coups légaux | Propositions possibles | Examens/traitements | Notes/accords possibles |
| `ID()` | Position du plateau | État de la négociation | Dossier patient | Partition en cours |

## Le morpion comme implémentation de référence

Le morpion (`decision/board/tictactoe`) n'est pas le framework -- c'est **une** implémentation de `State` :

```
decision/
├── state.go              ← le contrat générique (State, ActorID)
└── board/
    ├── board.go           ← les abstractions plateau (Boarder, ActionRecorder, Tensorizable)
    └── tictactoe/
        └── ttt.go         ← une implémentation concrète pour le morpion
```

Pour résoudre un autre problème, il suffit d'implémenter `State` avec la logique spécifique au domaine. Le moteur MCTS (`mcts/`) fonctionne avec n'importe quelle implémentation.

## Exemples de problèmes modélisables

### Planification de traitement

```
State = état du patient (symptômes, traitements en cours, résultats d'examens)
CurrentActor() = le décideur (alternance médecin/maladie comme adversaire)
PossibleMoves() = examens prescriptibles, traitements disponibles
Evaluate() = guérison, aggravation, ou en cours
ID() = hash du dossier médical courant
```

L'adversaire ici est la maladie : on modélise l'incertitude comme un agent adverse qui "choisit" les réactions du patient. Le MCTS explore les arbres de traitements pour trouver la stratégie la plus robuste.

### Négociation

```
State = état des offres et contre-offres
CurrentActor() = quelle partie négocie
PossibleMoves() = propositions possibles (concéder, exiger, bluffer)
Evaluate() = accord trouvé, rupture, ou en cours
ID() = hash de l'historique des offres
```

Le MCTS simule des scénarios de négociation pour trouver la stratégie qui maximise les chances d'obtenir un bon accord.

### Génération de texte/code

```
State = texte généré jusqu'ici + contexte
CurrentActor() = le générateur (ou alternance générateur/critique)
PossibleMoves() = tokens ou blocs de code possibles
Evaluate() = qualité du texte final (cohérence, correction)
ID() = hash du texte courant
```

C'est l'idée derrière les modèles de raisonnement actuels : au lieu de générer du texte token par token, on explore un arbre de possibilités et on choisit le chemin le plus prometteur.

## De deux agents à N agents

### L'abstraction : `PreviousActor()`

Le moteur MCTS n'a pas besoin de connaître le nombre d'acteurs. Il a besoin de savoir **qui a effectué l'action menant à un état donné**. C'est le rôle de `PreviousActor()` :

```go
type State interface {
    CurrentActor() ActorID    // Qui doit agir maintenant ?
    PreviousActor() ActorID   // Qui a agi pour arriver ici ?
    Evaluate() ActorID        // L'état est-il terminal ?
    PossibleMoves() []State   // Quels états sont atteignables ?
    ID() string               // Identifiant unique
}
```

Chaque implémentation de `State` encapsule sa propre logique de tour :

| Nombre d'acteurs | Logique de `PreviousActor()` |
|-------------------|-------------------------------|
| 2 acteurs (morpion) | `3 - CurrentActor()` |
| 3 acteurs (round-robin) | `(CurrentActor() + 1) % 3 + 1` (le précédent) |
| 1 acteur (planification) | Toujours le même agent |

Le moteur MCTS utilise `PreviousActor()` dans la backpropagation pour créditer les victoires au bon agent, sans aucune arithmétique codée en dur.

### Le chemin MCTS pur : N acteurs immédiatement

La backpropagation fonctionne pour N acteurs : à chaque nœud, elle vérifie si le résultat terminal correspond à l'agent qui a effectué l'action (`result == PreviousActor()`). Si oui, elle crédite une victoire. Sinon, elle ne crédite rien (ou 0.5 pour un match nul).

### Le chemin AlphaZero : deux acteurs pour l'instant

La backpropagation AlphaZero utilise l'alternance de signe (`value = -value`), qui suppose un jeu à **somme nulle à deux acteurs**. Pour généraliser à N acteurs, il faudrait que l'évaluateur retourne un vecteur de N valeurs (une par acteur) au lieu d'un scalaire. C'est une évolution future.

### La contrainte sur les constantes

`Evaluate()` retourne directement un `ActorID` : le gagnant si le problème est résolu, `NoActor` (0) si il est en cours, ou `DrawResult` (-1) en cas de match nul. Il n'y a pas de type `Result` séparé. Pour un problème à 3+ acteurs, il suffit d'utiliser des `ActorID` distincts (3, 4, ...) ; `DrawResult` (-1) ne peut pas entrer en collision.

### Cas d'usage selon le nombre d'agents

- **1 agent** : planification, composition, optimisation. L'unique agent explore un arbre de décisions sans adversaire.
- **2 agents** (cas de référence) : jeux (morpion, Go, échecs), négociations bilatérales, vérification adversariale.
- **N agents** : négociations multilatérales, jeux à plusieurs joueurs, simulations multi-acteurs.

## L'Evaluator comme "oracle du domaine"

L'interface `Evaluator` abstrait la manière dont on estime la qualité d'un état :

```go
type Evaluator interface {
    Evaluate(state State) (policy []float64, value float64)
}
```

En MCTS pur, l'estimation se fait par simulation aléatoire (rollout). Avec AlphaZero, un réseau de neurones sert d'oracle. Mais l'évaluateur peut être **n'importe quoi** :

- Un modèle statistique
- Une heuristique experte
- Un modèle de langage
- Un ensemble de règles métier
- Même un humain dans la boucle

L'Evaluator est le point d'injection de la **connaissance du domaine**. Le MCTS fournit le mécanisme de recherche ; l'Evaluator fournit l'intuition.

## Résumé

```
Framework générique          Implémentation concrète
─────────────────            ─────────────────────
decision.State               tictactoe.TicTacToe
mcts.Evaluator               rolloutEvaluator, ONNXEvaluator, ...
board.Tensorizable           tictactoe.Features(), ...
mcts.MCTS                    (moteur, agnostique au domaine)
```

L'idée clé : **le MCTS ne sait rien du morpion**. Il sait explorer un arbre, équilibrer exploration et exploitation, et propager des résultats. Le domaine est entièrement encapsulé dans l'implémentation de `State` et d'`Evaluator`.
