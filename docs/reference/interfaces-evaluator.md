# Interfaces Go pour le reseau de neurones

Specification des interfaces dans `decision/state.go`, `decision/board/board.go` et `mcts/evaluator.go` pour le MCTS d'AlphaZeGo.

## Interface State

`State` est l'interface centrale du framework. Elle represente un etat dans un probleme de decision sequentiel a un ou plusieurs acteurs. Definie dans `decision/state.go`.

```go
type State interface {
    // CurrentActor retourne l'acteur dont c'est le tour d'agir.
    CurrentActor() ActorID
    // PreviousActor retourne l'acteur qui a effectue l'action menant a cet etat.
    PreviousActor() ActorID
    // Evaluate retourne l'issue du probleme : NoActor (en cours), DrawResult, ou l'ActorID gagnant.
    Evaluate() ActorID
    // PossibleMoves retourne tous les etats atteignables depuis l'etat courant.
    PossibleMoves() []State
    // ID retourne un identifiant unique pour cet etat.
    ID() string
}
```

`PreviousActor()` permet au moteur MCTS de savoir qui a effectue la derniere action sans connaitre la logique de tour (2 acteurs en alternance, N acteurs en round-robin, etc.).

## Interface Boarder

`Boarder` combine `decision.State` et `ActionRecorder` pour representer un etat de jeu de plateau complet. Definie dans `decision/board/board.go`.

```go
type Boarder interface {
    decision.State
    ActionRecorder
}
```

## Interface ActionRecorder

`ActionRecorder` est une interface complementaire a `State`, definie dans `decision/board/board.go`. Elle n'est pas utilisee par le moteur MCTS mais par l'appelant pour extraire l'action choisie.

```go
type ActionRecorder interface {
    // LastAction retourne l'action qui a ete effectuee pour atteindre cet etat.
    LastAction() int
}
```

Le type de retour est `int` (pas `uint8`) pour supporter des espaces d'action superieurs a 256 (Go 19x19 : 362, echecs : ~4672).

Utilisation typique :

```go
bestState := m.RunMCTS(currentState, 1000)
move := bestState.(board.ActionRecorder).LastAction()
```

## Interface Evaluator

L'`Evaluator` est le point d'entree entre le MCTS et le reseau de neurones. Il est defini dans `mcts/evaluator.go` (package `mcts`).

```go
// Evaluator fournit une evaluation d'une position.
// Il est utilise par le MCTS pour remplacer les rollouts aleatoires (value)
// et guider l'exploration (policy).
type Evaluator interface {
    // Evaluate prend un etat et retourne :
    //   - policy : probabilite a priori pour chaque action legale,
    //     dans le meme ordre que state.PossibleMoves()
    //   - value : estimation de victoire pour l'acteur courant, dans [-1, 1]
    //
    // La somme des elements de policy doit etre egale a 1.
    // Les actions illegales ne doivent pas apparaitre dans policy.
    Evaluate(state decision.State) (policy []float64, value float64)
}
```

### Contrat

- `policy` a exactement `len(state.PossibleMoves())` elements, dans le meme ordre
- `sum(policy) = 1.0` (distribution de probabilites normalisee)
- `value ∈ [-1, 1]` du point de vue de `state.CurrentActor()`
- L'implementation doit etre **thread-safe** si le MCTS est parallelise

### Exemple d'implementation (evaluation aleatoire, pour tests)

```go
type RandomEvaluator struct{}

func (r *RandomEvaluator) Evaluate(state decision.State) ([]float64, float64) {
    moves := state.PossibleMoves()
    n := len(moves)
    policy := make([]float64, n)
    for i := range policy {
        policy[i] = 1.0 / float64(n) // distribution uniforme
    }
    value := 0.0 // position neutre
    return policy, value
}
```

## Interface Tensorizable

`Tensorizable` est implementee par les etats de jeu qui savent se convertir en tenseur pour le reseau de neurones. Definie dans `decision/board/board.go`.

```go
// Tensorizable est implemente par les etats de jeu qui peuvent etre convertis
// en tenseur pour l'evaluation par un reseau de neurones.
type Tensorizable interface {
    // Features retourne l'etat du jeu sous forme de tenseur aplati.
    // Le format attendu est [C * H * W] en row-major order,
    // ou C est le nombre de canaux (plans de features),
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
    // (pas seulement les actions legales dans l'etat courant).
    // Morpion : 9, Go 19x19 : 362 (361 + passe).
    ActionSize() int
}
```

### Contrat

- `len(Features())` == `C * H * W` (coherent avec `FeatureShape()`)
- `Features()` ne modifie pas l'etat
- `ActionSize()` est une constante du jeu, pas de l'etat courant

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
    if current == decision.Actor1 {
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

En mode AlphaZero, la selection utilise PUCT (avec les priors du policy network) au lieu de UCB1, et la value du reseau remplace les rollouts aleatoires. Les details internes (noeuds, PUCT, backpropagation) sont encapsules dans le package `mcts` et ne sont pas exposes.

## Choix d'implementation du reseau en Go

| Option | Avantages | Inconvenients |
|---|---|---|
| **ONNX Runtime** (`onnxruntime_go`) | Performant, modele entraine en Python | Dependance C, cross-compilation complexe |
| **Gorgonia** | Natif Go, pas de CGo | Moins mature, API en evolution |
| **gRPC vers Python** | Decouplage total, ecosysteme Python | Latence reseau, deploiement plus complexe |
| **TensorFlow Go** | Officiel Google | Bindings Go peu maintenus |

### Recommandation

Pour ce projet : **ONNX Runtime**.

1. Entrainer le reseau en Python (PyTorch)
2. Exporter en ONNX : `torch.onnx.export(model, ...)`
3. Charger en Go avec `onnxruntime_go`
4. Implementer `Evaluator` autour du runtime ONNX

L'`Evaluator` encapsule les details du runtime. Le MCTS ne connait que l'interface.

## References

- Package `github.com/yalue/onnxruntime_go` -- Bindings Go pour ONNX Runtime
- Package `gorgonia.org/gorgonia` -- Framework de deep learning natif Go
- ONNX format specification -- https://onnx.ai/
