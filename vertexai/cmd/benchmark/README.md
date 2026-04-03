# Benchmark : MCTS + LLM pour l'ordonnancement de taches

## Objectif

Ce benchmark mesure si un **petit modele LLM assiste par MCTS** peut rivaliser avec un **gros modele LLM seul** sur des problemes de planification sous contraintes.

L'hypothese est que la recherche arborescente (MCTS) compense les faiblesses de raisonnement des petits modeles en explorant systematiquement plusieurs chemins de decision, la ou un appel unique ("one-shot") peut s'engager dans un raisonnement sous-optimal sans possibilite de retour en arriere.

## Principe

Chaque probleme est un **ordonnancement de taches** avec des dependances et des durees. L'objectif est de trouver l'ordre d'execution qui minimise le temps total (makespan), en respectant toutes les contraintes de precedence.

Exemple (construction maison) :

```
Fondations (3j) --> Murs (5j) --> Toiture (2j)
                             \--> Electricite (3j)
                             \--> Plomberie (2j)
                                         \-------> Finitions (2j)
```

Chemin critique : Fondations -> Murs -> Electricite -> Finitions = **13 jours**.

Un petit modele peut oublier une dependance, mal calculer le chemin critique, ou ne pas voir que certaines taches sont parallelisables. Le MCTS explore differentes strategies de planification et identifie la meilleure.

## Les 4 configurations comparees

| Config | Modele | Methode | Description |
|--------|--------|---------|-------------|
| **A** | `gemini-3.1-flash-lite-preview` | One-shot | Petit modele, un seul appel |
| **B** | `gemini-3.1-flash-lite-preview` | MCTS | Petit modele + raisonnement arborescent |
| **C** | `gemini-3.1-pro-preview` | One-shot | Gros modele, un seul appel |
| **D** | `gemini-3.1-pro-preview` | MCTS | Gros modele + raisonnement arborescent |

La comparaison cle est **A vs B** : le MCTS aide-t-il un petit modele ? La comparaison secondaire est **B vs C** : un petit modele + MCTS fait-il aussi bien qu'un gros modele seul ?

## Les 10 problemes

Le benchmark comprend 10 problemes de difficulte croissante, avec des structures de dependances variees :

| # | Probleme | Taches | Structure | Makespan optimal |
|---|----------|--------|-----------|------------------|
| 1 | Chaine lineaire | 4 | A->B->C->D | 8 jours |
| 2 | Fourche parallele | 5 | A->{B,C,D}->E | 6 jours |
| 3 | Diamant | 5 | A->{B,C}->D->E | 10 jours |
| 4 | Construction maison | 6 | Multi-dependances | 13 jours |
| 5 | Deploiement logiciel | 7 | Merge de branches | 10 jours |
| 6 | Organisation evenement | 7 | 2 racines paralleles | 6 jours |
| 7 | Projet web fullstack | 8 | 3 branches paralleles | 12 jours |
| 8 | Pipeline data ETL | 9 | 3 sources -> fusion | 12 jours |
| 9 | Renovation appartement | 10 | Dependances croisees | 18 jours |
| 10 | Lancement produit | 12 | Tech/marketing/legal | 25 jours |

Chaque probleme a un **makespan optimal** calcule manuellement via le chemin critique. C'est la reference contre laquelle les solutions sont evaluees.

## Comment fonctionne le mode MCTS

En mode MCTS (configs B et D), le probleme est traite via le package `decision/reasoning` :

1. **State** = le raisonnement construit jusqu'ici (question + etapes precedentes)
2. **PossibleMoves** = le LLM genere 3 etapes de raisonnement candidates (temperature=0.8 pour la diversite)
3. **Evaluator** = le meme LLM (temperature=0.0) score la qualite de chaque chemin de raisonnement
4. **MCTS** = explore l'arbre des raisonnements possibles avec PUCT (15 iterations par etape)
5. **Terminal** = le LLM produit une etape commencant par `CONCLUSION:` avec l'ordonnancement final

```
                    [Probleme]
                   /     |     \
          [Etape A]  [Etape B]  [Etape C]    <- 3 candidats generes
           /    \       |
    [Suite A1] [A2]  [Suite B1]               <- exploration MCTS
       |
  [CONCLUSION]                                <- meilleure solution trouvee
```

Le MCTS selectionne a chaque niveau l'etape de raisonnement la plus prometteuse, en equilibrant exploration (essayer de nouvelles approches) et exploitation (approfondir les pistes prometteuses).

## Evaluation : LLM-as-Judge

Chaque solution est evaluee par un modele juge (`gemini-2.5-flash`) qui verifie :

1. **Les dependances sont-elles respectees ?** (une tache ne commence qu'apres la fin de ses dependances)
2. **Le makespan est-il correct ?** (le calcul du temps total est-il juste)
3. **Le makespan est-il optimal ?** (correspond-il au chemin critique)

Scoring :

| Score | Signification |
|-------|--------------|
| **1.0** | Dependances respectees ET makespan optimal |
| **0.5** | Dependances respectees MAIS makespan non optimal |
| **0.0** | Dependances violees ou reponse incomprehensible |

## Prerequis

- Go 1.22+
- Un projet Google Cloud avec l'API Vertex AI activee
- Authentification configuree (`gcloud auth application-default login`)

## Utilisation

```bash
# Variables d'environnement requises
export GCP_PROJECT=mon-projet-gcp
export GCP_REGION=us-central1

# Depuis le repertoire vertexai/
cd vertexai

# Lancer le benchmark complet (10 problemes x 4 configs = 40 evaluations)
go run ./cmd/benchmark/

# Filtrer par probleme (substring sur le nom)
go run ./cmd/benchmark/ -problem "maison"

# Filtrer par configuration
go run ./cmd/benchmark/ -config "A"        # petit modele seul
go run ./cmd/benchmark/ -config "MCTS"     # les 2 configs MCTS

# Combiner les filtres
go run ./cmd/benchmark/ -problem "lineaire" -config "A"
```

## Sortie attendue

Pendant l'execution, le benchmark affiche la progression :

```
Benchmark : 10 problemes x 4 configurations
Projet: mon-projet, Region: us-central1

━━━ 1/10 : Chaine lineaire (4 taches, optimal=8) ━━━
  A (flash-lite) ... score=1.0 (Dependances OK, makespan optimal)
  B (flash-lite+MCTS) ... score=1.0 (Dependances OK, makespan optimal)
  C (pro) ... score=1.0 (Dependances OK, makespan optimal)
  D (pro+MCTS) ... score=1.0 (Dependances OK, makespan optimal)

━━━ 2/10 : Fourche parallele (5 taches, optimal=6) ━━━
  A (flash-lite) ... score=0.5 (Dependances OK, makespan=7 au lieu de 6)
  B (flash-lite+MCTS) ... score=1.0 (Dependances OK, makespan optimal)
  ...
```

A la fin, un rapport recapitulatif :

```
================================================================================
RAPPORT FINAL
================================================================================
Probleme                       | A (flash-lite)   | B (flash+MCTS)   | C (pro)          | D (pro+MCTS)
-------------------------------+------------------+------------------+------------------+-----------------
Chaine lineaire                | 1.0              | 1.0              | 1.0              | 1.0
Fourche parallele              | 0.5              | 1.0              | 1.0              | 1.0
Diamant                        | 0.5              | 1.0              | 1.0              | 1.0
Construction maison            | 0.0              | 0.5              | 1.0              | 1.0
...
-------------------------------+------------------+------------------+------------------+-----------------
ACCURACY                       | 35%              | 75%              | 80%              | 95%
================================================================================
```

## Architecture

```
vertexai/cmd/benchmark/
├── main.go        # Point d'entree, boucle probleme x config, rapport final
├── problems.go    # Definition des 10 problemes (taches, durees, dependances)
├── runner.go      # Execution one-shot et MCTS pour chaque configuration
├── judge.go       # LLM-as-judge : evaluation des solutions
└── README.md      # Ce fichier
```

Dependances internes (module principal `alphazego`) :

- `decision/reasoning` -- interfaces `Generator`, `Judge`, `State` pour le raisonnement arborescent
- `mcts` -- moteur MCTS avec support AlphaZero (PUCT + evaluator)
- `vertexai` -- implementations Vertex AI des interfaces Generator/Judge

## Interpretation des resultats

| Resultat | Interpretation |
|----------|---------------|
| B >> A | Le MCTS aide significativement le petit modele |
| B ~= C | Le petit modele + MCTS rivalise avec le gros modele seul |
| D > C | Le MCTS aide aussi le gros modele |
| A ~= B | Le MCTS n'apporte pas de valeur (problemes trop simples ou iterations insuffisantes) |

Les problemes faciles (1-3) devraient etre reussis par toutes les configurations. La differentiation se fait sur les problemes difficiles (7-10), ou les dependances croisees et le nombre de taches rendent le raisonnement one-shot fragile.
