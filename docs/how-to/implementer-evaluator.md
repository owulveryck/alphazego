# Implementer un Evaluator

Guide pour creer une implementation de l'interface `board.Evaluator`, qui permet d'utiliser le MCTS en mode AlphaZero.

## Prerequis

- Comprendre le role de l'`Evaluator` : [reference/interfaces-evaluator.md](../reference/interfaces-evaluator.md)
- Comprendre les differences MCTS pur / AlphaZero : [explanation/de-mcts-a-alphazero.md](../explanation/de-mcts-a-alphazero.md)

## L'interface

```go
// dans board/interfaces.go
type Evaluator interface {
    Evaluate(state State) (policy []float64, value float64)
}
```

L'`Evaluator` est appele par `RunMCTS` a chaque expansion de noeud. Il recoit un etat de jeu et doit retourner :

| Retour | Description | Contraintes |
|--------|-------------|-------------|
| `policy` | Probabilite a priori de chaque coup legal | Meme ordre que `state.PossibleMoves()`, somme = 1.0 |
| `value` | Estimation de victoire pour le joueur courant | Dans [-1, 1]. +1 = victoire certaine, -1 = defaite |

## Etape 1 : Definir la structure

```go
package evaluator

import "github.com/owulveryck/alphazego/board"

type MonEvaluator struct {
    // champs internes : modele charge, session, etc.
}
```

## Etape 2 : Implementer Evaluate

Le corps de `Evaluate` suit toujours le meme schema :

```go
func (e *MonEvaluator) Evaluate(state board.State) ([]float64, float64) {
    // 1. Obtenir les coups legaux
    moves := state.PossibleMoves()
    n := len(moves)
    if n == 0 {
        return nil, 0.0
    }

    // 2. Calculer la policy et la value
    //    (specifique a chaque implementation)
    policy := make([]float64, n)
    var value float64
    // ... remplir policy et value ...

    // 3. Normaliser la policy (somme = 1)
    sum := 0.0
    for _, p := range policy {
        sum += p
    }
    for i := range policy {
        policy[i] /= sum
    }

    return policy, value
}
```

## Exemples d'implementation

### Evaluateur uniforme (pour tests)

Policy uniforme et value neutre. Utile pour verifier que le chemin AlphaZero fonctionne sans signal.

```go
type UniformEvaluator struct{}

func (u *UniformEvaluator) Evaluate(state board.State) ([]float64, float64) {
    moves := state.PossibleMoves()
    n := len(moves)
    if n == 0 {
        return nil, 0.0
    }
    policy := make([]float64, n)
    for i := range policy {
        policy[i] = 1.0 / float64(n)
    }
    return policy, 0.0 // value neutre
}
```

### Evaluateur a rollout (pour tests tactiques)

Policy uniforme, mais value estimee par un rollout aleatoire. Cela fournit un vrai signal pour guider l'arbre. Cet evaluateur est utilise dans les tests d'integration (`mcts/puct_test.go`).

```go
type RolloutEvaluator struct{}

func (r *RolloutEvaluator) Evaluate(state board.State) ([]float64, float64) {
    moves := state.PossibleMoves()
    n := len(moves)
    if n == 0 {
        return nil, 0.0
    }

    // Policy uniforme
    policy := make([]float64, n)
    for i := range policy {
        policy[i] = 1.0 / float64(n)
    }

    // Value par rollout aleatoire
    currentState := state
    for currentState.Evaluate() == board.NoPlayer {
        possibleMoves := currentState.PossibleMoves()
        currentState = possibleMoves[rand.Intn(len(possibleMoves))]
    }
    result := currentState.Evaluate()
    current := state.CurrentPlayer()
    if result == current {
        return policy, 1.0
    }
    if result == board.DrawResult {
        return policy, 0.0
    }
    return policy, -1.0
}
```

### Evaluateur ONNX (pour un vrai reseau)

Utilise un modele ONNX exporte depuis PyTorch. L'etat doit implementer `board.Tensorizable` pour la conversion en tenseur.

```go
type ONNXEvaluator struct {
    session    *ort.Session
    actionSize int
}

func NewONNXEvaluator(modelPath string, actionSize int) (*ONNXEvaluator, error) {
    session, err := ort.NewSession(modelPath)
    if err != nil {
        return nil, err
    }
    return &ONNXEvaluator{session: session, actionSize: actionSize}, nil
}

func (e *ONNXEvaluator) Evaluate(state board.State) ([]float64, float64) {
    moves := state.PossibleMoves()
    n := len(moves)
    if n == 0 {
        return nil, 0.0
    }

    // 1. Convertir l'etat en tenseur via Tensorizable
    t := state.(board.Tensorizable)
    features := t.Features()
    shape := t.FeatureShape()

    // 2. Appeler le reseau
    input := ort.NewTensor(features, []int64{1, int64(shape[0]), int64(shape[1]), int64(shape[2])})
    outputs, _ := e.session.Run([]ort.Tensor{input})

    // 3. Extraire la policy brute (sur tout l'espace d'action)
    rawPolicy := outputs[0].Float64s() // taille = actionSize

    // 4. Masquer les coups illegaux et normaliser
    //    La policy retournee ne doit contenir que les coups legaux,
    //    dans le meme ordre que state.PossibleMoves().
    policy := filterLegalMoves(state, rawPolicy)

    // 5. Extraire la value
    value := outputs[1].Float64s()[0] // deja dans [-1, 1] grace au tanh

    return policy, value
}

// filterLegalMoves extrait les probabilites des coups legaux
// depuis le vecteur brut du reseau et les normalise.
func filterLegalMoves(state board.State, rawPolicy []float64) []float64 {
    moves := state.PossibleMoves()
    policy := make([]float64, len(moves))
    sum := 0.0

    for i, move := range moves {
        // Identifier l'index de l'action dans l'espace complet
        // En utilisant Playable pour obtenir le numero du coup
        action := state.(board.Playable).GetMoveFromState(move)
        p := math.Exp(rawPolicy[action]) // softmax sur les logits
        policy[i] = p
        sum += p
    }

    // Normaliser
    for i := range policy {
        policy[i] /= sum
    }
    return policy
}
```

## Utilisation avec le MCTS

```go
// Creer l'evaluateur
eval := &MonEvaluator{...}

// Creer le MCTS AlphaZero avec cpuct = 1.5
m := mcts.NewAlphaMCTS(eval, 1.5)

// Jouer un coup
bestState := m.RunMCTS(currentState, 800)
move := board.State(currentState).(board.Playable).GetMoveFromState(bestState)
```

## Points de vigilance

### Ordre de la policy

`policy[i]` doit correspondre a `state.PossibleMoves()[i]`. Si le reseau produit un vecteur sur tout l'espace d'action (ex: 9 cases pour le morpion), il faut filtrer et reordonner pour ne garder que les coups legaux.

### Perspective de la value

`value` est du point de vue de `state.CurrentPlayer()`. Si le joueur courant est en position de gagner, `value` doit etre positif. Le MCTS se charge de l'inversion de signe lors de la backpropagation.

### Thread-safety

Si le MCTS est parallelise, l'`Evaluator` doit etre thread-safe. Typiquement, proteger l'appel au reseau avec un `sync.Mutex` ou utiliser un pool de sessions.

### Performance

L'`Evaluate` est appele une fois par expansion de noeud (pas une fois par iteration). Pour le morpion avec 800 iterations, il est appele au plus ~800 fois. Pour le Go avec 1600 iterations, idem. Le batching (evaluer plusieurs positions en un seul appel reseau) est une optimisation possible pour le MCTS parallele.

## Verification

Tester l'evaluateur independamment du MCTS :

```go
func TestMonEvaluator(t *testing.T) {
    eval := &MonEvaluator{...}
    state := tictactoe.NewTicTacToe()

    policy, value := eval.Evaluate(state)

    // Verifier le nombre d'elements
    if len(policy) != len(state.PossibleMoves()) {
        t.Errorf("policy length mismatch")
    }

    // Verifier la normalisation
    sum := 0.0
    for _, p := range policy {
        sum += p
    }
    if math.Abs(sum-1.0) > 1e-6 {
        t.Errorf("policy sum = %f, want 1.0", sum)
    }

    // Verifier la borne de la value
    if value < -1.0 || value > 1.0 {
        t.Errorf("value = %f, want [-1, 1]", value)
    }
}
```

Puis tester avec le MCTS :

```go
func TestMonEvaluator_AvecMCTS(t *testing.T) {
    eval := &MonEvaluator{...}
    m := mcts.NewAlphaMCTS(eval, 1.5)
    state := tictactoe.NewTicTacToe()

    result := m.RunMCTS(state, 100)
    if result == nil {
        t.Fatal("expected non-nil result")
    }
}
```
