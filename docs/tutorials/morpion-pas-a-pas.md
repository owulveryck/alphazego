# Tutoriel : Le morpion pas à pas

Dans ce tutoriel, vous allez construire un morpion (tic-tac-toe) jouable contre une IA MCTS, en implémentant l'interface `decision.State` de zéro.

## Ce que vous allez construire

- Un morpion 3x3 avec détection de victoire et de match nul
- Une IA basée sur le MCTS (Monte Carlo Tree Search)
- Une boucle de jeu interactive dans le terminal

Le code final correspond à l'implémentation dans `decision/board/samples/tictactoe/` et `main.go`.

## Prérequis

- Go installé (1.21+)
- Le module `alphazego` disponible :

```bash
go mod init mon-morpion
go get github.com/owulveryck/alphazego
```

## Étape 1 : Le plateau

Le morpion est une grille 3x3 = 9 cases :

```
0 | 1 | 2
──┼───┼──
3 | 4 | 5
──┼───┼──
6 | 7 | 8
```

Chaque case contient `0` (vide), `1` (Actor1 / X) ou `2` (Actor2 / O).

Définissez le struct et le constructeur :

```go
package morpion

import "github.com/owulveryck/alphazego/decision"

const BoardSize = 9

type Morpion struct {
    board      []uint8
    actorTurn  decision.ActorID
    lastAction int
}

func New() *Morpion {
    return &Morpion{
        board:     make([]uint8, BoardSize),
        actorTurn: decision.ActorID(1),
    }
}
```

Trois champs :
- `board` : les 9 cases du plateau
- `actorTurn` : qui doit jouer (alternance Actor1/Actor2)
- `lastAction` : l'action qui a mené à cet état (pour l'interface `ActionRecorder`)

## Étape 2 : `Evaluate()` -- détecter la fin de partie

Les combinaisons gagnantes au morpion sont 8 : 3 lignes, 3 colonnes, 2 diagonales.

```go
var winningPositions = [][]uint8{
    {0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // lignes
    {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // colonnes
    {0, 4, 8}, {2, 4, 6},             // diagonales
}

func (m *Morpion) Evaluate() decision.ActorID {
    for _, pos := range winningPositions {
        if m.board[pos[0]] != 0 &&
            m.board[pos[0]] == m.board[pos[1]] &&
            m.board[pos[1]] == m.board[pos[2]] {
            return decision.ActorID(m.board[pos[0]])
        }
    }
    // Match nul si toutes les cases sont occupées
    for _, cell := range m.board {
        if cell == 0 {
            return decision.Undecided // partie en cours
        }
    }
    return decision.Stalemate
}
```

**Vérification** : écrivez un test pour valider les cas courants.

```go
func TestEvaluate_Actor1Wins(t *testing.T) {
    m := New()
    // X en haut : positions 0, 1, 2
    m.board[0], m.board[1], m.board[2] = 1, 1, 1
    if m.Evaluate() != decision.ActorID(1) {
        t.Error("Actor1 devrait gagner avec la ligne du haut")
    }
}

func TestEvaluate_Draw(t *testing.T) {
    m := New()
    m.board = []uint8{1, 2, 1, 1, 1, 2, 2, 1, 2}
    if m.Evaluate() != decision.Stalemate {
        t.Error("devrait être un match nul")
    }
}

func TestEvaluate_InProgress(t *testing.T) {
    m := New()
    if m.Evaluate() != decision.Undecided {
        t.Error("plateau vide = partie en cours")
    }
}
```

```bash
go test -v ./...
```

## Étape 3 : `PossibleMoves()` -- générer les états fils

C'est la méthode la plus importante pour le MCTS. Elle retourne un `[]decision.State` où chaque élément est un plateau avec un coup en plus.

**Règle critique** : ne jamais modifier `m.board` directement. Chaque état fils doit être une copie indépendante.

```go
func (m *Morpion) PossibleMoves() []decision.State {
    var moves []decision.State
    for i := 0; i < BoardSize; i++ {
        if m.board[i] == 0 {
            // Copier le plateau
            newBoard := make([]uint8, BoardSize)
            copy(newBoard, m.board)
            newBoard[i] = uint8(m.actorTurn)
            moves = append(moves, &Morpion{
                board:      newBoard,
                actorTurn:  3 - m.actorTurn,
                lastAction: i,
            })
        }
    }
    return moves
}
```

Points importants :
- On copie le slice avec `copy()` -- sans cela, tous les états partagent le même tableau
- On alterne l'acteur avec `3 - m.actorTurn` (1↔2)
- On enregistre `lastAction` pour que `LastAction()` fonctionne (interface `ActionRecorder`)

## Étape 4 : Les autres méthodes de State

```go
func (m *Morpion) CurrentActor() decision.ActorID {
    return m.actorTurn
}

func (m *Morpion) PreviousActor() decision.ActorID {
    return 3 - m.actorTurn
}

func (m *Morpion) ID() string {
    id := make([]byte, BoardSize+1)
    copy(id, m.board)
    id[BoardSize] = byte(m.actorTurn)
    return string(id)
}

func (m *Morpion) LastAction() int {
    return m.lastAction
}
```

L'`ID()` encode le plateau + l'acteur courant en une chaîne de 10 octets. C'est suffisant pour identifier de manière unique chaque position.

## Étape 5 : Premier test MCTS

Avant de construire l'interface utilisateur, vérifions que le MCTS fonctionne avec notre implémentation.

```go
func TestMCTS_FullGame(t *testing.T) {
    m := mcts.NewMCTS()
    game := New()

    moves := 0
    for game.Evaluate() == decision.Undecided {
        bestState := m.RunMCTS(game, 500)
        move := bestState.(board.ActionRecorder).LastAction()
        game.board[move] = uint8(game.actorTurn) // appliquer le coup
        game.actorTurn = 3 - game.actorTurn
        moves++
    }

    if game.Evaluate() == decision.Undecided {
        t.Error("la partie devrait être terminée")
    }
    if moves < 5 || moves > 9 {
        t.Errorf("nombre de coups invalide : %d", moves)
    }
}
```

```bash
go test -v -run TestMCTS
```

## Étape 6 : `Play()` pour l'interaction humaine

`Play()` n'est pas dans l'interface `State`, mais permet à un humain de jouer :

```go
func (m *Morpion) Play(p uint8) error {
    if p >= BoardSize {
        return fmt.Errorf("position %d hors limites (0-%d)", p, BoardSize-1)
    }
    if m.board[p] != 0 {
        return fmt.Errorf("position %d déjà occupée", p)
    }
    if m.Evaluate() != decision.Undecided {
        return fmt.Errorf("la partie est terminée")
    }
    m.board[p] = uint8(m.actorTurn)
    m.lastAction = int(p)
    m.actorTurn = 3 - m.actorTurn
    return nil
}
```

## Étape 7 : Affichage du plateau

Pour un affichage agréable dans le terminal, utilisez des couleurs ANSI :

```go
func (m *Morpion) String() string {
    symbols := map[uint8]string{
        0: " ",
        1: "\033[31mX\033[0m", // rouge
        2: "\033[34mO\033[0m", // bleu
    }
    var b strings.Builder
    b.WriteString(" ┌───┬───┬───┐\n")
    for row := 0; row < 3; row++ {
        b.WriteString(" │")
        for col := 0; col < 3; col++ {
            b.WriteString(" " + symbols[m.board[row*3+col]] + " │")
        }
        b.WriteString("\n")
        if row < 2 {
            b.WriteString(" ├───┼───┼───┤\n")
        }
    }
    b.WriteString(" └───┴───┴───┘\n")
    return b.String()
}
```

## Étape 8 : La boucle de jeu complète

Assemblez le tout dans un `main.go` :

```go
package main

import (
    "fmt"
    "log"
    "strconv"

    "github.com/owulveryck/alphazego/decision"
    "github.com/owulveryck/alphazego/decision/board"
    "github.com/owulveryck/alphazego/mcts"
)

func main() {
    game := morpion.New()
    m := mcts.NewMCTS()

    for game.Evaluate() == decision.Undecided {
        fmt.Println(game)

        // Tour de l'humain
        fmt.Print("Votre coup (0-8) : ")
        var input string
        fmt.Scan(&input)
        val, err := strconv.ParseUint(input, 10, 8)
        if err != nil {
            fmt.Println("Entrée invalide")
            continue
        }
        if err := game.Play(uint8(val)); err != nil {
            fmt.Println("Coup invalide :", err)
            continue
        }

        // Vérifier si la partie est finie
        if game.Evaluate() != decision.Undecided {
            break
        }

        // Tour de l'IA
        bestState := m.RunMCTS(game, 1000)
        aiMove := bestState.(board.ActionRecorder).LastAction()
        fmt.Printf("L'IA joue en %d\n", aiMove)
        game.Play(uint8(aiMove))
    }

    // Résultat
    fmt.Println(game)
    switch game.Evaluate() {
    case decision.ActorID(1):
        fmt.Println("Vous avez gagné !")
    case decision.ActorID(2):
        fmt.Println("L'IA a gagné !")
    case decision.Stalemate:
        fmt.Println("Match nul !")
    }
}
```

```bash
go run decision/board/samples/tictactoe/cmd/main.go
```

## Pour aller plus loin

- **Augmenter les itérations** : plus d'itérations = IA plus forte (essayez 5000 ou 10000)
- **Implémenter `Tensorizable`** : pour connecter un réseau de neurones, voir la [référence des interfaces](../reference/interfaces-evaluator.md)
- **Implémenter un `Evaluator`** : pour remplacer les rollouts aléatoires par une évaluation intelligente, voir le [how-to Evaluator](../how-to/implementer-evaluator.md)
- **Mode AlphaZero** : utiliser `mcts.NewAlphaMCTS(evaluator, cpuct)` avec un réseau entraîné, voir [de MCTS à AlphaZero](../explanation/de-mcts-a-alphazero.md)
- **Autre jeu** : adaptez ce tutoriel à un autre jeu en suivant le [how-to implémenter un jeu](../how-to/implementer-un-jeu.md)
