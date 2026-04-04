# ActorID : type et conventions

## Type

### ActorID

```go
type ActorID int
```

Type distinct basé sur `int`. Identifie un décideur dans un problème de décision séquentiel. Sert aussi de résultat : `Evaluate()` retourne directement le `ActorID` du gagnant.

Il n'y a pas de type `Result` séparé : le résultat EST l'identifiant du gagnant.

## Constantes

| Constante | Valeur | Description |
|-----------|--------|-------------|
| `Undecided` | `0` | Aucun joueur. Jeu en cours, ou case vide. |
| `Stalemate` | `-1` | Match nul. La partie est terminée sans vainqueur. |
| `Actor1` | `1` | Premier joueur (X au morpion) |
| `Actor2` | `2` | Second joueur (O au morpion) |

### Pourquoi Stalemate = -1 ?

Avec une valeur négative, `Stalemate` ne peut jamais entrer en collision avec un identifiant de joueur (qui est toujours positif). Un jeu à 3, 4 ou N joueurs peut utiliser les identifiants 1, 2, 3, 4... sans risque de confusion avec le match nul.

## Méthodes de State retournant ActorID

```go
type State interface {
    CurrentActor() ActorID      // le joueur dont c'est le tour d'agir
    PreviousActor() ActorID     // le joueur qui a effectué le dernier coup
    Evaluate() ActorID           // le gagnant, Undecided, ou Stalemate
    // ...
}
```

### CurrentActor

Retourne le joueur qui doit prendre la prochaine décision. À l'état initial, c'est le premier joueur.

### PreviousActor

Retourne le joueur qui a effectué le coup menant à cet état. Permet au moteur MCTS de créditer les victoires au bon joueur sans connaître la logique de tour.

Pour l'état initial (aucun coup joué), le comportement est défini par l'implémentation.

### Evaluate

Retourne :
- `Undecided` (0) si le jeu est en cours
- `Stalemate` (-1) en cas de match nul
- un `ActorID` positif si ce joueur a gagné

## Utilisation dans le MCTS

### Backpropagation discrète (MCTS pur, N joueurs)

```go
playerWhoMovedHere := n.state.PreviousActor()
if result == playerWhoMovedHere {    // le gagnant == celui qui a joué ici
    n.wins += 1
} else if result == decision.Stalemate {
    n.wins += 0.5
}
```

### Valeur terminale (AlphaZero, 2 joueurs)

```go
playerWhoMovedHere := s.PreviousActor()
if result == playerWhoMovedHere {
    return -1.0   // défaite pour le joueur courant
}
if result == decision.Stalemate {
    return 0.0
}
return 1.0        // victoire pour le joueur courant
```

## Exemples d'implémentation

### 2 joueurs (morpion)

```go
func (t *TicTacToe) CurrentActor() decision.ActorID {
    return t.PlayerTurn
}

func (t *TicTacToe) PreviousActor() decision.ActorID {
    return 3 - t.PlayerTurn   // alternance stricte : 1↔2
}
```

### 3 joueurs (round-robin)

```go
func (s *ThreePlayerGame) CurrentActor() decision.ActorID {
    return s.current   // valeurs 10, 11, 12
}

func (s *ThreePlayerGame) PreviousActor() decision.ActorID {
    return s.previous  // le joueur qui vient de jouer
}
```

### 1 joueur (planification)

```go
func (s *Planner) CurrentActor() decision.ActorID  { return 5 }
func (s *Planner) PreviousActor() decision.ActorID { return 5 }
```
