# Tutoriel : Le morpion pas a pas

Dans ce tutoriel, vous allez construire un morpion (tic-tac-toe) jouable contre une IA MCTS, en implementant l'interface `board.State` de zero.

## Ce que vous allez construire

- Un morpion 3x3 avec detection de victoire et de match nul
- Une IA basee sur le MCTS (Monte Carlo Tree Search)
- Une boucle de jeu interactive dans le terminal

Le code final correspond a l'implementation dans `board/tictactoe/` et `main.go`.

## Prerequis

- Go installe (1.21+)
- Le module `alphazego` disponible :

```bash
go mod init mon-morpion
go get github.com/owulveryck/alphazego
```

## Etape 1 : Le plateau

Le morpion est une grille 3x3 = 9 cases :

```
0 | 1 | 2
──┼───┼──
3 | 4 | 5
──┼───┼──
6 | 7 | 8
```

Chaque case contient `0` (vide), `1` (Player1 / X) ou `2` (Player2 / O).

Definissez le struct et le constructeur :

```go
package morpion

import "github.com/owulveryck/alphazego/board"

const BoardSize = 9

type Morpion struct {
    board      []uint8
    playerTurn board.PlayerID
    lastMove   uint8
}

func New() *Morpion {
    return &Morpion{
        board:      make([]uint8, BoardSize),
        playerTurn: board.Player1,
    }
}
```

Trois champs :
- `board` : les 9 cases du plateau
- `playerTurn` : qui doit jouer (alternance Player1/Player2)
- `lastMove` : le coup qui a mene a cet etat (necessaire pour l'interface `State`)

## Etape 2 : `Evaluate()` — detecter la fin de partie

Les combinaisons gagnantes au morpion sont 8 : 3 lignes, 3 colonnes, 2 diagonales.

```go
var winningPositions = [][]uint8{
    {0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // lignes
    {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // colonnes
    {0, 4, 8}, {2, 4, 6},             // diagonales
}

func (m *Morpion) Evaluate() board.PlayerID {
    for _, pos := range winningPositions {
        if m.board[pos[0]] != 0 &&
            m.board[pos[0]] == m.board[pos[1]] &&
            m.board[pos[1]] == m.board[pos[2]] {
            return board.PlayerID(m.board[pos[0]])
        }
    }
    // Match nul si toutes les cases sont occupees
    for _, cell := range m.board {
        if cell == 0 {
            return board.NoPlayer // partie en cours
        }
    }
    return board.DrawResult
}
```

**Verification** : ecrivez un test pour valider les cas courants.

```go
func TestEvaluate_Player1Wins(t *testing.T) {
    m := New()
    // X en haut : positions 0, 1, 2
    m.board[0], m.board[1], m.board[2] = 1, 1, 1
    if m.Evaluate() != board.Player1 {
        t.Error("Player1 devrait gagner avec la ligne du haut")
    }
}

func TestEvaluate_Draw(t *testing.T) {
    m := New()
    m.board = []uint8{1, 2, 1, 1, 1, 2, 2, 1, 2}
    if m.Evaluate() != board.DrawResult {
        t.Error("devrait etre un match nul")
    }
}

func TestEvaluate_InProgress(t *testing.T) {
    m := New()
    if m.Evaluate() != board.NoPlayer {
        t.Error("plateau vide = partie en cours")
    }
}
```

```bash
go test -v ./...
```

## Etape 3 : `PossibleMoves()` — generer les etats fils

C'est la methode la plus importante pour le MCTS. Elle retourne un `[]board.State` ou chaque element est un plateau avec un coup en plus.

**Regle critique** : ne jamais modifier `m.board` directement. Chaque etat fils doit etre une copie independante.

```go
func (m *Morpion) PossibleMoves() []board.State {
    var moves []board.State
    for i := 0; i < BoardSize; i++ {
        if m.board[i] == 0 {
            // Copier le plateau
            newBoard := make([]uint8, BoardSize)
            copy(newBoard, m.board)
            newBoard[i] = uint8(m.playerTurn)
            moves = append(moves, &Morpion{
                board:      newBoard,
                playerTurn: 3 - m.playerTurn,
                lastMove:   uint8(i),
            })
        }
    }
    return moves
}
```

Points importants :
- On copie le slice avec `copy()` — sans cela, tous les etats partagent le meme tableau
- On alterne le joueur avec `3 - m.playerTurn` (1↔2)
- On enregistre `lastMove` pour que `LastMove()` fonctionne

## Etape 4 : Les autres methodes de State

```go
func (m *Morpion) CurrentPlayer() board.PlayerID {
    return m.playerTurn
}

func (m *Morpion) PreviousPlayer() board.PlayerID {
    return 3 - m.playerTurn
}

func (m *Morpion) ID() string {
    id := make([]byte, BoardSize+1)
    copy(id, m.board)
    id[BoardSize] = byte(m.playerTurn)
    return string(id)
}

func (m *Morpion) LastMove() uint8 {
    return m.lastMove
}
```

L'`ID()` encode le plateau + le joueur courant en une chaine de 10 octets. C'est suffisant pour identifier de maniere unique chaque position.

## Etape 5 : Premier test MCTS

Avant de construire l'interface utilisateur, verifions que le MCTS fonctionne avec notre implementation.

```go
func TestMCTS_FullGame(t *testing.T) {
    m := mcts.NewMCTS()
    game := New()

    moves := 0
    for game.Evaluate() == board.NoPlayer {
        bestState := m.RunMCTS(game, 500)
        move := bestState.LastMove()
        game.board[move] = uint8(game.playerTurn) // appliquer le coup
        game.playerTurn = 3 - game.playerTurn
        moves++
    }

    if game.Evaluate() == board.NoPlayer {
        t.Error("la partie devrait etre terminee")
    }
    if moves < 5 || moves > 9 {
        t.Errorf("nombre de coups invalide : %d", moves)
    }
}
```

```bash
go test -v -run TestMCTS
```

## Etape 6 : `Play()` pour l'interaction humaine

`Play()` n'est pas dans l'interface `State`, mais permet a un humain de jouer :

```go
func (m *Morpion) Play(p uint8) error {
    if p >= BoardSize {
        return fmt.Errorf("position %d hors limites (0-%d)", p, BoardSize-1)
    }
    if m.board[p] != 0 {
        return fmt.Errorf("position %d deja occupee", p)
    }
    if m.Evaluate() != board.NoPlayer {
        return fmt.Errorf("la partie est terminee")
    }
    m.board[p] = uint8(m.playerTurn)
    m.lastMove = p
    m.playerTurn = 3 - m.playerTurn
    return nil
}
```

## Etape 7 : Affichage du plateau

Pour un affichage agreable dans le terminal, utilisez des couleurs ANSI :

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

## Etape 8 : La boucle de jeu complete

Assemblez le tout dans un `main.go` :

```go
package main

import (
    "fmt"
    "log"
    "strconv"

    "github.com/owulveryck/alphazego/board"
    "github.com/owulveryck/alphazego/mcts"
)

func main() {
    game := morpion.New()
    m := mcts.NewMCTS()

    for game.Evaluate() == board.NoPlayer {
        fmt.Println(game)

        // Tour de l'humain
        fmt.Print("Votre coup (0-8) : ")
        var input string
        fmt.Scan(&input)
        val, err := strconv.ParseUint(input, 10, 8)
        if err != nil {
            fmt.Println("Entree invalide")
            continue
        }
        if err := game.Play(uint8(val)); err != nil {
            fmt.Println("Coup invalide :", err)
            continue
        }

        // Verifier si la partie est finie
        if game.Evaluate() != board.NoPlayer {
            break
        }

        // Tour de l'IA
        bestState := m.RunMCTS(game, 1000)
        aiMove := bestState.LastMove()
        fmt.Printf("L'IA joue en %d\n", aiMove)
        game.Play(aiMove)
    }

    // Resultat
    fmt.Println(game)
    switch game.Evaluate() {
    case board.Player1:
        fmt.Println("Vous avez gagne !")
    case board.Player2:
        fmt.Println("L'IA a gagne !")
    case board.DrawResult:
        fmt.Println("Match nul !")
    }
}
```

```bash
go run main.go
```

## Pour aller plus loin

- **Augmenter les iterations** : plus d'iterations = IA plus forte (essayez 5000 ou 10000)
- **Implementer `Tensorizable`** : pour connecter un reseau de neurones, voir la [reference des interfaces](../reference/interfaces-evaluator.md)
- **Implementer un `Evaluator`** : pour remplacer les rollouts aleatoires par une evaluation intelligente, voir le [how-to Evaluator](../how-to/implementer-evaluator.md)
- **Mode AlphaZero** : utiliser `mcts.NewAlphaMCTS(evaluator, cpuct)` avec un reseau entraine, voir [de MCTS a AlphaZero](../explanation/de-mcts-a-alphazero.md)
- **Autre jeu** : adaptez ce tutoriel a un autre jeu en suivant le [how-to implementer un jeu](../how-to/implementer-un-jeu.md)
