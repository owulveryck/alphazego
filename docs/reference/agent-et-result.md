# Agent et Result : types et conventions

## Types

### Agent

```go
type Agent = uint8
```

Alias de `uint8`. Identifie un decideur dans un probleme de decision sequentiel.

### Result

```go
type Result = uint8
```

Alias de `uint8`. Represente l'issue de l'evaluation d'un etat.

### Lien entre Agent et Result

La convention fondamentale du framework est :

> **`Result(a)` signifie que l'agent `a` a gagne.**

Cette convention est formalisee par la fonction `PlayerWins` :

```go
func PlayerWins(a Agent) Result {
    return Result(a)
}
```

## Constantes

### Agents predefinies (2 joueurs)

| Constante | Valeur | Description |
|-----------|--------|-------------|
| `Player1` | `1` | Premier agent (X au morpion) |
| `Player2` | `2` | Second agent (O au morpion) |

### Resultats

| Constante | Valeur | Description |
|-----------|--------|-------------|
| `GameOn` | `0` | Probleme en cours, pas d'issue terminale |
| `Player1Wins` | `1` (`= Player1`) | L'agent 1 a atteint son objectif |
| `Player2Wins` | `2` (`= Player2`) | L'agent 2 a atteint son objectif |
| `Draw` | `3` | Egalite, aucun agent n'a gagne |
| `Stalemat` | `4` | Blocage sans vainqueur |

### Valeurs reservees

Les valeurs 0 a 4 sont reservees par les constantes ci-dessus. Pour un jeu a N joueurs (N > 2), les identifiants d'agents doivent eviter ces valeurs pour ne pas entrer en collision avec `Draw` (3) ou `Stalemat` (4). Utiliser des valeurs >= 5.

## Methodes de State retournant Agent

```go
type State interface {
    CurrentPlayer() Agent      // l'agent dont c'est le tour d'agir
    PreviousPlayer() Agent     // l'agent qui a effectue le dernier coup
    // ...
}
```

### CurrentPlayer

Retourne l'agent qui doit prendre la prochaine decision. A l'etat initial, c'est le premier joueur.

### PreviousPlayer

Retourne l'agent qui a effectue le coup menant a cet etat. Permet au moteur MCTS de crediter les victoires au bon agent sans connaitre la logique de tour.

Pour l'etat initial (aucun coup joue), le comportement est defini par l'implementation.

## Utilisation dans le MCTS

### Backpropagation discrete (MCTS pur, N joueurs)

```go
playerWhoMovedHere := n.state.PreviousPlayer()
if result == playerWhoMovedHere {    // Result == Agent → cet agent a gagne
    n.wins += 1
} else if result == board.Draw {
    n.wins += 0.5
}
```

### Valeur terminale (AlphaZero, 2 joueurs)

```go
playerWhoMovedHere := s.PreviousPlayer()
if result == playerWhoMovedHere {
    return -1.0   // defaite pour le joueur courant
}
if result == board.Draw {
    return 0.0
}
return 1.0        // victoire pour le joueur courant
```

## Exemples d'implementation

### 2 joueurs (morpion)

```go
func (t *TicTacToe) CurrentPlayer() board.Agent {
    return t.PlayerTurn
}

func (t *TicTacToe) PreviousPlayer() board.Agent {
    return 3 - t.PlayerTurn   // alternance stricte : 1↔2
}
```

### 3 joueurs (round-robin)

```go
func (s *ThreePlayerGame) CurrentPlayer() board.Agent {
    return s.current   // valeurs 10, 11, 12 (evitent la collision avec Draw=3)
}

func (s *ThreePlayerGame) PreviousPlayer() board.Agent {
    return s.previous  // le joueur qui vient de jouer
}
```

### 1 joueur (planification)

```go
func (s *Planner) CurrentPlayer() board.Agent  { return 5 }
func (s *Planner) PreviousPlayer() board.Agent { return 5 }
```
