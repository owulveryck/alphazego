# PlayerID : type et conventions

## Type

### PlayerID

```go
type PlayerID int
```

Type distinct base sur `int`. Identifie un decideur dans un probleme de decision sequentiel. Sert aussi de resultat : `Evaluate()` retourne directement le `PlayerID` du gagnant.

Il n'y a pas de type `Result` separe : le resultat EST l'identifiant du gagnant.

## Constantes

| Constante | Valeur | Description |
|-----------|--------|-------------|
| `NoPlayer` | `0` | Aucun joueur. Jeu en cours, ou case vide. |
| `DrawResult` | `-1` | Match nul. La partie est terminee sans vainqueur. |
| `Player1` | `1` | Premier joueur (X au morpion) |
| `Player2` | `2` | Second joueur (O au morpion) |

### Pourquoi DrawResult = -1 ?

Avec une valeur negative, `DrawResult` ne peut jamais entrer en collision avec un identifiant de joueur (qui est toujours positif). Un jeu a 3, 4 ou N joueurs peut utiliser les identifiants 1, 2, 3, 4... sans risque de confusion avec le match nul.

## Methodes de State retournant PlayerID

```go
type State interface {
    CurrentPlayer() PlayerID      // le joueur dont c'est le tour d'agir
    PreviousPlayer() PlayerID     // le joueur qui a effectue le dernier coup
    Evaluate() PlayerID           // le gagnant, NoPlayer, ou DrawResult
    // ...
}
```

### CurrentPlayer

Retourne le joueur qui doit prendre la prochaine decision. A l'etat initial, c'est le premier joueur.

### PreviousPlayer

Retourne le joueur qui a effectue le coup menant a cet etat. Permet au moteur MCTS de crediter les victoires au bon joueur sans connaitre la logique de tour.

Pour l'etat initial (aucun coup joue), le comportement est defini par l'implementation.

### Evaluate

Retourne :
- `NoPlayer` (0) si le jeu est en cours
- `DrawResult` (-1) en cas de match nul
- un `PlayerID` positif si ce joueur a gagne

## Utilisation dans le MCTS

### Backpropagation discrete (MCTS pur, N joueurs)

```go
playerWhoMovedHere := n.state.PreviousPlayer()
if result == playerWhoMovedHere {    // le gagnant == celui qui a joue ici
    n.wins += 1
} else if result == board.DrawResult {
    n.wins += 0.5
}
```

### Valeur terminale (AlphaZero, 2 joueurs)

```go
playerWhoMovedHere := s.PreviousPlayer()
if result == playerWhoMovedHere {
    return -1.0   // defaite pour le joueur courant
}
if result == board.DrawResult {
    return 0.0
}
return 1.0        // victoire pour le joueur courant
```

## Exemples d'implementation

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
