# AlphaZeGo

[![Go](https://github.com/owulveryck/alphazego/actions/workflows/go.yml/badge.svg)](https://github.com/owulveryck/alphazego/actions/workflows/go.yml)
[![Go Référence](https://pkg.go.dev/badge/github.com/owulveryck/alphazego.svg)](https://pkg.go.dev/github.com/owulveryck/alphazego)

Implémentation d'AlphaZero en Go, from scratch.

## Principe

Le projet s'articule autour de trois couches d'abstraction :

```
decision.State           -- problème de décision séquentiel générique
    board.Boarder        -- spécialisation pour les jeux de plateau
        tictactoe        -- implémentation concrète (morpion)
```

### Décision générique (`decision`)

Le package `decision` définit [`State`](https://pkg.go.dev/github.com/owulveryck/alphazego/decision#State), une interface pour tout problème de décision séquentiel à un ou plusieurs acteurs : jeux, négociations, planification, diagnostic, etc.

```go
type State interface {
    CurrentActor() ActorID
    PreviousActor() ActorID
    Evaluate() ActorID
    PossibleMoves() []State
    ID() string
}
```

### MCTS générique (`mcts`)

Le moteur [MCTS](https://pkg.go.dev/github.com/owulveryck/alphazego/mcts) travaille uniquement avec `decision.State`. Il ne connaît ni les règles du jeu, ni le domaine. Deux modes :

- **MCTS pur** (`NewMCTS`) : exploration par rollouts aléatoires + UCB1
- **AlphaZero** (`NewAlphaMCTS`) : guidé par un [`Evaluator`](https://pkg.go.dev/github.com/owulveryck/alphazego/mcts#Evaluator) qui fournit policy et value

L'`Evaluator` est le point d'injection de la connaissance du domaine. Il peut être un réseau de neurones (ONNX), un rollout, une heuristique, un modèle de langage, ou toute autre source d'évaluation.

### Jeux de plateau (`decision/board`)

Le package [`board`](https://pkg.go.dev/github.com/owulveryck/alphazego/decision/board) spécialise `State` pour les jeux de plateau avec l'interface `Boarder` (= `State` + `ActionRecorder`) et `Tensorizable` pour la conversion en tenseur.

### Morpion (`decision/board/tictactoe`)

Le [morpion](https://pkg.go.dev/github.com/owulveryck/alphazego/decision/board/tictactoe) est l'implémentation de référence. Il sert à valider le moteur MCTS sur un jeu simple avant d'attaquer des domaines plus complexes.

## Utilisation

```go
// MCTS pur
m := mcts.NewMCTS()
best := m.RunMCTS(state, 1000)

// Avec un évaluateur (style AlphaZero)
m := mcts.NewAlphaMCTS(evaluator, 1.5)
best := m.RunMCTS(state, 800)
```

```bash
go run main.go  # jouer au morpion contre l'IA
```

## Structure

```
decision/              Problème de décision générique (State, ActorID)
decision/board/        Abstractions plateau (Boarder, ActionRecorder, Tensorizable)
decision/board/tictactoe/  Morpion (implémentation de référence)
mcts/                  Moteur MCTS (UCB1, PUCT, Evaluator)
```

## Documentation

- [Framework générique](docs/explanation/framework-générique.md) -- pourquoi State est plus qu'un jeu
- [Implémenter un jeu](docs/how-to/implementer-un-jeu.md) -- guide pas à pas
- [Implémenter un Evaluator](docs/how-to/implementer-evaluator.md) -- connecter un réseau de neurones
- [Référence des interfaces](docs/référence/interfaces-evaluator.md) -- contrats et spécifications
- [Tutoriel morpion](docs/tutorials/morpion-pas-a-pas.md) -- construire le morpion de zéro
