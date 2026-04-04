# Comment implémenter un nouveau jeu pour le MCTS

Ce guide explique comment implémenter l'interface `decision.State` pour connecter un jeu (ou tout problème de décision séquentiel) au moteur MCTS d'AlphaZeGo.

## Prérequis

- Connaissance de base de Go (structs, interfaces, slices)
- Avoir lu la [documentation du framework générique](../explanation/framework-generique.md)

## 1. Définir le struct d'état

Votre struct doit contenir toutes les informations nécessaires pour décrire une position du jeu :

```go
type MonJeu struct {
    plateau    [N]uint8          // représentation du plateau
    acteur     decision.ActorID  // acteur dont c'est le tour
    lastAction int               // action qui a mené à cet état (pour ActionRecorder)
}
```

**Optionnel** : le champ `lastAction` est nécessaire si vous implémentez `board.ActionRecorder`.

## 2. Implémenter `decision.State`

L'interface `State` comporte 5 méthodes :

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

Pour un jeu à deux acteurs en alternance :

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

Retourne l'`ActorID` du gagnant, `decision.Stalemate` pour un nul, ou `decision.Undecided` si la partie est en cours :

```go
func (j *MonJeu) Evaluate() decision.ActorID {
    // Vérifier les conditions de victoire
    if victoire(decision.ActorID(1)) {
        return decision.ActorID(1)
    }
    if victoire(decision.ActorID(2)) {
        return decision.ActorID(2)
    }
    // Vérifier le match nul (plateau plein, etc.)
    if plateauPlein() {
        return decision.Stalemate
    }
    return decision.Undecided // partie en cours
}
```

### `PossibleMoves()`

Retourne un slice de `decision.State`, chaque élément représentant un état après un coup légal.

**Règle critique : ne jamais muter le récepteur.** Chaque état fils doit être une copie indépendante.

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

Retourne un identifiant unique pour cet état. Deux états identiques (même plateau, même acteur) doivent retourner le même ID :

```go
func (j *MonJeu) ID() string {
    id := make([]byte, N+1)
    copy(id, j.plateau[:])
    id[N] = byte(j.acteur)
    return string(id)
}
```

## 3. Implémenter `ActionRecorder` (optionnel)

Si vous voulez pouvoir extraire l'action choisie par le MCTS, implémentez `board.ActionRecorder` :

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

`Play()` n'est pas dans l'interface `State` mais est nécessaire pour permettre à un humain de jouer :

```go
func (j *MonJeu) Play(position uint8) error {
    if position >= N {
        return fmt.Errorf("position %d hors limites (0-%d)", position, N-1)
    }
    if j.plateau[position] != 0 {
        return fmt.Errorf("position %d déjà occupée", position)
    }
    if j.Evaluate() != decision.Undecided {
        return fmt.Errorf("la partie est terminée")
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
game := NewMonJeu() // état initial

// Trouver le meilleur coup avec 1000 itérations
bestState := m.RunMCTS(game, 1000)
move := bestState.(board.ActionRecorder).LastAction()

// Appliquer le coup
game.Play(uint8(move))
```

## 6. Optionnel : `Tensorizable` pour AlphaZero

Si vous voulez utiliser un réseau de neurones, implémentez `board.Tensorizable` :

```go
func (j *MonJeu) Features() []float32    { /* tenseur aplati [C*H*W] */ }
func (j *MonJeu) FeatureShape() [3]int   { return [3]int{C, H, W} }
func (j *MonJeu) ActionSize() int        { return N }
```

Voir la [référence des interfaces](../reference/interfaces-evaluator.md) pour les détails du contrat.

## 7. Optionnel : `Evaluator` pour une évaluation custom

Implémentez `mcts.Evaluator` pour remplacer les rollouts aléatoires :

```go
type MonÉvaluateur struct{}

func (e *MonÉvaluateur) Evaluate(state decision.State) ([]float64, float64) {
    moves := state.PossibleMoves()
    policy := make([]float64, len(moves))
    // ... calculer les probabilités pour chaque coup ...
    value := 0.0 // estimation de la position
    return policy, value
}

m := mcts.NewAlphaMCTS(évaluateur, 1.5) // cpuct entre 1.0 et 5.0
```

Voir le [how-to Evaluator](implementer-evaluator.md) pour un guide détaillé.

## Erreurs courantes

| Erreur | Conséquence | Solution |
|--------|------------|----------|
| Muter le plateau dans `PossibleMoves()` | Tous les états partagent le même tableau | Toujours copier le plateau avant de modifier |
| Oublier `lastAction` dans `PossibleMoves()` | `LastAction()` retourne 0 pour tous les coups | Affecter `lastAction: i` dans chaque état fils |
| Collisions d'`ID()` | Le MCTS confond des états différents | Inclure l'acteur courant dans l'ID |
| `Evaluate()` ne détecte pas le nul | Le MCTS boucle indéfiniment | Vérifier le plateau plein / absence de coups |
| `Play()` sans validation | État incohérent silencieux | Valider bornes, occupation, partie terminée |
