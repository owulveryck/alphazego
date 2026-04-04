# AlphaZeGo

Implémentation d'AlphaZero en Go, from scratch. Le jeu de référence est le morpion (tic-tac-toe) pour sa simplicité ; l'objectif est d'ajouter progressivement les briques deep learning.

## Structure du projet

```
alphazego/
├── mcts/                              # CORE — Moteur Monte Carlo Tree Search
│   ├── doc.go                         # Documentation du package
│   ├── evaluator.go                   # Interface Evaluator (policy + value)
│   ├── mcts.go                        # Boucle principale RunMCTS, NewMCTS, NewAlphaMCTS
│   ├── node.go                        # mctsNode (interne), isTerminal, selectChildUCB, selectBestMove
│   ├── ucb1.go                        # Formule UCB1 (MCTS pur)
│   ├── puct.go                        # Formule PUCT (AlphaZero)
│   ├── expand.go                      # Expansion (expand un-par-un, expandAll avec priors)
│   ├── simulate.go                    # Rollout aléatoire (MCTS pur)
│   ├── backpropagate.go               # Rétropropagation (discrète et continue)
│   ├── mcts_test.go                   # Tests unitaires
│   ├── puct_test.go                   # Tests PUCT, backpropagateValue, expandAll, AlphaMCTS
│   └── example_test.go               # Exemples testables (godoc)
├── decision/                          # CORE — Framework de décision
│   ├── doc.go                         # Documentation du package
│   ├── state.go                       # Interfaces génériques (State, ActorID)
│   ├── board/                         # Abstraction plateau de jeu
│   │   ├── board.go                   # Interfaces (Boarder, ActionRecorder, Tensorizable)
│   │   ├── example_test.go            # Exemples testables (godoc)
│   │   └── samples/                   # Implémentations de référence
│   │       ├── tictactoe/             # Morpion (State + ActionRecorder + Tensorizable)
│   │       │   └── cmd/main.go        # Programme jouable (humain vs IA)
│   │       └── taquin/                # Puzzle à glissement (1 acteur)
│   │           └── cmd/main.go        # IA résout le taquin avec MCTS
│   └── reasoning/                     # Raisonnement LLM via MCTS
│       ├── reasoning.go               # State, Generator, Judge interfaces
│       ├── evaluator.go               # Evaluator wrapping Judge pour MCTS
│       └── prompt.go                  # Templates de prompts
├── exp/benchmark/                         # Benchmarks et providers LLM
│   ├── problems/                      # Problèmes d'ordonnancement partagés (root module)
│   │   └── problems.go
│   ├── vertexai/                      # Module séparé — Provider Google Vertex AI
│   │   ├── go.mod
│   │   ├── vertexai.go                # Generator + Judge (Gemini)
│   │   └── cmd/
│   │       ├── benchmark/             # Benchmark 4 configs (A-D)
│   │       └── reasoning/             # CLI de raisonnement
│   └── ollama/                        # Module séparé — Provider Ollama (local)
│       ├── go.mod
│       ├── ollama.go                  # Generator + Judge (modèle local)
│       └── cmd/
│           └── benchmark/             # Benchmark 2 configs (E-F)
├── docs/                              # Documentation Divio (explanation, référence, how-to, tutorials)
└── README.md
```

## Commandes

- `go build ./...` -- compiler le module principal
- `go test ./...` -- lancer tous les tests
- `go test -cover ./...` -- tests avec couverture
- `go run decision/board/samples/tictactoe/cmd/main.go` -- jouer au morpion contre l'IA
- `go run decision/board/samples/taquin/cmd/main.go` -- l'IA résout un taquin 3x2 avec MCTS
- `cd exp/benchmark/vertexai && go run ./cmd/exp/benchmark/` -- benchmark Vertex AI
- `cd exp/benchmark/ollama && go run ./cmd/exp/benchmark/ -model qwen2.5:7b` -- benchmark Ollama
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
