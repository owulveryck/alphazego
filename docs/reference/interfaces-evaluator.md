# Interfaces Go pour le réseau de neurones

Spécification des interfaces dans `decision/state.go`, `decision/board/board.go` et `mcts/evaluator.go` pour le MCTS d'AlphaZeGo.

## Interface State

`State` est l'interface centrale du framework. Elle représente un état dans un problème de décision séquentiel à un ou plusieurs acteurs. Définie dans `decision/state.go`.

```go
type State interface {
    // CurrentActor retourne l'acteur dont c'est le tour d'agir.
    CurrentActor() ActorID
    // PreviousActor retourne l'acteur qui a effectué l'action menant à cet état.
    PreviousActor() ActorID
    // Evaluate retourne l'issue du problème : Undecided (en cours), Stalemate, ou l'ActorID gagnant.
    Evaluate() ActorID
    // PossibleMoves retourne tous les états atteignables depuis l'état courant.
    PossibleMoves() []State
    // ID retourne un identifiant unique pour cet état.
    ID() string
}
```

`PreviousActor()` permet au moteur MCTS de savoir qui a effectué la dernière action sans connaître la logique de tour (2 acteurs en alternance, N acteurs en round-robin, etc.).

## Interface Boarder

`Boarder` combine `decision.State` et `ActionRecorder` pour représenter un état de jeu de plateau complet. Définie dans `decision/board/board.go`.

```go
type Boarder interface {
    decision.State
    ActionRecorder
}
```

## Interface ActionRecorder

`ActionRecorder` est une interface complémentaire à `State`, définie dans `decision/board/board.go`. Elle n'est pas utilisée par le moteur MCTS mais par l'appelant pour extraire l'action choisie.

```go
type ActionRecorder interface {
    // LastAction retourne l'action qui a été effectuée pour atteindre cet état.
    LastAction() int
}
```

Le type de retour est `int` (pas `uint8`) pour supporter des espaces d'action supérieurs à 256 (Go 19x19 : 362, échecs : ~4672).

Utilisation typique :

```go
bestState := m.RunMCTS(currentState, 1000)
move := bestState.(board.ActionRecorder).LastAction()
```

## Interface Evaluator

L'`Evaluator` est le point d'entrée entre le MCTS et le réseau de neurones. Il est défini dans `mcts/evaluator.go` (package `mcts`).

```go
// Evaluator fournit une évaluation d'une position.
// Il est utilisé par le MCTS pour remplacer les rollouts aléatoires (values)
// et guider l'exploration (policy).
type Evaluator interface {
    // Evaluate prend un état et retourne :
    //   - policy : probabilité a priori pour chaque action légale,
    //     dans le même ordre que state.PossibleMoves()
    //   - values : estimation de victoire par acteur, dans [-1, 1]
    //
    // La somme des éléments de policy doit être égale à 1.
    // Les actions illégales ne doivent pas apparaître dans policy.
    Evaluate(state decision.State) (policy []float64, values map[decision.ActorID]float64)
}
```

### Contrat

- `policy` a exactement `len(state.PossibleMoves())` éléments, dans le même ordre
- `sum(policy) = 1.0` (distribution de probabilités normalisée)
- `values[actorID] ∈ [-1, 1]` pour chaque acteur du problème
- La map doit contenir au minimum `CurrentActor()` et `PreviousActor()`
- L'implémentation doit être **thread-safe** si le MCTS est parallélisé

### Exemple d'implémentation (évaluation uniforme, pour tests)

```go
type RandomEvaluator struct{}

func (r *RandomEvaluator) Evaluate(state decision.State) ([]float64, map[decision.ActorID]float64) {
    moves := state.PossibleMoves()
    n := len(moves)
    policy := make([]float64, n)
    for i := range policy {
        policy[i] = 1.0 / float64(n) // distribution uniforme
    }
    values := map[decision.ActorID]float64{
        state.CurrentActor():  0.0,
        state.PreviousActor(): 0.0,
    }
    return policy, values
}
```

## Interface Tensorizable

`Tensorizable` est implémentée par les états de jeu qui savent se convertir en tenseur pour le réseau de neurones. Définie dans `decision/board/board.go`.

```go
// Tensorizable est implémenté par les états de jeu qui peuvent être convertis
// en tenseur pour l'évaluation par un réseau de neurones.
type Tensorizable interface {
    // Features retourne l'état du jeu sous forme de tenseur aplati.
    // Le format attendu est [C * H * W] en row-major order,
    // où C est le nombre de canaux (plans de features),
    // H la hauteur et W la largeur du plateau.
    //
    // Exemple pour le morpion : [3 * 3 * 3] = 27 float32
    //   Plan 0 : positions de l'acteur courant (binaire)
    //   Plan 1 : positions de l'adversaire (binaire)
    //   Plan 2 : indicateur de l'acteur courant (constant)
    Features() []float32

    // FeatureShape retourne les dimensions du tenseur [C, H, W].
    FeatureShape() [3]int

    // ActionSize retourne le nombre total d'actions possibles dans le jeu
    // (pas seulement les actions légales dans l'état courant).
    // Morpion : 9, Go 19x19 : 362 (361 + passe).
    ActionSize() int
}
```

### Contrat

- `len(Features())` == `C * H * W` (cohérent avec `FeatureShape()`)
- `Features()` ne modifie pas l'état
- `ActionSize()` est une constante du jeu, pas de l'état courant

### Exemple pour le morpion

```go
func (t *TicTacToe) Features() []float32 {
    features := make([]float32, 3*3*3) // [3][3][3]
    current := t.CurrentActor()
    opponent := 3 - current

    for i := 0; i < 9; i++ {
        if t.board[i] == current {
            features[i] = 1.0         // Plan 0 : acteur courant
        }
        if t.board[i] == opponent {
            features[9+i] = 1.0       // Plan 1 : adversaire
        }
    }

    // Plan 2 : indicateur de l'acteur courant
    val := float32(0.0)
    if current == tictactoe.Cross {
        val = 1.0
    }
    for i := 18; i < 27; i++ {
        features[i] = val
    }

    return features
}

func (t *TicTacToe) FeatureShape() [3]int {
    return [3]int{3, 3, 3} // 3 canaux, 3x3
}

func (t *TicTacToe) ActionSize() int {
    return 9
}
```

## Utilisation du MCTS

### MCTS pur

```go
m := mcts.NewMCTS()
bestState := m.RunMCTS(currentState, 1000)
move := bestState.(board.ActionRecorder).LastAction()
```

### MCTS avec Evaluator (style AlphaZero)

```go
m := mcts.NewAlphaMCTS(evaluator, 1.5)
bestState := m.RunMCTS(currentState, 800)
move := bestState.(board.ActionRecorder).LastAction()
```

En mode AlphaZero, la sélection utilise PUCT (avec les priors du policy network) au lieu de UCB1, et les values par acteur du réseau remplacent les rollouts aléatoires. Les détails internes (nœuds, PUCT, backpropagation) sont encapsulés dans le package `mcts` et ne sont pas exposés.

## Choix d'implémentation du réseau en Go

| Option | Avantages | Inconvénients |
|---|---|---|
| **ONNX Runtime** (`onnxruntime_go`) | Performant, modèle entraîné en Python | Dépendance C, cross-compilation complexe |
| **Gorgonia** | Natif Go, pas de CGo | Moins mature, API en évolution |
| **gRPC vers Python** | Découplage total, écosystème Python | Latence réseau, déploiement plus complexe |
| **TensorFlow Go** | Officiel Google | Bindings Go peu maintenus |

### Recommandation

Pour ce projet : **ONNX Runtime**.

1. Entrainer le réseau en Python (PyTorch)
2. Exporter en ONNX : `torch.onnx.export(model, ...)`
3. Charger en Go avec `onnxruntime_go`
4. Implémenter `Evaluator` autour du runtime ONNX

L'`Evaluator` encapsule les détails du runtime. Le MCTS ne connaît que l'interface.

## Références

- Package `github.com/yalue/onnxruntime_go` -- Bindings Go pour ONNX Runtime
- Package `gorgonia.org/gorgonia` -- Framework de deep learning natif Go
- ONNX format specification -- https://onnx.ai/
