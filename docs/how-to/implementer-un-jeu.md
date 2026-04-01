# Comment implementer un nouveau jeu pour le MCTS

Ce guide explique comment implementer l'interface `decision.State` pour connecter un jeu (ou tout probleme de decision sequentiel) au moteur MCTS d'AlphaZeGo.

## Prerequis

- Connaissance de base de Go (structs, interfaces, slices)
- Avoir lu la [documentation du framework generique](../explanation/framework-generique.md)

## 1. Definir le struct d'etat

Votre struct doit contenir toutes les informations necessaires pour decrire une position du jeu :

```go
type MonJeu struct {
    plateau    [N]uint8          // representation du plateau
    acteur     decision.ActorID  // acteur dont c'est le tour
    lastAction int               // action qui a mene a cet etat (pour ActionRecorder)
}
```

**Optionnel** : le champ `lastAction` est necessaire si vous implementez `board.ActionRecorder`.

## 2. Implementer `decision.State`

L'interface `State` comporte 5 methodes :

```go
type State interface {
    CurrentActor() ActorID
    PreviousActor() ActorID
    Evaluate() ActorID
    PossibleMoves() []State
    ID() string
}
```

### `CurrentActor()` et `PreviousActor()`

Pour un jeu a deux acteurs en alternance :

```go
func (j *MonJeu) CurrentActor() decision.ActorID {
    return j.acteur
}

func (j *MonJeu) PreviousActor() decision.ActorID {
    return 3 - j.acteur // alternance : 1↔2
}
```

Pour N acteurs en round-robin, adaptez la logique (modulo, etc.).

### `Evaluate()`

Retourne l'`ActorID` du gagnant, `decision.DrawResult` pour un nul, ou `decision.NoActor` si la partie est en cours :

```go
func (j *MonJeu) Evaluate() decision.ActorID {
    // Verifier les conditions de victoire
    if victoire(decision.Actor1) {
        return decision.Actor1
    }
    if victoire(decision.Actor2) {
        return decision.Actor2
    }
    // Verifier le match nul (plateau plein, etc.)
    if plateauPlein() {
        return decision.DrawResult
    }
    return decision.NoActor // partie en cours
}
```

### `PossibleMoves()`

Retourne un slice de `decision.State`, chaque element representant un etat apres un coup legal.

**Regle critique : ne jamais muter le recepteur.** Chaque etat fils doit etre une copie independante.

```go
func (j *MonJeu) PossibleMoves() []decision.State {
    var moves []decision.State
    for i := 0; i < N; i++ {
        if j.plateau[i] == 0 { // case vide
            // Copier le plateau
            nouveau := &MonJeu{
                acteur:     3 - j.acteur,
                lastAction: i,
            }
            copy(nouveau.plateau[:], j.plateau[:])
            nouveau.plateau[i] = uint8(j.acteur)
            moves = append(moves, nouveau)
        }
    }
    return moves
}
```

### `ID()`

Retourne un identifiant unique pour cet etat. Deux etats identiques (meme plateau, meme acteur) doivent retourner le meme ID :

```go
func (j *MonJeu) ID() string {
    id := make([]byte, N+1)
    copy(id, j.plateau[:])
    id[N] = byte(j.acteur)
    return string(id)
}
```

## 3. Implementer `ActionRecorder` (optionnel)

Si vous voulez pouvoir extraire l'action choisie par le MCTS, implementez `board.ActionRecorder` :

```go
func (j *MonJeu) LastAction() int {
    return j.lastAction
}
```

L'utilisation se fait via une assertion de type :

```go
bestState := m.RunMCTS(game, 1000)
move := bestState.(board.ActionRecorder).LastAction()
```

## 4. Ajouter `Play()` pour l'interaction humaine

`Play()` n'est pas dans l'interface `State` mais est necessaire pour permettre a un humain de jouer :

```go
func (j *MonJeu) Play(position uint8) error {
    if position >= N {
        return fmt.Errorf("position %d hors limites (0-%d)", position, N-1)
    }
    if j.plateau[position] != 0 {
        return fmt.Errorf("position %d deja occupee", position)
    }
    if j.Evaluate() != decision.NoActor {
        return fmt.Errorf("la partie est terminee")
    }
    j.plateau[position] = uint8(j.acteur)
    j.lastAction = int(position)
    j.acteur = 3 - j.acteur
    return nil
}
```

## 5. Connecter au MCTS

```go
m := mcts.NewMCTS()
game := NewMonJeu() // etat initial

// Trouver le meilleur coup avec 1000 iterations
bestState := m.RunMCTS(game, 1000)
move := bestState.(board.ActionRecorder).LastAction()

// Appliquer le coup
game.Play(uint8(move))
```

## 6. Optionnel : `Tensorizable` pour AlphaZero

Si vous voulez utiliser un reseau de neurones, implementez `board.Tensorizable` :

```go
func (j *MonJeu) Features() []float32    { /* tenseur aplati [C*H*W] */ }
func (j *MonJeu) FeatureShape() [3]int   { return [3]int{C, H, W} }
func (j *MonJeu) ActionSize() int        { return N }
```

Voir la [reference des interfaces](../reference/interfaces-evaluator.md) pour les details du contrat.

## 7. Optionnel : `Evaluator` pour une evaluation custom

Implementez `mcts.Evaluator` pour remplacer les rollouts aleatoires :

```go
type MonEvaluateur struct{}

func (e *MonEvaluateur) Evaluate(state decision.State) ([]float64, float64) {
    moves := state.PossibleMoves()
    policy := make([]float64, len(moves))
    // ... calculer les probabilites pour chaque coup ...
    value := 0.0 // estimation de la position
    return policy, value
}

m := mcts.NewAlphaMCTS(evaluateur, 1.5) // cpuct entre 1.0 et 5.0
```

Voir le [how-to Evaluator](implementer-evaluator.md) pour un guide detaille.

## Erreurs courantes

| Erreur | Consequence | Solution |
|--------|------------|----------|
| Muter le plateau dans `PossibleMoves()` | Tous les etats partagent le meme tableau | Toujours copier le plateau avant de modifier |
| Oublier `lastAction` dans `PossibleMoves()` | `LastAction()` retourne 0 pour tous les coups | Affecter `lastAction: i` dans chaque etat fils |
| Collisions d'`ID()` | Le MCTS confond des etats differents | Inclure l'acteur courant dans l'ID |
| `Evaluate()` ne detecte pas le nul | Le MCTS boucle indefiniment | Verifier le plateau plein / absence de coups |
| `Play()` sans validation | Etat incoherent silencieux | Valider bornes, occupation, partie terminee |
