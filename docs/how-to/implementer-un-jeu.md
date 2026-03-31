# Comment implementer un nouveau jeu pour le MCTS

Ce guide explique comment implementer l'interface `board.State` pour connecter un jeu (ou tout probleme de decision sequentiel) au moteur MCTS d'AlphaZeGo.

## Prerequis

- Connaissance de base de Go (structs, interfaces, slices)
- Avoir lu la [documentation du framework generique](../explanation/framework-generique.md)

## 1. Definir le struct d'etat

Votre struct doit contenir toutes les informations necessaires pour decrire une position du jeu :

```go
type MonJeu struct {
    plateau    [N]uint8       // representation du plateau
    joueur     board.PlayerID // joueur dont c'est le tour
    lastMove   uint8          // coup qui a mene a cet etat
}
```

**Important** : le champ `lastMove` est necessaire pour implementer `LastMove()`.

## 2. Implementer `board.State`

L'interface `State` comporte 6 methodes :

```go
type State interface {
    CurrentPlayer() PlayerID
    PreviousPlayer() PlayerID
    Evaluate() PlayerID
    PossibleMoves() []State
    ID() string
    LastMove() uint8
}
```

### `CurrentPlayer()` et `PreviousPlayer()`

Pour un jeu a deux joueurs en alternance :

```go
func (j *MonJeu) CurrentPlayer() board.PlayerID {
    return j.joueur
}

func (j *MonJeu) PreviousPlayer() board.PlayerID {
    return 3 - j.joueur // alternance : 1↔2
}
```

Pour N joueurs en round-robin, adaptez la logique (modulo, etc.).

### `Evaluate()`

Retourne le `PlayerID` du gagnant, `board.DrawResult` pour un nul, ou `board.NoPlayer` si la partie est en cours :

```go
func (j *MonJeu) Evaluate() board.PlayerID {
    // Verifier les conditions de victoire
    if victoire(board.Player1) {
        return board.Player1
    }
    if victoire(board.Player2) {
        return board.Player2
    }
    // Verifier le match nul (plateau plein, etc.)
    if plateauPlein() {
        return board.DrawResult
    }
    return board.NoPlayer // partie en cours
}
```

### `PossibleMoves()`

Retourne un slice de `board.State`, chaque element representant un etat apres un coup legal.

**Regle critique : ne jamais muter le recepteur.** Chaque etat fils doit etre une copie independante.

```go
func (j *MonJeu) PossibleMoves() []board.State {
    var moves []board.State
    for i := 0; i < N; i++ {
        if j.plateau[i] == 0 { // case vide
            // Copier le plateau
            nouveau := &MonJeu{
                joueur:   3 - j.joueur,
                lastMove: uint8(i),
            }
            copy(nouveau.plateau[:], j.plateau[:])
            nouveau.plateau[i] = uint8(j.joueur)
            moves = append(moves, nouveau)
        }
    }
    return moves
}
```

### `ID()`

Retourne un identifiant unique pour cet etat. Deux etats identiques (meme plateau, meme joueur) doivent retourner le meme ID :

```go
func (j *MonJeu) ID() string {
    id := make([]byte, N+1)
    copy(id, j.plateau[:])
    id[N] = byte(j.joueur)
    return string(id)
}
```

### `LastMove()`

Retourne le coup qui a produit cet etat :

```go
func (j *MonJeu) LastMove() uint8 {
    return j.lastMove
}
```

## 3. Ajouter `Play()` pour l'interaction humaine

`Play()` n'est pas dans l'interface `State` mais est necessaire pour permettre a un humain de jouer :

```go
func (j *MonJeu) Play(position uint8) error {
    if position >= N {
        return fmt.Errorf("position %d hors limites (0-%d)", position, N-1)
    }
    if j.plateau[position] != 0 {
        return fmt.Errorf("position %d deja occupee", position)
    }
    if j.Evaluate() != board.NoPlayer {
        return fmt.Errorf("la partie est terminee")
    }
    j.plateau[position] = uint8(j.joueur)
    j.lastMove = position
    j.joueur = 3 - j.joueur
    return nil
}
```

## 4. Connecter au MCTS

```go
m := mcts.NewMCTS()
game := NewMonJeu() // etat initial

// Trouver le meilleur coup avec 1000 iterations
bestState := m.RunMCTS(game, 1000)
move := bestState.LastMove()

// Appliquer le coup
game.Play(move)
```

## 5. Optionnel : `Tensorizable` pour AlphaZero

Si vous voulez utiliser un reseau de neurones, implementez `board.Tensorizable` :

```go
func (j *MonJeu) Features() []float32    { /* tenseur aplati [C*H*W] */ }
func (j *MonJeu) FeatureShape() [3]int   { return [3]int{C, H, W} }
func (j *MonJeu) ActionSize() int        { return N }
```

Voir la [reference des interfaces](../reference/interfaces-evaluator.md) pour les details du contrat.

## 6. Optionnel : `Evaluator` pour une evaluation custom

Implementez `mcts.Evaluator` pour remplacer les rollouts aleatoires :

```go
type MonEvaluateur struct{}

func (e *MonEvaluateur) Evaluate(state board.State) ([]float64, float64) {
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
| Oublier `lastMove` dans `PossibleMoves()` | `LastMove()` retourne 0 pour tous les coups | Affecter `lastMove: uint8(i)` dans chaque etat fils |
| Collisions d'`ID()` | Le MCTS confond des etats differents | Inclure le joueur courant dans l'ID |
| `Evaluate()` ne detecte pas le nul | Le MCTS boucle indefiniment | Verifier le plateau plein / absence de coups |
| `Play()` sans validation | Etat incoherent silencieux | Valider bornes, occupation, partie terminee |
