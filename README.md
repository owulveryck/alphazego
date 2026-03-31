# alphazego

Une implementation d'AlphaZero en Go, from scratch.

## Introduction

L'objectif de ce projet est de comprendre AlphaGo et AlphaZero en les implementant de zero.
Le jeu utilise est le morpion (tic-tac-toe), car ses regles sont simples et permettent de se concentrer sur l'algorithme.

La premiere etape est d'implementer le MCTS (Monte Carlo Tree Search).
La suite sera d'ajouter les parties deep-learning.

## Qu'est-ce que le MCTS ?

Le MCTS est un algorithme de recherche utilise en intelligence artificielle pour choisir le meilleur coup dans un jeu.
Il est a la base d'AlphaGo et AlphaZero.

L'idee centrale : plutot que d'explorer toutes les possibilites (impossible pour un jeu comme le Go), on **simule des milliers de parties aleatoires** et on utilise les resultats pour guider la recherche vers les coups les plus prometteurs.

### Le jeu vu comme un arbre

Chaque position du jeu est un **noeud** de l'arbre. Chaque coup possible est une **branche** qui mene a un nouveau noeud (une nouvelle position).

```
Position actuelle (racine)
├── Coup A → Position A
│   ├── Reponse 1 → ...
│   └── Reponse 2 → ...
├── Coup B → Position B
│   └── Reponse 1 → ...
└── Coup C → Position C
    ├── Reponse 1 → ...
    └── Reponse 2 → ...
```

Le MCTS construit progressivement cet arbre en repetant quatre phases a chaque iteration.

### Les quatre phases

#### 1. Selection

On part de la racine et on descend dans l'arbre en choisissant a chaque etape l'enfant le plus prometteur, jusqu'a atteindre un noeud **feuille** (un noeud qui n'a pas encore ete entierement explore).

Le choix se fait avec la formule **UCB1** (voir plus bas), qui equilibre deux objectifs contradictoires :
- **Exploitation** : aller vers les coups qui ont bien marche jusqu'ici
- **Exploration** : essayer les coups peu visites, qui pourraient se reveler meilleurs

#### 2. Expansion

Quand on atteint un noeud feuille qui n'est pas en fin de partie, on ajoute **un seul** nouveau noeud enfant correspondant a un coup pas encore essaye.

#### 3. Simulation (rollout)

A partir du nouveau noeud, on joue une **partie aleatoire** jusqu'a la fin (victoire, defaite ou match nul). On ne cherche pas a bien jouer : on choisit des coups au hasard. L'idee est d'obtenir une estimation rapide de la valeur du coup.

#### 4. Retropropagation (backpropagation)

Le resultat de la simulation est **propage vers le haut** de l'arbre, du noeud simule jusqu'a la racine. Chaque noeud traverse met a jour deux compteurs :
- **visits** : combien de fois on est passe par ce noeud
- **wins** : combien de victoires ont ete observees pour le joueur qui a choisi ce coup

Apres des milliers d'iterations, l'arbre contient suffisamment de statistiques pour choisir le meilleur coup : celui dont l'enfant a recu le **plus de visites**.

## UCB1 : le compromis exploration/exploitation

### L'intuition

Imaginez que vous etes dans une fete foraine avec plein de stands de jeux. Vous avez un temps limite. Certains jeux, vous les connaissez et savez qu'ils sont amusants. D'autres, vous ne les avez jamais essayes.

Comment repartir votre temps ? Si vous ne jouez qu'aux jeux que vous connaissez, vous passez a cote de jeux potentiellement meilleurs. Si vous ne faites qu'essayer des nouveautes, vous perdez du temps sur des jeux mediocres.

UCB1 resout ce dilemme en donnant a chaque jeu un **score** qui combine :
- **La recompense moyenne** : les jeux amusants ont un score eleve
- **Un bonus d'exploration** : les jeux peu essayes recoivent un bonus qui diminue au fur et a mesure qu'on les teste

On choisit toujours le jeu avec le score le plus eleve.

### La formule

$$ \text{UCB1} = \underbrace{\frac{W}{N}}_{\text{exploitation}} + \underbrace{C \sqrt{\frac{\ln N_p}{N}}}_{\text{exploration}} $$

| Symbole | Signification |
|---------|---------------|
| $W$ | Nombre de victoires du noeud |
| $N$ | Nombre de visites du noeud |
| $N_p$ | Nombre de visites du noeud parent |
| $C$ | Constante d'exploration ($\sqrt{2}$ par defaut) |

- Un noeud jamais visite ($N = 0$) recoit un score de $+\infty$ pour garantir qu'il sera explore au moins une fois.
- Plus $C$ est grand, plus l'algorithme explore. Plus $C$ est petit, plus il exploite.

## Structure du code

```
alphazego/
├── board/
│   ├── interfaces.go          # Interfaces generiques (State, Evaluator, Tensorizable)
│   └── tictactoe/
│       ├── ttt.go             # Implementation du morpion
│       ├── console.go         # Affichage du plateau
│       └── cmd/main.go        # Petit programme jouable en console
├── mcts/
│   ├── doc.go                 # Documentation du package
│   ├── mcts.go                # Boucle principale (RunMCTS, NewMCTS, NewAlphaMCTS)
│   ├── node.go                # Noeud interne + methodes utilitaires
│   ├── ucb1.go                # Formule UCB1 (MCTS pur)
│   ├── puct.go                # Formule PUCT (AlphaZero)
│   ├── expand.go              # Phase d'expansion
│   ├── simulate.go            # Phase de simulation (rollout)
│   └── backpropagate.go       # Phase de retropropagation
├── docs/                      # Documentation Divio (explanation, reference, how-to, tutorials)
└── main.go                    # Programme principal (humain vs MCTS)
```

### Separer le jeu du MCTS

Le MCTS ne connait pas les regles du morpion. Il manipule des **interfaces** definies dans `board/interfaces.go` :

```go
type State interface {
    CurrentPlayer() PlayerID       // Quel joueur doit agir ?
    PreviousPlayer() PlayerID      // Quel joueur a effectue le dernier coup ?
    Evaluate() PlayerID            // Le gagnant, NoPlayer (en cours), ou DrawResult (nul)
    PossibleMoves() []State        // Quels sont les etats atteignables ?
    ID() string                    // Identifiant unique de l'etat
    LastMove() uint8               // Le coup qui a mene a cet etat
}
```

Cette interface est **generique** : elle ne presuppose ni un jeu de plateau, ni un nombre fixe de joueurs. Tout probleme de decision sequentiel a un ou plusieurs decideurs peut etre modelise ainsi -- jeux, negociations, planification, etc. `PreviousPlayer()` permet au moteur MCTS de savoir qui a joue le dernier coup sans connaitre la logique de tour. `Evaluate()` retourne directement le `PlayerID` du gagnant (pas de type `Result` separe) : `NoPlayer` (0) si le jeu est en cours, `DrawResult` (-1) en cas de match nul. Le morpion est un premier exemple simple (deux joueurs en alternance).

### Le morpion

Le plateau est represente par un slice de 9 cases (`[]uint8`), numerotees de 0 a 8 :

```
 0 | 1 | 2
───┼───┼───
 3 | 4 | 5
───┼───┼───
 6 | 7 | 8
```

Chaque case vaut `0` (vide), `1` (joueur 1 / X) ou `2` (joueur 2 / O). L'alternance des joueurs est geree par `3 - PlayerTurn` : si c'est au joueur 1, le prochain sera `3 - 1 = 2`, et vice versa.

```go
type TicTacToe struct {
    board      []uint8        // 9 cases
    PlayerTurn board.PlayerID // 1 ou 2
    lastMove   uint8          // coup qui a mene a cet etat
}

func (t *TicTacToe) Play(p uint8) error {
    // Valide bornes, occupation, et fin de partie
    t.board[p] = uint8(t.PlayerTurn)
    t.lastMove = p
    t.PlayerTurn = 3 - t.PlayerTurn
    return nil
}
```

### Le noeud MCTS

Chaque noeud de l'arbre represente une position de jeu et stocke les statistiques accumulees. Les noeuds sont internes au package `mcts` (type `mctsNode` non exporte) — l'API publique se limite a `NewMCTS()`, `NewAlphaMCTS()` et `RunMCTS()` :

```go
// Structure interne (non exportee)
type mctsNode struct {
    state    board.State      // La position de jeu
    parent   *mctsNode        // Le noeud parent (nil pour la racine)
    children []*mctsNode      // Les noeuds enfants (coups explores)
    wins     float64          // Victoires observees
    visits   float64          // Nombre de visites
    prior    float64          // Prior du policy network (AlphaZero)
    mcts     *MCTS            // Reference vers l'instance MCTS
}
```

## Implementation des quatre phases

### 1. Selection (`node.go`, `ucb1.go`)

La boucle de selection dans `RunMCTS` descend dans l'arbre tant que le noeud courant n'est pas terminal et que tous ses enfants ont ete explores au moins une fois :

```go
node := root
for !node.isTerminal() && node.isFullyExpanded() {
    node = node.selectChildUCB()
}
```

`selectChildUCB` choisit l'enfant avec le meilleur score UCB1 (ou PUCT en mode AlphaZero) parmi les enfants immediats. La methode est **non recursive** : c'est la boucle ci-dessus qui gere la descente dans l'arbre.

```go
func (n *mctsNode) selectChildUCB() *mctsNode {
    bestScore := math.Inf(-1)
    var bestChild *mctsNode
    for _, child := range n.children {
        score := child.ucb1() // ou child.puct() en mode AlphaZero
        if score > bestScore {
            bestScore = score
            bestChild = child
        }
    }
    return bestChild
}
```

Le score UCB1 est calcule dans `ucb1.go`. Un noeud jamais visite retourne `+Inf` pour etre explore en priorite :

```go
func (n *mctsNode) ucb1() float64 {
    if n.visits == 0 {
        return math.Inf(1)
    }
    C := math.Sqrt(2)
    avgReward := n.wins / n.visits
    if n.parent == nil {
        return avgReward
    }
    return avgReward + C*math.Sqrt(math.Log(n.parent.visits)/n.visits)
}
```

Les methodes utilitaires :
- `isTerminal()` : verifie si `Evaluate()` retourne autre chose que `NoPlayer` (partie finie)
- `isFullyExpanded()` : verifie si le nombre d'enfants est egal au nombre de coups possibles

### 2. Expansion (`expand.go`)

`expand` ajoute **un seul** nouvel enfant a chaque appel. Il determine les coups non encore explores en comparant les `ID` des enfants existants avec ceux des coups possibles :

```go
func (node *mctsNode) expand() *mctsNode {
    possibleMoves := node.state.PossibleMoves()

    existingIDs := make(map[string]bool)
    for _, child := range node.children {
        existingIDs[child.state.ID()] = true
    }

    for _, move := range possibleMoves {
        if !existingIDs[move.ID()] {
            child := &mctsNode{
                state:    move,
                parent:   node,
                children: []*mctsNode{},
                mcts:     node.mcts,
            }
            node.children = append(node.children, child)
            return child
        }
    }
    return nil
}
```

**Point important** : chaque noeud enfant est cree independamment. On ne partage pas les noeuds entre differentes branches via une table de transposition. Si un meme etat est atteint par deux chemins differents, deux noeuds distincts sont crees. C'est essentiel car la retropropagation remonte via le champ `parent` : partager un noeud ferait remonter les statistiques dans la mauvaise branche.

### 3. Simulation (`simulate.go`)

`simulate` joue une partie aleatoire depuis l'etat du noeud jusqu'a la fin. Les coups sont choisis au hasard parmi les coups possibles :

```go
func (node *mctsNode) simulate() board.PlayerID {
    currentState := node.state
    for currentState.Evaluate() == board.NoPlayer {
        possibleMoves := currentState.PossibleMoves()
        currentState = possibleMoves[rand.Intn(len(possibleMoves))]
    }
    return currentState.Evaluate()
}
```

`simulate` ne modifie jamais l'etat du noeud : elle travaille sur des copies locales creees par `PossibleMoves()`.

### 4. Retropropagation (`backpropagate.go`)

Apres la simulation, on remonte le resultat du noeud simule jusqu'a la racine. A chaque noeud, on incremente le compteur de visites et on credite une victoire si le resultat correspond au joueur qui a **fait le coup** menant a ce noeud.

Le point subtil : `CurrentPlayer()` retourne le joueur dont c'est le tour (celui qui va jouer), pas celui qui vient de jouer. Le joueur qui a amene la partie dans cet etat est `PreviousPlayer()`. Cela permet au moteur de fonctionner quel que soit le nombre de joueurs.

```go
func (node *mctsNode) backpropagate(result board.PlayerID) {
    for n := node; n != nil; n = n.parent {
        n.visits += 1

        playerWhoMovedHere := n.state.PreviousPlayer()
        if result == playerWhoMovedHere {
            n.wins += 1
        } else if result == board.DrawResult {
            n.wins += 0.5
        }
    }
}
```

Cette convention garantit que les `wins` d'un noeud representent les victoires du point de vue du **parent**. Quand le parent utilise UCB1 pour choisir parmi ses enfants, il compare directement leurs `wins/visits` : l'enfant avec le meilleur ratio est celui qui mene le plus souvent a une victoire pour le parent.

Les matchs nuls comptent pour 0.5, ce qui les place entre une victoire (1.0) et une defaite (0.0).

## Choix final du coup

Apres toutes les iterations, le MCTS choisit le coup correspondant a l'enfant de la racine qui a recu le **plus de visites** (pas le plus de victoires). Le nombre de visites est un indicateur plus fiable que le ratio de victoires car il reflete la confiance globale de l'algorithme.

```go
func (n *mctsNode) selectBestMove() *mctsNode {
    var bestChild *mctsNode
    bestVisits := float64(-1)
    for _, child := range n.children {
        if child.visits > bestVisits {
            bestVisits = child.visits
            bestChild = child
        }
    }
    return bestChild
}
```

## Utilisation

```bash
go run main.go
```

Le programme lance une partie de morpion ou vous jouez contre le MCTS. Entrez le numero de la case (0-8) pour jouer votre coup.

## Tests

```bash
go test ./...
```

Les tests couvrent :
- Le jeu de morpion : creation, coups, evaluation des victoires/matchs nuls, generation des coups possibles
- Le MCTS : UCB1, selection, expansion, simulation, retropropagation, et tests d'integration verifiant que l'IA **bloque les victoires adverses** et **saisit les coups gagnants**
