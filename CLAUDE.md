# AlphaZeGo

Implémentation d'AlphaZero en Go, from scratch. Le jeu de référence est le morpion (tic-tac-toe) pour sa simplicité ; l'objectif est d'ajouter progressivement les briques deep learning.

## Structure du projet

```
alphazego/
├── decision/
│   ├── doc.go                     # Documentation du package
│   ├── state.go                   # Interfaces génériques (State, ActorID)
│   └── board/
│       ├── doc.go                 # Documentation du package
│       ├── board.go               # Interfaces plateau (Boarder, ActionRecorder, Tensorizable)
│       ├── example_test.go        # Exemples testables (godoc)
│       ├── tictactoe/
│       │   ├── ttt.go             # Implémentation du morpion (State + ActionRecorder + Tensorizable)
│       │   ├── console.go         # Affichage ANSI du plateau
│       │   ├── ttt_test.go        # Tests unitaires
│       │   ├── example_test.go    # Exemples testables (godoc)
│       │   └── cmd/main.go        # Programme jouable (humain vs IA)
│       └── taquin/
│           ├── doc.go             # Documentation du package
│           ├── taquin.go          # Puzzle à glissement (State + ActionRecorder + Tensorizable, 1 acteur)
│           ├── console.go         # Affichage ANSI du plateau
│           ├── taquin_test.go     # Tests unitaires
│           ├── console_test.go    # Tests d'affichage
│           ├── example_test.go    # Exemples testables (MCTS résout le taquin)
│           └── cmd/main.go        # IA résout le taquin avec MCTS
├── mcts/
│   ├── doc.go                     # Documentation du package
│   ├── evaluator.go               # Interface Evaluator (policy + value)
│   ├── mcts.go                    # Boucle principale RunMCTS, NewMCTS, NewAlphaMCTS
│   ├── node.go                    # mctsNode (interne), isTerminal, selectChildUCB, selectBestMove
│   ├── ucb1.go                    # Formule UCB1 (MCTS pur)
│   ├── puct.go                    # Formule PUCT (AlphaZero)
│   ├── expand.go                  # Expansion (expand un-par-un, expandAll avec priors)
│   ├── simulate.go                # Rollout aléatoire (MCTS pur)
│   ├── backpropagate.go           # Rétropropagation (discrète et continue)
│   ├── mcts_test.go               # Tests unitaires
│   ├── puct_test.go               # Tests PUCT, backpropagateValue, expandAll, AlphaMCTS
│   └── example_test.go            # Exemples testables (godoc)
├── docs/                          # Documentation Divio (explanation, référence, how-to, tutorials)
├── main.go                        # Programme principal (humain vs IA)
└── README.md                      # Présentation et explication détaillée en français
```

## Commandes

- `go build ./...` -- compiler tout
- `go test ./...` -- lancer tous les tests
- `go test -cover ./...` -- tests avec couverture
- `go run main.go` -- jouer au morpion contre l'IA
- `go run decision/board/tictactoe/cmd/main.go` -- idem (version alternative)
- `go run decision/board/taquin/cmd/main.go` -- l'IA résout un taquin 3x2 avec MCTS
- `goimports -w .` -- corriger les imports après modification

## Conventions

- Langue du code : Go idiomatique, noms en anglais
- Langue de la doc/README/commentaires : français
- Chaque symbole exporté doit avoir un commentaire godoc
- Les packages doivent avoir un `doc.go` ou un commentaire de package
- Les tests utilisent des `Example` functions (testables par `go test`) autant que possible
- Pas de table de transposition dans le MCTS : chaque nœud est créé indépendamment pour que la backpropagation remonte correctement via `parent`

## Règles de contribution

1. **Tests** : toute modification doit être couverte par des tests. Viser un coverage > 90%. Privilégier les `Example` functions pour leur double rôle de test et de documentation.
2. **Godoc** : tout symbole exporté (type, fonction, méthode, constante) doit avoir un commentaire godoc. Les packages doivent avoir une documentation de package.
3. **Documentation** : mettre à jour les fichiers dans `docs/` quand une fonctionnalité est ajoutée ou modifiée. Respecter la structure Divio (explanation, référence, how-to, tutorials).
4. **Imports** : après toute modification d'un fichier `.go`, lancer `goimports` pour corriger les imports.
5. **README** : garder le README à jour avec la structure du code et les explications.
