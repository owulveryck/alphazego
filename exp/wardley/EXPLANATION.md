# Explication complète du fonctionnement

Ce document détaille le fonctionnement interne de `wardley-explore` : comment une carte Wardley est transformée en problème de décision, comment MCTS explore l'espace des stratégies, et comment un LLM guide cette exploration.

## Table des matières

1. [Vue d'ensemble](#vue-densemble)
2. [Modélisation : carte Wardley → problème de décision](#modélisation--carte-wardley--problème-de-décision)
3. [Génération des moves possibles](#génération-des-moves-possibles)
4. [Le moteur MCTS](#le-moteur-mcts)
5. [L'évaluateur : le rôle du LLM](#lévaluateur--le-rôle-du-llm)
6. [Les prompts envoyés au LLM](#les-prompts-envoyés-au-llm)
7. [La boucle greedy step-by-step](#la-boucle-greedy-step-by-step)
8. [Flux de données complet](#flux-de-données-complet)
9. [Formules mathématiques](#formules-mathématiques)
10. [Paramètres de l'outil](#paramètres-de-loutil)
11. [Coût et performance](#coût-et-performance)

---

## Vue d'ensemble

Le système combine trois briques :

- **Une carte Wardley** (fichier `.wtg2`) qui décrit un paysage stratégique : des composants positionnés sur un axe d'évolution (Genesis → Commodity) et reliés par une chaîne de valeur.
- **MCTS (Monte Carlo Tree Search)** dans sa variante AlphaZero (PUCT), qui explore systématiquement l'arbre des séquences de décisions stratégiques possibles.
- **Un LLM (Gemini via Vertex AI)** qui remplace le réseau de neurones d'AlphaZero pour évaluer la qualité de chaque état stratégique.

Le résultat est une séquence ordonnée de décisions stratégiques (évoluer un composant, appliquer un gameplay) optimisée pour répondre à la question stratégique définie dans la carte.

---

## Modélisation : carte Wardley → problème de décision

### L'interface `decision.State`

Le framework alphazego définit une interface générique pour tout problème de décision :

```go
type State interface {
    CurrentActor() ActorID
    PreviousActor() ActorID
    Evaluate() ActorID          // Undecided, Player, Stalemate
    PossibleMoves() []State     // états enfants
    ID() []byte                 // identifiant unique
}
```

La carte Wardley est modélisée comme un **puzzle mono-acteur** (comme le taquin) : un seul décideur (`Player`) explore séquentiellement les moves possibles. Il n'y a pas d'adversaire — `CurrentActor()` et `PreviousActor()` retournent toujours `Player`.

### Le `State` Wardley

L'état (`exp/wardley/state.go`) contient :

| Champ | Type | Description |
|-------|------|-------------|
| `title` | `string` | Titre de la carte |
| `question` | `string` | Question stratégique à résoudre |
| `components` | `[]Component` | Liste des composants avec phase, visibilité, type, inertie, gameplays |
| `edges` | `[]Edge` | Relations de dépendance entre composants |
| `history` | `[]Move` | Séquence de moves appliqués depuis l'état initial |
| `maxDepth` | `int` | Profondeur maximale (condition d'arrêt) |

Chaque `Component` possède :
- **Name** : identifiant unique du composant
- **Phase** : position sur l'axe d'évolution (Genesis=0, Custom=1, Product=2, Commodity=3)
- **Visibility** : position sur l'axe de la chaîne de valeur (0-100, du moins visible au plus visible)
- **Type** : stratégie d'acquisition (build, buy, outsource)
- **Inertia** : résistance au changement (0 = aucune, 1-3 = bloque l'évolution)
- **Gameplays** : liste des manœuvres stratégiques appliquées à ce composant

### Condition terminale

L'état est terminal quand `len(history) >= maxDepth`. Le système retourne `Stalemate` (pas de gagnant/perdant dans un puzzle). L'exploration s'arrête et la séquence de moves accumulée constitue la stratégie proposée.

### Identité des états

`ID()` calcule un SHA-256 sur la question + les composants triés par nom (avec leurs phases et gameplays). Deux états avec les mêmes composants dans les mêmes phases et avec les mêmes gameplays ont le même ID, quel que soit l'ordre des moves qui y ont mené. Cela permet au MCTS de détecter les transpositions.

---

## Génération des moves possibles

`PossibleMoves()` génère tous les états enfants accessibles depuis l'état courant. Deux types de moves existent :

### 1. Evolve (évolution d'un composant)

Un composant peut avancer d'une phase sur l'axe d'évolution :

```
Genesis (I) → Custom (II) → Product (III) → Commodity (IV)
```

**Conditions** :
- Le composant n'est pas déjà en phase Commodity (phase < 3)
- Le composant n'a pas d'inertie (Inertia == 0)

L'inertie représente la résistance organisationnelle au changement. Un composant avec `!!` (inertie 2) ou `!!!` (inertie 3) dans le fichier WTG2 ne peut pas évoluer — il est bloqué.

### 2. ApplyGameplay (application d'une manœuvre stratégique)

Un gameplay peut être appliqué à n'importe quel composant, à condition qu'il ne soit pas déjà appliqué à ce composant. Les six gameplays disponibles sont :

| Gameplay | Description stratégique |
|----------|------------------------|
| `open-source` | Commoditiser via l'ouverture du code source |
| `ILC` | Innovate-Leverage-Commoditize : cycle d'innovation |
| `land-grab` | Capturer rapidement le marché |
| `embrace-extend` | Adopter un standard puis l'étendre |
| `tower-moat` | Construire des barrières à l'entrée |
| `strangler-fig` | Remplacer progressivement un composant legacy |

### Copie d'état

Chaque move crée une **copie complète** de l'état. Les `Component` sont des structs (copiés par valeur), mais les slices internes (`Gameplays`) sont explicitement dupliquées pour éviter les mutations partagées. L'historique est étendu avec le nouveau move.

### Taille de l'espace

Pour une carte avec $N$ composants dont $N_e$ évoluables et $G = 6$ gameplays disponibles, le branching factor est :

$$b = N_e + \sum_{i=1}^{N} \bigl(G - |\text{gameplays}_i|\bigr)$$

Exemple concret : 10 composants, 5 évoluables, aucun gameplay appliqué → $b = 5 + 10 \times 6 = 65$ moves possibles par état. Avec une profondeur de 5, l'arbre complet contiendrait $\sim 65^5 \approx 1{,}2 \times 10^9$ feuilles. C'est pourquoi MCTS est nécessaire : il échantillonne intelligemment cet arbre au lieu de l'explorer exhaustivement.

---

## Le moteur MCTS

### Principe général

MCTS (Monte Carlo Tree Search) construit progressivement un arbre de recherche en répétant quatre phases :

```
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│  1. SÉLECTION          2. EXPANSION       3. ÉVALUATION          │
│                                                                  │
│  Descendre l'arbre     Créer un           Demander au LLM       │
│  en choisissant        nouveau nœud       d'évaluer l'état       │
│  les enfants les       pour un move       → policy + value       │
│  plus prometteurs      non exploré                               │
│  (PUCT)                                                          │
│                                                                  │
│  4. RÉTROPROPAGATION                                             │
│                                                                  │
│  Remonter la value                                               │
│  du nouveau nœud                                                 │
│  jusqu'à la racine                                               │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### Phase 1 : Sélection (PUCT)

Depuis la racine, on descend dans l'arbre en choisissant à chaque nœud l'enfant qui maximise la formule **PUCT** (Predictor + Upper Confidence bound applied to Trees) :

$$\text{PUCT}(s, a) = Q(s, a) + C_{\text{puct}} \cdot P(s, a) \cdot \frac{\sqrt{N(\text{parent})}}{1 + N(s, a)}$$

Où :
- $Q(s,a)$ : valeur moyenne de l'action $a$ depuis l'état $s$ (moyenne des values rétropropagées)
- $C_{\text{puct}}$ : constante d'exploration (paramètre `-cpuct`, défaut 1.4)
- $P(s,a)$ : prior de policy — probabilité attribuée par le LLM à ce move
- $N(\text{parent})$ : nombre total de visites du nœud parent
- $N(s,a)$ : nombre de visites de cet enfant

Cette formule équilibre **exploitation** ($Q$ élevé = move historiquement bon) et **exploration** ($N$ faible = move peu visité, $P$ élevé = move jugé prometteur par le LLM). Le terme $C_{\text{puct}}$ contrôle cet équilibre :
- $C_{\text{puct}}$ élevé (ex: 2.5) → favorise l'exploration de moves peu visités
- $C_{\text{puct}}$ faible (ex: 0.5) → favorise l'exploitation des moves connus

La descente s'arrête quand on atteint un nœud non complètement développé (il reste des moves non explorés) ou un nœud terminal.

### Phase 2 : Expansion

Quand on atteint un nœud non complètement développé, on crée un nouveau nœud enfant pour un move non encore exploré.

En mode **AlphaMCTS** (celui utilisé ici), la première expansion d'un nœud crée **tous les enfants d'un coup** (`expandAll`) et leur attribue les priors de policy fournis par le LLM. C'est différent du MCTS classique qui crée les enfants un par un.

```go
func expandAll(node *mctsNode, priors []float64) []*mctsNode {
    moves := node.state.PossibleMoves()
    for i, move := range moves {
        child := &mctsNode{
            state:  move,
            parent: node,
            prior:  priors[i],  // ← prior du LLM
        }
        node.children = append(node.children, child)
    }
    return node.children
}
```

Les priors influencent directement la formule PUCT : un move avec un prior élevé (le LLM le juge prometteur) sera exploré plus tôt, même sans visite préalable.

### Phase 3 : Évaluation (détaillée dans la section suivante)

L'évaluateur (wrappant le LLM) est appelé pour obtenir :
- **policy** : un vecteur de probabilités sur les moves possibles
- **value** : une estimation scalaire de la qualité de l'état courant

### Phase 4 : Rétropropagation

La value obtenue remonte de l'enfant nouvellement créé jusqu'à la racine. À chaque nœud traversé :

```go
func backpropagateValue(node *mctsNode, values map[ActorID]float64) {
    for n := node; n != nil; n = n.parent {
        n.visits++
        actor := n.state.CurrentActor()
        n.value += values[actor]  // accumule la value pour cet acteur
    }
}
```

La valeur $Q(s,a)$ utilisée dans PUCT est $\frac{\text{node.value}}{\text{node.visits}}$ — la moyenne des values accumulées. Plus un sous-arbre est visité et contient des états bien évalués par le LLM, plus son $Q$ est élevé.

### Choix du meilleur move

Après toutes les itérations, le move retenu est celui dont l'enfant a le **plus grand nombre de visites** (pas la meilleure valeur moyenne). Le nombre de visites est plus robuste car il intègre à la fois la qualité ($Q$) et l'exploration (les bons moves attirent naturellement plus de visites via PUCT).

```go
func selectBestMove(root *mctsNode) decision.State {
    best := root.children[0]
    for _, child := range root.children[1:] {
        if child.visits > best.visits {
            best = child
        }
    }
    return best.state
}
```

---

## L'évaluateur : le rôle du LLM

L'évaluateur (`exp/wardley/evaluator.go`) est le pont entre le MCTS et le LLM. Il implémente l'interface `mcts.Evaluator` :

```go
type Evaluator interface {
    Evaluate(state decision.State) (policy []float64, values map[ActorID]float64)
}
```

### Calcul de la policy — mode batch (par défaut)

Si le Judge implémente l'interface `BatchScorer`, la policy est calculée en **un seul appel LLM** :

1. Sérialiser la carte courante en WTG2
2. Lister les $N$ moves candidats avec descriptions enrichies (ex: `EVOLVE "DB" (Custom → Product)`)
3. Construire un prompt batch contenant la carte + la liste numérotée des moves
4. Appeler `ScoreBatch(ctx, prompt, N)` → tableau de $N$ scores
5. Normaliser les scores pour qu'ils somment à 1

L'interface `BatchScorer` est définie dans le package wardley :

```go
type BatchScorer interface {
    ScoreBatch(ctx context.Context, prompt string, count int) ([]float64, error)
}
```

Le LLM retourne un tableau JSON (ex: `[0.7, 0.3, 0.8, ...]`) parsé automatiquement.

### Calcul de la policy — mode individuel (fallback)

Si le Judge n'implémente pas `BatchScorer`, le fallback évalue chaque move séparément :

1. Pour chaque move : appliquer → sérialiser l'état enfant → prompt individuel
2. Appeler `Judge.Score(ctx, prompt)` → score $\in [0, 1]$
3. Minimum $10^{-8}$ pour éviter les zéros

Ce mode fait $N$ appels LLM au lieu de 1.

### Normalisation

Dans les deux modes, les scores sont normalisés :

$$P(s, a_i) = \frac{\max(\text{score}_i,\; 10^{-8})}{\sum_j \max(\text{score}_j,\; 10^{-8})}$$

### Calcul de la value

L'état courant (pas les enfants) est évalué séparément :

1. Sérialiser l'état courant en texte WTG2
2. Construire un prompt de value (cf. section prompts) incluant l'historique des moves
3. Envoyer au LLM → score $\in [0, 1]$
4. Convertir en échelle MCTS : $v_{\text{MCTS}} = 2s - 1$ → valeur $\in [-1, 1]$

La conversion est nécessaire car le MCTS travaille dans $[-1, 1]$ (convention issue des jeux : $-1$ = défaite, $+1$ = victoire) tandis que le LLM retourne un score de qualité dans $[0, 1]$.

### Nombre d'appels LLM par évaluation

| Mode | Appels par Evaluate() | Avec 65 moves |
|------|----------------------|---------------|
| **Batch** (défaut) | $2$ (1 policy + 1 value) | $2$ |
| **Individuel** (fallback) | $N + 1$ | $66$ |

Le MCTS appelle `Evaluate()` une fois par expansion de nœud. En mode batch avec 100 itérations, ça fait au maximum $100 \times 2 = 200$ appels LLM par step, contre $100 \times 66 = 6\,600$ en mode individuel — une réduction de $\sim 33\times$ pour un branching factor de 65.

En réalité, `Evaluate()` n'est appelé que lors de la première visite d'un nœud. Les visites suivantes réutilisent les priors et values déjà calculés.

---

## Les prompts envoyés au LLM

### Contexte stratégique embarqué

Chaque prompt commence par le contenu intégral de `skill.md` (~750 lignes), embarqué dans le binaire via `go:embed`. Ce fichier contient toute la connaissance stratégique Wardley : syntaxe WTG2, concepts de climat/doctrine/manœuvre, description détaillée des gameplays, signaux climatiques, EVT (Evolution-Value-Team alignment), violations de doctrine, et la Strategic Completeness Checklist.

Ce contexte permet au LLM de raisonner comme un expert en stratégie Wardley, pas seulement comme un modèle de langage généraliste.

### Prompt de policy (batch)

Objectif : scorer tous les moves candidats en un seul appel LLM. Le prompt envoie la carte courante et la liste des moves, au lieu de sérialiser chaque état enfant séparément.

```
<contenu intégral de skill.md>

---

Tu es un expert en stratégie Wardley. Évalue chaque move candidat.

Carte Wardley actuelle :
<texte WTG2 de l'état courant>

Question stratégique : <question>

Moves candidats :
  1. EVOLVE "Component A" (Custom → Product)
  2. GAMEPLAY "open-source" sur "Component B"
  3. EVOLVE "Component C" (Genesis → Custom)
  ...

Pour chaque move, évalue sa pertinence stratégique sur [0, 1].
Considère : cohérence de la chaîne de valeur, doctrine, EVT, inertie.

Réponds uniquement par un tableau JSON de scores, ex: [0.7, 0.3, 0.8]
```

Le LLM retourne un tableau JSON de $N$ scores. Le parsing extrait le premier `[...]` trouvé dans la réponse et le décode via `json.Unmarshal`.

### Prompt de policy (individuel, fallback)

Utilisé quand le Judge n'implémente pas `BatchScorer`. Chaque état enfant est évalué séparément :

```
<contenu intégral de skill.md>

---

Tu es un expert en stratégie Wardley. Évalue la carte suivante.

Carte Wardley actuelle :
<texte WTG2 de l'état enfant>

Question stratégique : <question de la carte>

Évalue la qualité stratégique de cette carte sur une échelle de 0 à 1.
Utilise la Strategic Completeness Checklist et les concepts de
climat/doctrine/manœuvre pour ton évaluation. Considère :
- Cohérence de la chaîne de valeur
- Pertinence des gameplays par rapport au contexte
- Alignement EVT (teams vs phases d'évolution)
- Violations de doctrine (NIH, dispersion, strategy theatre...)
- Gestion de l'inertie et des signaux climatiques
```

Le LLM retourne un seul nombre décimal entre 0 et 1.

### Prompt de value

Objectif : estimer la progression de la séquence stratégique en cours.

```
<contenu intégral de skill.md>

---

Tu es un expert en stratégie Wardley. Évalue la progression de cette
séquence stratégique.

Carte Wardley :
<texte WTG2 de l'état courant>

Question stratégique : <question>
Moves effectués :
  1. <description du move 1>
  2. <description du move 2>
  ...

Estime la probabilité que cette séquence stratégique mène à une bonne
réponse à la question. Utilise le framework OODA et le Value Flywheel
pour juger la progression.
0 = stratégie incohérente ou contre-productive
1 = stratégie optimale répondant clairement à la question
```

La différence clé : le prompt de value inclut l'**historique des moves**, permettant au LLM de juger la cohérence de la séquence dans son ensemble, pas seulement l'état résultant.

### Extraction du score

La réponse du LLM (texte libre) est parsée pour en extraire un nombre :

```go
func parseScore(text string) (float64, error) {
    for _, word := range splitWords(text) {
        if _, err := fmt.Sscanf(word, "%f", &score); err == nil {
            return clamp(score), nil
        }
    }
    return 0, fmt.Errorf("aucun nombre trouvé")
}
```

Le premier nombre trouvé dans la réponse est utilisé, borné entre 0 et 1 par `clamp()`. Si le parsing échoue, un score par défaut de 0.5 (neutre) est retourné.

---

## La boucle greedy step-by-step

Le CLI n'exécute **pas** un seul MCTS profond. Il utilise une approche **greedy itérative** :

```
Pour chaque step de 1 à maxDepth :
    1. Vérifier si l'état courant est terminal → si oui, arrêter
    2. Lancer RunMCTS(état_courant, nb_itérations) → meilleur état enfant
    3. L'état enfant devient l'état courant
    4. Afficher le move effectué
```

Cette approche est imposée par l'architecture : `RunMCTS()` retourne le meilleur état à un pas de profondeur (l'enfant de la racine avec le plus de visites). Les nœuds internes de l'arbre MCTS ne sont pas accessibles depuis l'extérieur.

**Conséquence** : chaque step reconstruit un arbre MCTS depuis zéro. L'arbre du step précédent est perdu. C'est sous-optimal par rapport à un MCTS qui réutiliserait le sous-arbre, mais c'est plus simple et suffisant pour le cas d'usage.

**Avantage** : l'indicateur de progression peut être réinitialisé à chaque step, donnant un retour clair à l'utilisateur sur l'avancement.

### Progression

L'évaluateur expose un callback `Progress` appelé après chaque appel LLM :

```go
eval.Progress = func(info ProgressInfo) {
    fmt.Fprintf(os.Stderr,
        "\r  iter %d/%d | %d appels LLM | policy: %d/%d",
        info.EvalCount, iterations, info.LLMCalls,
        info.PolicyScored, info.PolicyTotal)
}
```

Le `\r` (retour chariot) écrase la ligne précédente sur stderr, créant un compteur animé. Deux formats alternent :
- Pendant le calcul de la policy : `policy: 3/15` (3 moves scorés sur 15)
- Pendant le calcul de la value : `value: +0.72` (score de l'état)

---

## Flux de données complet

```
Fichier .wtg2
     │
     │ wardleyToGo/parser/wtg2.NewParser().Parse()
     │ → Document AST (titre, question, nodes, edges, gameplays)
     ▼
ParseWTG2() ─── exp/wardley/parse.go
     │
     │ Extraction : positions → Phase, types, inertie, gameplays
     ▼
State initial ─── exp/wardley/state.go
     │
     │ Boucle greedy (maxDepth steps)
     ▼
┌────────────────────────────────────────────────────┐
│  RunMCTS(state, iterations)                        │
│                                                    │
│  Pour chaque itération :                           │
│                                                    │
│  1. Sélection : descente PUCT dans l'arbre         │
│     └─ PUCT(s,a) = Q + C·P·√N_parent/(1+N)        │
│                                                    │
│  2. Expansion : expandAll (si première visite)      │
│     └─ Appel Evaluator.Evaluate(state)             │
│        │                                           │
│        │  ┌──────────────────────────────────┐      │
│        │  │  Pour chaque move possible :     │      │
│        │  │    SerializeWTG2(enfant)          │      │
│        │  │    formatPolicyPrompt(wtg2, q)    │      │
│        │  │    Judge.Score(prompt) → [0,1]    │      │
│        │  │  Normaliser → policy              │      │
│        │  │                                  │      │
│        │  │  Pour l'état courant :            │      │
│        │  │    SerializeWTG2(état)             │      │
│        │  │    formatValuePrompt(wtg2, q, h)  │      │
│        │  │    Judge.Score(prompt) → [0,1]    │      │
│        │  │    Conversion → [-1,1]            │      │
│        │  └──────────────────────────────────┘      │
│        │                                           │
│        ▼                                           │
│  3. Rétropropagation : value remonte à la racine   │
│                                                    │
│  Fin : selectBestMove (enfant le plus visité)      │
└────────────────────────────────────────────────────┘
     │
     │ État enfant sélectionné → nouvel état courant
     │ Répéter pour le step suivant
     ▼
State final (avec history complète)
     │
     ├─── SerializeWTG2(state) → texte WTG2 (stdout ou fichier)
     │
     └─── RenderSVG(writer, state) → fichier SVG
               │
               │ Reconstruction wardleyToGo.Map :
               │   Component → wardley.Component (position, type, gameplays)
               │   Edge → wardley.Collaboration
               │
               │ encoding/svg.NewEncoder → Encode → Close
               ▼
          Fichier SVG (visualisation de la carte résultante)
```

---

## Formules mathématiques

### PUCT (sélection)

À chaque nœud de l'arbre, le MCTS choisit l'enfant qui maximise le score PUCT :

$$\text{PUCT}(s, a) = Q(s, a) + C_{\text{puct}} \cdot P(s, a) \cdot \frac{\sqrt{N(\text{parent})}}{1 + N(s, a)}$$

Où :

- $Q(s, a) = \dfrac{\sum \text{values}}{N(s,a)}$ — valeur moyenne de l'action $a$ depuis l'état $s$
- $P(s, a) \in [0, 1]$ — prior de policy, normalisé, fourni par le LLM
- $N(\text{parent})$ — nombre total de visites du nœud parent
- $N(s, a)$ — nombre de visites de l'enfant correspondant à l'action $a$
- $C_{\text{puct}} \in \mathbb{R}^+$ — constante d'exploration (paramètre `-cpuct`, défaut 1.4)

Le premier terme $Q(s,a)$ favorise l'**exploitation** : les moves historiquement bons. Le second terme favorise l'**exploration** : les moves peu visités ($N(s,a)$ petit) et/ou jugés prometteurs par le LLM ($P(s,a)$ élevé). Le ratio $\frac{\sqrt{N(\text{parent})}}{1 + N(s,a)}$ décroît au fur et à mesure que l'enfant est visité, réduisant progressivement le bonus d'exploration.

### Normalisation de la policy

Les scores bruts du LLM sont transformés en distribution de probabilité :

$$P(s, a_i) = \frac{\max\bigl(\text{score}_{\text{LLM}}(a_i),\; 10^{-8}\bigr)}{\displaystyle\sum_{j} \max\bigl(\text{score}_{\text{LLM}}(a_j),\; 10^{-8}\bigr)}$$

Le plancher $10^{-8}$ empêche un move d'avoir un prior exactement nul, ce qui le rendrait inaccessible au MCTS. Après normalisation, $\sum_i P(s, a_i) = 1$.

### Conversion value

Le LLM retourne un score de qualité dans $[0, 1]$. Le MCTS travaille dans $[-1, 1]$ (convention des jeux : $-1$ = défaite, $+1$ = victoire). La conversion est linéaire :

$$v_{\text{MCTS}} = 2 \cdot s_{\text{LLM}} - 1$$

| $s_{\text{LLM}}$ | $v_{\text{MCTS}}$ | Interprétation |
|:-:|:-:|:--|
| $0.0$ | $-1.0$ | Stratégie très mauvaise |
| $0.5$ | $\phantom{-}0.0$ | Neutre |
| $1.0$ | $+1.0$ | Stratégie excellente |

### Identité d'état

L'identifiant unique d'un état est calculé par hachage cryptographique :

$$\text{ID} = \text{SHA-256}\!\left(\text{question} \;\|\; \bigsqcup_{\text{composants triés par nom}} \bigl(\text{nom} \;\|\; \text{phase} \;\|\; \text{gameplays triés}\bigr)\right)$$

Deux états avec les mêmes composants, phases et gameplays ont le même ID, quel que soit l'ordre des moves qui y ont mené.

### Estimation du branching factor

Pour une carte à $N$ composants dont $N_e$ sont évoluables (phase $< 3$, inertie $= 0$) et $G = 6$ gameplays disponibles :

$$b = N_e + \sum_{i=1}^{N} \bigl(G - |\text{gameplays}_i|\bigr)$$

où $|\text{gameplays}_i|$ est le nombre de gameplays déjà appliqués au composant $i$.

---

## Paramètres de l'outil

Tous les paramètres sont passés en ligne de commande via des flags. Cette section détaille leur rôle, leur impact sur le comportement du système, et comment les ajuster.

### `-input` (requis)

Chemin vers le fichier WTG2 d'entrée. C'est la carte Wardley à analyser.

Le fichier doit contenir au minimum un `title:` et des composants. Le champ `question:` est fortement recommandé — c'est la question stratégique que l'outil tente de résoudre. Sans question, les prompts envoyés au LLM manquent de direction et les scores sont moins discriminants.

### `-depth` (défaut : 5)

Nombre maximal de moves stratégiques dans la séquence résultante. C'est la profondeur de la boucle greedy : l'outil exécute au maximum `depth` steps, chacun ajoutant un move à l'historique.

**Impact sur le comportement :**

$$\text{longueur de la séquence} \leq \text{depth}$$

- **Depth faible (2-3)** : séquences courtes, résultats rapides, stratégies tactiques à court terme.
- **Depth moyen (4-6)** : bon compromis entre profondeur stratégique et coût. Recommandé pour la plupart des cartes.
- **Depth élevé (7+)** : séquences longues, mais les moves tardifs sont souvent moins pertinents car le LLM perd de la discrimination. Le coût est proportionnel.

L'exploration s'arrête aussi si l'état devient terminal (tous les composants en Commodity sans moves restants), mais cela arrive rarement avant d'atteindre `maxDepth`.

**Coût** : le coût total est multiplié par `depth` (un MCTS complet par step).

### `-iterations` (défaut : 100)

Nombre d'itérations du MCTS pour chaque step. C'est le paramètre qui contrôle la **qualité de la décision** à chaque étape.

À chaque itération, le MCTS :
1. Descend dans l'arbre (sélection PUCT)
2. Développe un nœud non visité (expansion + évaluation LLM)
3. Rétropropage la value

**Impact sur le comportement :**

- **Peu d'itérations (10-30)** : exploration superficielle. Le MCTS a à peine le temps de visiter les enfants directs. Les décisions sont fortement guidées par les priors du LLM (policy). C'est rapide mais les moves choisis peuvent être sous-optimaux.
- **Itérations modérées (50-200)** : le MCTS explore suffisamment pour que la valeur Q influence la sélection au-delà des priors. Bon compromis.
- **Beaucoup d'itérations (500+)** : exploration profonde de l'arbre. Les Q convergent vers leurs vraies valeurs. Coûteux en appels LLM.

**Relation avec le branching factor :**

$$\text{couverture} = \frac{\text{iterations}}{b}$$

Si le branching factor $b = 65$ et `iterations = 20`, chaque enfant direct est visité en moyenne $\frac{20}{65} \approx 0.3$ fois — la plupart ne sont même pas visités. Augmenter les itérations améliore la couverture.

**Coût** : en mode batch, le nombre d'appels LLM par step est au maximum $\text{iterations} \times 2$, indépendant du branching factor.

### `-cpuct` (défaut : 1.4)

Constante d'exploration dans la formule PUCT. C'est le paramètre qui contrôle l'**équilibre exploration/exploitation**.

$$\text{PUCT}(s, a) = \underbrace{Q(s, a)}_{\text{exploitation}} + \underbrace{C_{\text{puct}} \cdot P(s, a) \cdot \frac{\sqrt{N(\text{parent})}}{1 + N(s, a)}}_{\text{exploration}}$$

**Impact sur le comportement :**

- **$C_{\text{puct}}$ faible (0.5-1.0)** : le MCTS exploite fortement. Il concentre les visites sur les moves déjà identifiés comme bons (Q élevé). Risque de convergence prématurée vers un optimum local.
- **$C_{\text{puct}}$ standard (1.0-2.0)** : équilibre classique. La valeur 1.4 est issue de la littérature AlphaZero.
- **$C_{\text{puct}}$ élevé (2.0-4.0)** : le MCTS explore largement. Il visite des moves même si leur Q est faible, pourvu que leur prior P soit raisonnable ou qu'ils soient peu visités. Utile pour des cartes complexes où l'espace stratégique est riche.

**Quand ajuster :**

| Situation | Recommandation |
|-----------|---------------|
| Carte simple, peu de composants | Baisser à 1.0 (convergence rapide) |
| Carte complexe, nombreux gameplays | Monter à 2.0-2.5 (exploration large) |
| Les résultats semblent toujours identiques | Monter (trop d'exploitation) |
| Les résultats semblent aléatoires | Baisser (trop d'exploration) |

**Coût** : aucun impact direct sur le nombre d'appels LLM. Seule la qualité/diversité des résultats change.

### `-model` (défaut : `gemini-3-flash`)

Identifiant du modèle Gemini utilisé via Vertex AI pour évaluer les états stratégiques. Ce modèle est appelé des centaines à des milliers de fois par exécution.

**Impact sur le comportement :**

Le modèle influence la qualité des scores (policy et value). Un modèle plus puissant produit des évaluations stratégiques plus fines, ce qui donne de meilleurs priors au MCTS et des values plus discriminantes.

**Modèles recommandés :**

| Modèle | Vitesse | Qualité | Coût/appel | Usage |
|--------|---------|---------|------------|-------|
| `gemini-3.1-flash-lite` | Très rapide | Correcte | Très faible | Prototypage rapide |
| `gemini-3-flash` | Rapide | Bonne | Faible | Bon compromis quotidien (défaut) |
| `gemini-2.5-flash` | Rapide | Très bonne | Modéré | Analyse sérieuse |
| `gemini-3-pro` | Lent | Excellente | Élevé | Analyse approfondie, peu d'itérations |

Avec un modèle lent et coûteux, réduire `-iterations` pour compenser : moins d'appels mais de meilleure qualité.

**Configuration LLM fixée dans le code :**

Le modèle est toujours appelé avec :
- `Temperature: 0.0` — reproductibilité maximale des scores
- `ThinkingBudget: 256` tokens — raisonnement court, adapté au volume d'appels

Ces paramètres ne sont pas modifiables en ligne de commande.

### `-project` (requis)

Identifiant du projet Google Cloud Platform. L'API Vertex AI doit être activée sur ce projet, et l'authentification doit être configurée (`gcloud auth application-default login`).

### `-region` (défaut : `us-central1`)

Région GCP pour les appels Vertex AI. Choisir une région proche pour réduire la latence. Toutes les régions ne supportent pas tous les modèles.

### `-output-wtg2` (défaut : stdout)

Chemin du fichier de sortie pour la carte résultante en format WTG2. Si omis, le WTG2 est affiché sur stdout (mélangé avec les logs de progression). Spécifier un fichier permet de séparer proprement la sortie structurée des logs.

### `-output-svg` (optionnel)

Chemin du fichier de sortie SVG. Si spécifié, l'outil génère un rendu visuel de la carte résultante (après application de tous les moves). Le SVG utilise le moteur de rendu de wardleyToGo et peut être ouvert dans un navigateur.

### Interactions entre paramètres

Les paramètres ne sont pas indépendants. Voici les interactions importantes :

**`depth` × `iterations` = coût total :**

$$\text{appels LLM} \leq \text{depth} \times \text{iterations} \times 2 \quad \text{(mode batch)}$$

En mode batch, le coût est indépendant du branching factor. Doubler `depth` ou `iterations` double le coût.

**`iterations` vs `cpuct` :**

Avec peu d'itérations et un $C_{\text{puct}}$ élevé, le MCTS éparpille ses visites sans converger. Avec beaucoup d'itérations et un $C_{\text{puct}}$ faible, il surexploite les premiers moves trouvés. Les deux doivent être cohérents :

| Itérations | $C_{\text{puct}}$ recommandé |
|:----------:|:----------------------------:|
| 10-30      | 0.8 - 1.2                   |
| 50-200     | 1.2 - 2.0                   |
| 500+       | 1.4 - 2.5                   |

**`model` vs `iterations` :**

Un modèle puissant mais lent (ex: `gemini-2.5-pro`) donne de bons priors dès les premières itérations. On peut alors réduire `iterations` (ex: 20-50) car le MCTS n'a pas besoin de beaucoup corriger les priors. À l'inverse, un modèle rapide mais moins précis (ex: `flash-lite`) bénéficie de plus d'itérations pour compenser la qualité des priors.

### Profils d'utilisation recommandés

**Exploration rapide** (~60 appels LLM en batch) :
```bash
-depth 2 -iterations 15 -cpuct 1.2 -model gemini-3-flash
```

**Analyse standard** (~400 appels LLM en batch) :
```bash
-depth 4 -iterations 50 -cpuct 1.4 -model gemini-2.0-flash
```

**Analyse approfondie** (~300 appels LLM en batch, modèle puissant) :
```bash
-depth 5 -iterations 30 -cpuct 1.6 -model gemini-2.5-pro
```

**Exploration large** (~1 200 appels LLM en batch) :
```bash
-depth 3 -iterations 200 -cpuct 2.5 -model gemini-2.0-flash
```

---

## Coût et performance

### Appels LLM par exécution

En mode **batch** (par défaut quand le Judge implémente `BatchScorer`), chaque expansion coûte 2 appels LLM (1 batch policy + 1 value), indépendamment du branching factor :

$$\text{appels par step} \leq \text{iterations} \times 2$$

$$\text{appels totaux} \leq \text{depth} \times \text{iterations} \times 2$$

**Estimation pratique** pour `-depth 3 -iterations 20` :

$$\text{appels/step} \leq 20 \times 2 = 40$$

$$\text{appels totaux} \leq 3 \times 40 = 120$$

En mode **individuel** (fallback), le coût dépend du branching factor $b$ :

$$\text{appels totaux} \leq \text{depth} \times \text{iterations} \times (b + 1)$$

Pour $b = 65$, les mêmes paramètres donnent $3 \times 20 \times 66 = 3\,960$ appels — soit $\sim 33\times$ plus.

En pratique, les chiffres réels sont inférieurs car le MCTS revisite des nœuds déjà évalués sans appeler l'évaluateur.

### Température du LLM

Le LLM est appelé avec `Temperature: 0.0` pour maximiser la reproductibilité des scores. Un prompt identique devrait retourner un score identique (ou très proche). Le MCTS lui-même est stochastique (via la sélection PUCT et l'ordre d'exploration), donc deux exécutions donneront des résultats légèrement différents même avec température 0.

### Budget de réflexion

- **Mode individuel** : `ThinkingBudget: 256` tokens — raisonnement court, adapté au grand nombre d'appels.
- **Mode batch** : `ThinkingBudget: 512` tokens — budget plus large car le LLM doit évaluer tous les moves en une seule passe.
