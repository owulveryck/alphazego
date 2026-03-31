# AlphaZeGo

Implementation d'AlphaZero en Go, from scratch. Le jeu de reference est le morpion (tic-tac-toe) pour sa simplicite ; l'objectif est d'ajouter progressivement les briques deep learning.

## Structure du projet

```
alphazego/
├── board/
│   ├── interfaces.go              # Interfaces generiques (State, Evaluator, Tensorizable, Playable)
│   └── tictactoe/
│       ├── ttt.go                 # Implementation du morpion (State + Tensorizable)
│       ├── console.go             # Affichage ANSI du plateau
│       ├── ttt_test.go            # Tests unitaires
│       ├── example_test.go        # Exemples testables (godoc)
│       └── cmd/main.go            # Programme jouable (humain vs IA)
├── mcts/
│   ├── doc.go                     # Documentation du package
│   ├── mcts.go                    # Boucle principale RunMCTS, NewMCTS, NewAlphaMCTS
│   ├── node.go                    # MCTSNode, IsTerminal, SelectChildUCB, SelectBestMove
│   ├── ucb1.go                    # Formule UCB1 (MCTS pur)
│   ├── puct.go                    # Formule PUCT (AlphaZero)
│   ├── expand.go                  # Expansion (Expand un-par-un, ExpandAll avec priors)
│   ├── simulate.go                # Rollout aleatoire (MCTS pur)
│   ├── backpropagate.go           # Retropropagation (discrete et continue)
│   ├── mcts_test.go               # Tests unitaires
│   ├── puct_test.go               # Tests PUCT, BackpropagateValue, ExpandAll, AlphaMCTS
│   └── example_test.go            # Exemples testables (godoc)
├── docs/                          # Documentation Divio (explanation, reference, how-to, tutorials)
├── main.go                        # Programme principal (humain vs IA)
└── README.md                      # Presentation et explication detaillee en francais
```

## Commandes

- `go build ./...` -- compiler tout
- `go test ./...` -- lancer tous les tests
- `go test -cover ./...` -- tests avec couverture
- `go run main.go` -- jouer au morpion contre l'IA
- `go run board/tictactoe/cmd/main.go` -- idem (version alternative)
- `goimports -w .` -- corriger les imports apres modification

## Conventions

- Langue du code : Go idiomatique, noms en anglais
- Langue de la doc/README/commentaires : francais
- Chaque symbole exporte doit avoir un commentaire godoc
- Les packages doivent avoir un `doc.go` ou un commentaire de package
- Les tests utilisent des `Example` functions (testables par `go test`) autant que possible
- Pas de table de transposition dans le MCTS : chaque noeud est cree independamment pour que la backpropagation remonte correctement via `parent`

## Regles de contribution

1. **Tests** : toute modification doit etre couverte par des tests. Viser un coverage > 90%. Privilegier les `Example` functions pour leur double role de test et de documentation.
2. **Godoc** : tout symbole exporte (type, fonction, methode, constante) doit avoir un commentaire godoc. Les packages doivent avoir une documentation de package.
3. **Documentation** : mettre a jour les fichiers dans `docs/` quand une fonctionnalite est ajoutee ou modifiee. Respecter la structure Divio (explanation, reference, how-to, tutorials).
4. **Imports** : apres toute modification d'un fichier `.go`, lancer `goimports` pour corriger les imports.
5. **README** : garder le README a jour avec la structure du code et les explications.
