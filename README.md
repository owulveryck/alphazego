# AlphaZeGo

[![Go](https://github.com/owulveryck/alphazego/actions/workflows/go.yml/badge.svg)](https://github.com/owulveryck/alphazego/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/owulveryck/alphazego.svg)](https://pkg.go.dev/github.com/owulveryck/alphazego)

Implementation d'AlphaZero en Go, from scratch.

## Principe

Le projet s'articule autour de trois couches d'abstraction :

```
decision.State           -- probleme de decision sequentiel generique
    board.Boarder        -- specialisation pour les jeux de plateau
        tictactoe        -- implementation concrete (morpion)
```

### Decision generique (`decision`)

Le package `decision` definit [`State`](https://pkg.go.dev/github.com/owulveryck/alphazego/decision#State), une interface pour tout probleme de decision sequentiel a un ou plusieurs acteurs : jeux, negociations, planification, diagnostic, etc.

```go
type State interface {
    CurrentActor() ActorID
    PreviousActor() ActorID
    Evaluate() ActorID
    PossibleMoves() []State
    ID() string
}
```

### MCTS generique (`mcts`)

Le moteur [MCTS](https://pkg.go.dev/github.com/owulveryck/alphazego/mcts) travaille uniquement avec `decision.State`. Il ne connait ni les regles du jeu, ni le domaine. Deux modes :

- **MCTS pur** (`NewMCTS`) : exploration par rollouts aleatoires + UCB1
- **AlphaZero** (`NewAlphaMCTS`) : guide par un [`Evaluator`](https://pkg.go.dev/github.com/owulveryck/alphazego/mcts#Evaluator) qui fournit policy et value

L'`Evaluator` est le point d'injection de la connaissance du domaine. Il peut etre un reseau de neurones (ONNX), un rollout, une heuristique, un modele de langage, ou toute autre source d'evaluation.

### Jeux de plateau (`decision/board`)

Le package [`board`](https://pkg.go.dev/github.com/owulveryck/alphazego/decision/board) specialise `State` pour les jeux de plateau avec l'interface `Boarder` (= `State` + `ActionRecorder`) et `Tensorizable` pour la conversion en tenseur.

### Morpion (`decision/board/tictactoe`)

Le [morpion](https://pkg.go.dev/github.com/owulveryck/alphazego/decision/board/tictactoe) est l'implementation de reference. Il sert a valider le moteur MCTS sur un jeu simple avant d'attaquer des domaines plus complexes.

## Utilisation

```go
// MCTS pur
m := mcts.NewMCTS()
best := m.RunMCTS(state, 1000)

// Avec un evaluateur (style AlphaZero)
m := mcts.NewAlphaMCTS(evaluator, 1.5)
best := m.RunMCTS(state, 800)
```

```bash
go run main.go  # jouer au morpion contre l'IA
```

## Structure

```
decision/              Probleme de decision generique (State, ActorID)
decision/board/        Abstractions plateau (Boarder, ActionRecorder, Tensorizable)
decision/board/tictactoe/  Morpion (implementation de reference)
mcts/                  Moteur MCTS (UCB1, PUCT, Evaluator)
```

## Documentation

- [Framework generique](docs/explanation/framework-generique.md) -- pourquoi State est plus qu'un jeu
- [Implementer un jeu](docs/how-to/implementer-un-jeu.md) -- guide pas a pas
- [Implementer un Evaluator](docs/how-to/implementer-evaluator.md) -- connecter un reseau de neurones
- [Reference des interfaces](docs/reference/interfaces-evaluator.md) -- contrats et specifications
- [Tutoriel morpion](docs/tutorials/morpion-pas-a-pas.md) -- construire le morpion de zero
