# PlayerID : type et conventions

## Type

### PlayerID

```go
type PlayerID int
```

Type distinct basé sur `int`. Identifie un décideur dans un problème de décision séquentiel. Sert aussi de résultat : `Evaluate()` retourne directement le `PlayerID` du gagnant.

Il n'y a pas de type `Result` séparé : le résultat EST l'identifiant du gagnant.

## Constantes

| Constante | Valeur | Description |
|-----------|--------|-------------|
| `NoPlayer` | `0` | Aucun joueur. Jeu en cours, ou case vide. |
| `DrawResult` | `-1` | Match nul. La partie est terminée sans vainqueur. |
| `Player1` | `1` | Premier joueur (X au morpion) |
| `Player2` | `2` | Second joueur (O au morpion) |

### Pourquoi DrawResult = -1 ?

Avec une valeur négative, `DrawResult` ne peut jamais entrer en collision avec un identifiant de joueur (qui est toujours positif). Un jeu à 3, 4 ou N joueurs peut utiliser les identifiants 1, 2, 3, 4... sans risque de confusion avec le match nul.

## Méthodes de State retournant PlayerID

```go
type State interface {
    CurrentPlayer() PlayerID      // le joueur dont c'est le tour d'agir
    PreviousPlayer() PlayerID     // le joueur qui a effectué le dernier coup
    Evaluate() PlayerID           // le gagnant, NoPlayer, ou DrawResult
    // ...
}
```

### CurrentPlayer

Retourne le joueur qui doit prendre la prochaine décision. À l'état initial, c'est le premier joueur.

### PreviousPlayer

Retourne le joueur qui a effectué le coup menant à cet état. Permet au moteur MCTS de créditer les victoires au bon joueur sans connaître la logique de tour.

Pour l'état initial (aucun coup joué), le comportement est défini par l'implémentation.

### Evaluate

Retourne :
- `NoPlayer` (0) si le jeu est en cours
- `DrawResult` (-1) en cas de match nul
- un `PlayerID` positif si ce joueur a gagné

## Utilisation dans le MCTS

### Backpropagation discrète (MCTS pur, N joueurs)

```go
playerWhoMovedHere := n.state.PreviousPlayer()
if result == playerWhoMovedHere {    // le gagnant == celui qui a joué ici
    n.wins += 1
} else if result == board.DrawResult {
    n.wins += 0.5
}
```

### Valeur terminale (AlphaZero, 2 joueurs)

```go
playerWhoMovedHere := s.PreviousPlayer()
if result == playerWhoMovedHere {
    return -1.0   // défaite pour le joueur courant
}
if result == board.DrawResult {
    return 0.0
}
return 1.0        // victoire pour le joueur courant
```

## Exemples d'implémentation

### 2 joueurs (morpion)

```go
func (t *TicTacToe) CurrentPlayer() board.PlayerID {
    return t.PlayerTurn
}

func (t *TicTacToe) PreviousPlayer() board.PlayerID {
    return 3 - t.PlayerTurn   // alternance stricte : 1↔2
}
```

### 3 joueurs (round-robin)

```go
func (s *ThreePlayerGame) CurrentPlayer() board.PlayerID {
    return s.current   // valeurs 10, 11, 12
}

func (s *ThreePlayerGame) PreviousPlayer() board.PlayerID {
    return s.previous  // le joueur qui vient de jouer
}
```

### 1 joueur (planification)

```go
func (s *Planner) CurrentPlayer() board.PlayerID  { return 5 }
func (s *Planner) PreviousPlayer() board.PlayerID { return 5 }
```
