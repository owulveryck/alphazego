# wardley-explore — Exploration stratégique de cartes Wardley via MCTS + LLM

## Qu'est-ce que c'est ?

`wardley-explore` traite une carte Wardley comme un **puzzle stratégique** et utilise
l'algorithme Monte Carlo Tree Search (MCTS) pour explorer systématiquement l'espace
des décisions possibles. Un LLM (Gemini via Vertex AI) joue le rôle de **juge
stratégique**, évaluant la qualité de chaque état de la carte.

L'outil prend en entrée un fichier `.wtg2` (le DSL de
[wardleyToGo](https://github.com/owulveryck/wardleyToGo)), explore les séquences
de moves stratégiques, et retourne la meilleure séquence trouvée avec la carte
résultante en WTG2 et en SVG.

---

## Pourquoi combiner MCTS et cartes Wardley ?

Les cartes Wardley posent un problème classique de **décision séquentielle** : à
chaque instant, le stratège peut faire évoluer un composant ou appliquer une
manœuvre, et chaque choix ouvre un nouvel espace de possibilités. C'est exactement
le type de problème que le MCTS résout bien — il a prouvé son efficacité dans les
jeux (Go, échecs) et les problèmes de planification.

La combinaison fonctionne en trois couches :

1. **L'espace de recherche** — la carte Wardley définit les composants, leurs
   positions sur l'axe d'évolution, et les dépendances. Chaque état de la carte
   est un nœud dans l'arbre MCTS.

2. **Les actions** — deux types de moves stratégiques : faire évoluer un composant
   d'une phase (Genesis → Custom → Product → Commodity) ou appliquer un gameplay
   (open-source, ILC, strangler-fig...).

3. **L'évaluation** — un LLM reçoit la carte en format WTG2 enrichi du contexte
   stratégique complet (le fichier `skill.md` embarqué) et score la qualité
   de la stratégie entre 0 et 1.

Le MCTS en mode AlphaZero (PUCT) utilise ces scores pour guider l'exploration
vers les branches les plus prometteuses, sans rollout aléatoire.

### Analogie avec les jeux

| Concept | Morpion (alphazego) | Carte Wardley |
|---------|---------------------|---------------|
| État | Plateau 3×3 | Carte avec positions + gameplays |
| Acteur | Joueur (croix/rond) | Décideur unique |
| Move | Placer un pion | Évoluer un composant ou appliquer un gameplay |
| Victoire | Alignement de 3 | Score stratégique élevé (jugé par le LLM) |
| Évaluateur | Réseau de neurones | Gemini avec contexte Wardley |

---

## Comment ça marche

### Architecture

```
 ┌──────────┐     parse      ┌──────────────┐
 │ .wtg2    │ ──────────────→ │ WardleyState │
 │ (entrée) │                 │              │
 └──────────┘                 └──────┬───────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
              PossibleMoves()    Evaluate()        ID()
              (evolve, gameplay) (terminal?)   (SHA256 hash)
                    │
                    ▼
            ┌───────────────┐
            │  AlphaMCTS    │   sélection PUCT + expansion
            │  (mcts pkg)   │
            └───────┬───────┘
                    │ appelle Evaluator.Evaluate(state)
                    ▼
            ┌───────────────┐
            │  Evaluator    │   sérialise state → WTG2
            │  (wraps Judge)│   construit le prompt
            └───────┬───────┘
                    │ Judge.Score(prompt)
                    ▼
            ┌───────────────┐
            │  Gemini       │   score la qualité stratégique
            │  (Vertex AI)  │   retourne un score [0, 1]
            └───────────────┘
```

### Les moves

L'outil génère automatiquement tous les moves possibles à chaque état :

**EVOLVE** — avancer un composant d'une phase d'évolution :

- Genesis (I) → Custom (II) → Product (III) → Commodity (IV)
- Un composant déjà en Commodity ne peut plus évoluer
- Un composant avec de l'inertie (`!`, `!!`, `!!!`) ne peut pas évoluer

**GAMEPLAY** — appliquer une manœuvre stratégique à un composant :

| Gameplay | Effet stratégique |
|----------|-------------------|
| `open-source` | Commoditiser via l'ouverture du code pour capturer la valeur adjacente |
| `ILC` | Innovate-Leverage-Commoditize : cycle de croissance auto-entretenu |
| `land-grab` | Sacrifier la profitabilité pour capturer le marché rapidement |
| `embrace-extend` | Adopter un standard puis ajouter des extensions propriétaires |
| `tower-moat` | Ériger des barrières : brevets, lock-in, protocoles fermés |
| `strangler-fig` | Remplacer progressivement un système legacy composant par composant |

Chaque gameplay ne peut être appliqué qu'une fois par composant.

### L'évaluation par le LLM

Chaque appel au LLM inclut :

1. Le **contexte stratégique complet** (`skill.md` — 750 lignes de connaissance
   Wardley : syntaxe WTG2, gameplays, signaux climatiques, alignement EVT,
   violations de doctrine, checklist stratégique)
2. La **carte WTG2 courante** avec toutes les modifications appliquées
3. La **question stratégique** issue du fichier d'entrée

Le LLM produit deux types de scores :

- **Policy** : score de chaque move possible → guide l'exploration vers les
  branches prometteuses
- **Value** : score de l'état courant → estime la qualité globale de la séquence

La conversion de score pour le MCTS : `value_mcts = score_judge × 2 - 1`
(Judge [0,1] → MCTS [-1,1]).

### Flux d'exécution

L'exploration procède step par step (approche greedy) :

1. Parser le fichier `.wtg2` → état initial
2. Pour chaque step (jusqu'à `maxDepth`) :
   - Lancer N itérations MCTS depuis l'état courant
   - Le MCTS explore l'arbre des moves possibles
   - À chaque nœud, l'Evaluator appelle Gemini pour scorer
   - Sélectionner le meilleur move (nœud le plus visité)
3. Afficher la séquence complète + carte résultante

---

## Prérequis

- **Go 1.22+**
- **Un projet Google Cloud** avec l'API Vertex AI activée
- **Authentification GCP** :
  ```bash
  gcloud auth application-default login
  ```
- **Une carte Wardley** au format `.wtg2`

---

## Installation

```bash
cd exp/wardley
go build -o wardley-explore ./cmd/explore/
```

---

## Utilisation

```
wardley-explore -input <fichier.wtg2> -project <projet-gcp> [options]
```

### Options

| Option | Défaut | Description |
|--------|--------|-------------|
| `-input` | *(requis)* | Fichier `.wtg2` d'entrée |
| `-project` | *(requis)* | Projet GCP pour Vertex AI |
| `-region` | `us-central1` | Région GCP |
| `-depth` | `5` | Nombre max de moves stratégiques par branche |
| `-iterations` | `100` | Nombre d'itérations MCTS par step |
| `-cpuct` | `1.4` | Constante d'exploration PUCT |
| `-output-wtg2` | *(stdout)* | Fichier de sortie WTG2 |
| `-output-svg` | *(aucun)* | Fichier de sortie SVG |
| `-model` | `gemini-3-flash` | Modèle Gemini pour l'évaluation |

### Premier essai (rapide)

```bash
wardley-explore \
  -input ma-carte.wtg2 \
  -depth 3 \
  -iterations 20 \
  -project mon-projet-gcp
```

3 moves, 20 itérations MCTS par step. Résultat en quelques minutes.

### Exploration approfondie avec rendu SVG

```bash
wardley-explore \
  -input ma-carte.wtg2 \
  -depth 5 \
  -iterations 200 \
  -project mon-projet-gcp \
  -output-wtg2 result.wtg2 \
  -output-svg result.svg
```

Ouvrir `result.svg` dans un navigateur pour la vue interactive (zoom, pan).

### Mode exploratoire (cpuct élevé)

```bash
wardley-explore \
  -input ma-carte.wtg2 \
  -depth 4 \
  -iterations 100 \
  -cpuct 2.5 \
  -project mon-projet-gcp
```

Un `cpuct` plus élevé pousse le MCTS à explorer des branches moins visitées
plutôt que d'exploiter les branches connues. Utile pour découvrir des stratégies
inattendues.

---

## Écrire sa carte d'entrée

La carte doit être au format WTG2. Deux champs sont particulièrement importants
pour l'exploration :

- **`question:`** — la question stratégique que l'outil va essayer de résoudre.
  Plus elle est précise, meilleurs sont les résultats.
- **L'inertie (`!`, `!!`, `!!!`)** — bloque l'évolution des composants concernés,
  contraignant l'espace de recherche de manière réaliste.

Exemple minimal :

```wtg2
title: Mon service cloud
question: "Faut-il commoditiser notre moteur de calcul ou investir dans l'IA ?"
stages: Genesis, Custom-Built, Product, Commodity

anchor Client : IV.5

Moteur de calcul : II.7 !!(tech) >> III.5 {
  type: build
}

Module IA : I.3 {
  type: build
}

API : III.5 (build)
Cloud : IV.8 (outsource)

Client -> API -> Moteur de calcul -> Cloud
API -> Module IA -> Cloud

gameplay open-source on API
```

Les gameplays déjà présents dans la carte sont pris en compte comme état initial —
ils ne seront pas re-proposés par le MCTS.

---

## Comprendre la sortie

```
=== Exploration stratégique ===
Carte : Mon service cloud
Question : Faut-il commoditiser notre moteur de calcul ou investir dans l'IA ?
Composants : 5
Profondeur max : 3 | Itérations/step : 50 | CPUCT : 1.4

--- Step 1 : exploration (50 itérations)...
  → EVOLVE "Moteur de calcul"
--- Step 2 : exploration (50 itérations)...
  → GAMEPLAY "ILC" sur "Module IA"
--- Step 3 : exploration (50 itérations)...
  → EVOLVE "Module IA"

--- Séquence de moves ---
  1. EVOLVE "Moteur de calcul"
  2. GAMEPLAY "ILC" sur "Module IA"
  3. EVOLVE "Module IA"

--- Carte résultante (WTG2) ---
title: Mon service cloud
question: "Faut-il commoditiser notre moteur de calcul ou investir dans l'IA ?"
stages: Genesis, Custom-Built, Product, Commodity

[...]

gameplay ILC on Module IA
```

**Lecture des moves :**

- **`EVOLVE "X"`** — le composant X avance d'une phase. Le MCTS recommande
  d'investir dans la maturation de ce composant.
- **`GAMEPLAY "Y" sur "X"`** — la manœuvre stratégique Y est appliquée au
  composant X. Le MCTS estime que cette manœuvre améliore la position
  stratégique globale.

---

## Paramétrage

### Profondeur (`-depth`)

Le nombre de moves dans la séquence. Chaque step est un choix stratégique.

| Profondeur | Usage |
|------------|-------|
| 2-3 | Prototype rapide, validation du setup |
| 4-5 | Exploration standard |
| 6-10 | Stratégie long terme (coûteux en appels API) |

### Itérations (`-iterations`)

Le nombre de simulations MCTS par step. Plus d'itérations = meilleure qualité
de décision, mais plus d'appels LLM.

| Itérations | Qualité | Appels LLM par step |
|------------|---------|---------------------|
| 10-20 | Approximative | ~(N+1) × 20 |
| 50-100 | Bonne | ~(N+1) × 100 |
| 200+ | Excellente | ~(N+1) × 200 |

Où N = nombre de moves possibles (dépend du nombre de composants et des
gameplays déjà appliqués).

### CPUCT (`-cpuct`)

Contrôle le ratio exploration / exploitation dans la formule PUCT :

| CPUCT | Comportement |
|-------|-------------|
| 0.5-1.0 | Conservateur — exploite les branches connues |
| 1.0-2.0 | Équilibré (défaut : 1.4) |
| 2.0-5.0 | Explorateur — teste des branches moins visitées |

### Coût API

Chaque itération MCTS fait ~(N+1) appels au LLM : N pour la policy (scorer
chaque move possible) et 1 pour la value (scorer l'état courant).

Estimation pour une carte à 10 composants, 6 gameplays, profondeur 5, 100 itérations :

- Moves possibles par état : ~10 (évolutions) + ~60 (gameplays) = ~70
- Appels par step : ~71 × 100 = ~7100
- Total pour 5 steps : ~35 500 appels

Le modèle par défaut est `gemini-3-flash` (rapide et économique).
Il est configurable via `-model` :

```bash
# Modèle par défaut (rapide, économique)
wardley-explore -input carte.wtg2 -project p -model gemini-3-flash

# Modèle plus capable (meilleur jugement, plus lent et coûteux)
wardley-explore -input carte.wtg2 -project p -model gemini-3.1-pro-preview
```

Commencer avec `-depth 3 -iterations 20` pour estimer le coût.

---

## Structure du code

```
exp/wardley/
├── doc.go              Package documentation
├── state.go            WardleyState implémentant decision.State
├── state_test.go       Tests + Example functions
├── parse.go            Parser .wtg2 → WardleyState
├── parse_test.go       Tests avec une carte WTG2 complète
├── serialize.go        WardleyState → texte WTG2
├── serialize_test.go   Tests + Example functions
├── evaluator.go        Evaluator wrapping un Judge (→ mcts.Evaluator)
├── evaluator_test.go   Tests avec mock Judge
├── prompt.go           Prompts Wardley avec skill.md embarqué
├── skill.md            Contexte stratégique complet (go:embed)
├── render.go           WardleyState → wardleyToGo.Map → SVG
├── render_test.go
├── go.mod              Module séparé (alphazego + wardleyToGo + genai)
└── cmd/explore/
    └── main.go         CLI wardley-explore
```

### Dépendances

| Package | Source | Rôle |
|---------|--------|------|
| `decision.State` | alphazego | Interface état pour MCTS |
| `mcts.NewAlphaMCTS` | alphazego | Moteur MCTS mode AlphaZero |
| `reasoning.Judge` | alphazego | Interface Judge (réutilisée) |
| `parser/wtg2` | wardleyToGo | Parser de fichiers `.wtg2` |
| `encoding/svg` | wardleyToGo | Rendu SVG |
| `components/wardley` | wardleyToGo | Types de composants Wardley |
| `google.golang.org/genai` | Vertex AI SDK | Client Gemini |

---

## Utilisation comme bibliothèque

Le package `wardley` peut aussi être utilisé programmatiquement, sans la CLI :

```go
package main

import (
    "context"
    "fmt"
    "strings"

    "github.com/owulveryck/alphazego/exp/wardley"
    "github.com/owulveryck/alphazego/mcts"
)

func main() {
    ctx := context.Background()

    // Parser une carte
    wtg2 := `
title: Exemple
question: "Que faire ?"
anchor User : IV.5
App : III.5 (build)
DB : II.3 (buy)
User -> App -> DB
`
    state, _ := wardley.ParseWTG2(strings.NewReader(wtg2), 3)

    // Créer un Judge (ici un mock, en vrai : Vertex AI)
    judge := monJuge{}
    eval := wardley.NewEvaluator(ctx, judge)

    // Lancer MCTS
    engine := mcts.NewAlphaMCTS(eval, 1.4)
    best := engine.RunMCTS(state, 50)

    // Afficher le résultat
    ws := best.(*wardley.State)
    fmt.Println(ws.LastMove())
    fmt.Println(wardley.SerializeWTG2(ws))
}
```

---

## Limitations et pistes d'évolution

**Limitations actuelles :**

- L'inertie bloque totalement l'évolution — pas de modèle de coût graduel
- Les gameplays n'ont pas d'effet mécanique sur l'évolution (seul le LLM
  interprète leur impact)
- L'approche greedy (step par step) ne garantit pas l'optimalité globale
- Le coût API peut être élevé sur de grandes cartes

**Pistes d'évolution :**

- Mode adversarial (deux acteurs concurrents sur la même carte)
- Effets mécaniques des gameplays (open-source → accélère l'évolution)
- Cache des scores LLM pour les états déjà évalués
- Support d'autres providers LLM (Ollama pour le local)
- Intégration du rendu WTG2 complet (signaux, groupes, annotations)
