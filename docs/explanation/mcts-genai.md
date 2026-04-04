# Coupler MCTS et IA générative : une exploration expérimentale

> **Statut** : expérimental. Le répertoire `exp/benchmark/` explore une piste
> de recherche. Les résultats sont préliminaires et l'approche est amenée
> à évoluer.

## Le constat

Un LLM appelé en une seule fois (*one-shot*) peut produire des réponses
impressionnantes, mais il commet régulièrement des erreurs de raisonnement
sur des problèmes combinatoires. Il oublie une contrainte, calcule mal un
chemin critique, ou s'engage dans un raisonnement sous-optimal sans
possibilité de retour en arrière.

Plus le modèle est petit (7B, flash-lite), plus ces erreurs sont
fréquentes. Les gros modèles (pro, 70B+) s'en sortent mieux, mais
coûtent plus cher et sont plus lents.

## L'hypothèse

**La recherche arborescente (MCTS) peut compenser les faiblesses de
raisonnement des petits modèles** en explorant systématiquement plusieurs
chemins de décision, là où un appel unique peut s'engager dans une voie
sans retour.

Concrètement :

- Un petit modèle + MCTS devrait faire **significativement mieux** que
  le même petit modèle seul (B >> A)
- Un petit modèle + MCTS pourrait rivaliser avec un gros modèle seul
  (B ~= C)

## Le mécanisme

Le couplage réutilise le framework `decision/reasoning` qui modélise le
raisonnement comme un problème de décision séquentielle compatible avec
le moteur MCTS :

```
                    [Question]
                   /     |     \
          [Étape A]  [Étape B]  [Étape C]    ← Generator produit 3 candidats
           /    \       |
    [Suite A1] [A2]  [Suite B1]               ← exploration MCTS
       |
  [CONCLUSION]                                ← meilleure solution trouvée
```

Les rôles sont distribués ainsi :

| Composant MCTS     | Implémentation LLM                              |
|---------------------|--------------------------------------------------|
| **State**           | Contexte de raisonnement (question + étapes)     |
| **PossibleMoves**   | Le Generator produit N étapes candidates          |
| **Evaluator**       | Le Judge score la qualité de chaque chemin        |
| **Terminal**         | Une étape commençant par `CONCLUSION:`           |

À chaque niveau de l'arbre, le MCTS :

1. Sélectionne le noeud le plus prometteur (PUCT)
2. Demande au Generator N nouvelles étapes de raisonnement
3. Demande au Judge d'évaluer chaque candidat (policy) et l'état courant (value)
4. Rétropropage les scores pour affiner la sélection future

Le MCTS équilibre ainsi **exploration** (essayer des approches nouvelles)
et **exploitation** (approfondir les pistes prometteuses), ce que le
one-shot ne peut pas faire.

## Le benchmark

Pour tester l'hypothèse, le benchmark utilise des problèmes
d'**ordonnancement de tâches** (package `exp/benchmark/problems`). Ce domaine
est bien adapté car :

- La solution optimale est calculable (chemin critique)
- Le scoring est objectif (dépendances respectées ? makespan optimal ?)
- La difficulté est graduable (4 à 12 tâches)

### Configurations testées

**Vertex AI** (`exp/benchmark/vertexai`) — 4 configurations :

| Config | Modèle           | Méthode  |
|--------|-------------------|----------|
| A      | flash-lite (petit)| One-shot |
| B      | flash-lite (petit)| MCTS     |
| C      | pro (gros)        | One-shot |
| D      | pro (gros)        | MCTS     |

**Ollama** (`exp/benchmark/ollama`) — 2 configurations :

| Config | Modèle            | Méthode  |
|--------|-------------------|----------|
| E      | local (ex: 7B)    | One-shot |
| F      | local (ex: 7B)    | MCTS     |

### Évaluation : LLM-as-Judge

Chaque solution est évaluée par un modèle juge qui vérifie les dépendances
et le makespan :

- **1.0** : dépendances respectées ET makespan optimal
- **0.5** : dépendances respectées MAIS makespan non optimal
- **0.0** : dépendances violées ou réponse incompréhensible

## Premiers résultats observés

Sur le problème « Rénovation appartement » (10 tâches, optimal 18 jours) :

- **A (flash-lite seul)** : score 0.0 — viole la dépendance Carrelage→Plomberie
- **B (flash-lite + MCTS)** : score 1.0 — trouve le makespan optimal de 18 jours
- **C (pro seul)** : score 1.0
- **D (pro + MCTS)** : score 1.0

Le MCTS a permis au petit modèle de passer de 0% à 100% sur ce problème.

## Limites

- **Coût** : le MCTS multiplie les appels LLM (typiquement ×50 à ×100
  par rapport au one-shot). Le gain en qualité doit justifier le surcoût.
- **Juge local** : quand le même petit modèle sert de juge (Ollama),
  la fiabilité du scoring est incertaine.
- **Domaine étroit** : le benchmark ne couvre que l'ordonnancement.
  L'hypothèse reste à valider sur d'autres types de raisonnement.
- **Taille d'échantillon** : les résultats sont sur un petit nombre
  d'exécutions, sans répétitions statistiques.

## Lien avec le framework

Le couplage MCTS + LLM est une application directe du framework
`decision` :

- `decision/reasoning` fournit le `State` et les interfaces `Generator`/`Judge`
- `mcts` fournit le moteur AlphaMCTS avec PUCT
- `exp/benchmark/vertexai` et `exp/benchmark/ollama` implémentent les interfaces
  pour des providers LLM concrets
- `exp/benchmark/problems` définit les problèmes d'évaluation

L'architecture est volontairement découplée : ajouter un nouveau provider
(OpenAI, Anthropic, etc.) revient à implémenter `Generator` et `Judge`
dans un nouveau module.
