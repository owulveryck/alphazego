# Interfaces Go pour le reseau de neurones

Specification des interfaces dans `board/interfaces.go` pour le MCTS d'AlphaZeGo.

## Interface State

`State` est l'interface centrale du framework. Elle represente un etat dans un probleme de decision sequentiel a un ou plusieurs agents.

```go
type State interface {
    // CurrentPlayer retourne le joueur dont c'est le tour d'agir.
    CurrentPlayer() PlayerID
    // PreviousPlayer retourne le joueur qui a effectue le coup menant a cet etat.
    PreviousPlayer() PlayerID
    // Evaluate retourne l'issue du probleme : NoPlayer (en cours), DrawResult, ou le PlayerID gagnant.
    Evaluate() PlayerID
    // PossibleMoves retourne tous les etats atteignables depuis l'etat courant.
    PossibleMoves() []State
    // ID retourne un identifiant unique pour cet etat.
    ID() string
    // LastMove retourne le coup (position) qui a ete joue pour atteindre cet etat.
    LastMove() uint8
}
```

`PreviousPlayer()` permet au moteur MCTS de savoir qui a joue le dernier coup sans connaitre la logique de tour (2 joueurs en alternance, N joueurs en round-robin, etc.).

## Interface Evaluator

L'`Evaluator` est le point d'entree entre le MCTS et le reseau de neurones. Il est defini dans `board/interfaces.go`.

```go
// Evaluator fournit une evaluation par reseau de neurones d'une position de jeu.
// Il est utilise par le MCTS pour remplacer les rollouts aleatoires (value)
// et guider l'exploration (policy).
type Evaluator interface {
    // Evaluate prend un etat de jeu et retourne :
    //   - policy : probabilite a priori pour chaque coup legal,
    //     dans le meme ordre que state.PossibleMoves()
    //   - value : estimation de victoire pour le joueur courant, dans [-1, 1]
    //
    // La somme des elements de policy doit etre egale a 1.
    // Les coups illegaux ne doivent pas apparaitre dans policy.
    Evaluate(state State) (policy []float64, value float64)
}
```

### Contrat

- `policy` a exactement `len(state.PossibleMoves())` elements, dans le meme ordre
- `sum(policy) = 1.0` (distribution de probabilites normalisee)
- `value ∈ [-1, 1]` du point de vue de `state.CurrentPlayer()`
- L'implementation doit etre **thread-safe** si le MCTS est parallelise

### Exemple d'implementation (evaluation aleatoire, pour tests)

```go
type RandomEvaluator struct{}

func (r *RandomEvaluator) Evaluate(state board.State) ([]float64, float64) {
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

`Tensorizable` est implementee par les etats de jeu qui savent se convertir en tenseur pour le reseau de neurones. Definie dans `board/interfaces.go`.

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
    //   Plan 0 : positions du joueur courant (binaire)
    //   Plan 1 : positions de l'adversaire (binaire)
    //   Plan 2 : indicateur du joueur courant (constant)
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
    current := t.CurrentPlayer()
    opponent := 3 - current

    for i := 0; i < 9; i++ {
        if t.board[i] == current {
            features[i] = 1.0         // Plan 0 : joueur courant
        }
        if t.board[i] == opponent {
            features[9+i] = 1.0       // Plan 1 : adversaire
        }
    }

    // Plan 2 : indicateur du joueur courant
    val := float32(0.0)
    if current == board.Player1 {
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
move := bestState.LastMove()
```

### MCTS avec Evaluator (style AlphaZero)

```go
m := mcts.NewAlphaMCTS(evaluator, 1.5)
bestState := m.RunMCTS(currentState, 800)
move := bestState.LastMove()
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
