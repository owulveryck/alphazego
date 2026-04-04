# Les problèmes du benchmark : ordonnancement de tâches

## Pourquoi l'ordonnancement ?

Pour tester si le MCTS améliore le raisonnement d'un LLM (voir [MCTS + IA générative](mcts-genai.md)), il faut un domaine qui réunit trois propriétés :

1. **Solution vérifiable** — on peut calculer le chemin critique et vérifier objectivement si la réponse est correcte
2. **Difficulté graduable** — en ajoutant des tâches et des dépendances, on augmente la complexité combinatoire
3. **Compréhensible par un LLM** — l'ordonnancement de projet est un sujet que les modèles de langage ont vu en quantité dans leurs données d'entraînement

L'ordonnancement de tâches sous contraintes de dépendances coche ces trois cases. Chaque problème se résume à : *étant donné un ensemble de tâches avec des durées et des dépendances, trouver un planning qui minimise le temps total (makespan).*

## Concepts clés

### Le makespan

Le makespan est le temps total d'exécution du projet, de la première tâche commencée à la dernière terminée. C'est la métrique à minimiser.

```
Temps →  0  1  2  3  4  5  6  7  8
         ┌──A──┐
                ┌────B────┐
                           ┌─C┐
                              ┌──D──┐
         makespan = 8
```

### Le chemin critique

Le chemin critique est la plus longue chaîne de dépendances. Il détermine le makespan optimal — aucun planning ne peut faire mieux, même avec un parallélisme infini.

```
         ┌──A──┐
                ┌────B────┐
                           ┌─────C─────┐   ← chemin critique : A→B→C = 10
         ┌───D───┐
                  ┌──E──┐               ← D→E = 5 (non critique)

         makespan optimal = 10
```

### Les dépendances

Une tâche ne peut commencer que lorsque **toutes** ses dépendances sont terminées. Deux tâches sans dépendance mutuelle peuvent s'exécuter en parallèle.

## Les 10 problèmes

Les problèmes sont ordonnés par difficulté croissante, de 4 à 12 tâches.

### 1. Chaîne linéaire (4 tâches, optimal = 8)

La forme la plus simple : une séquence A → B → C → D, sans parallélisme possible.

```
A(2) → B(3) → C(1) → D(2)
makespan = 2 + 3 + 1 + 2 = 8
```

**Intérêt** : vérifier que le modèle comprend les dépendances séquentielles et sait additionner les durées.

### 2. Fourche parallèle (5 tâches, optimal = 6)

Un point d'entrée (Init), trois branches parallèles, puis une fusion.

```
          ┌─ Branche1(3) ─┐
Init(1) ──┼─ Branche2(2) ─┼── Fusion(1)
          └─ Branche3(4) ─┘
```

Chemin critique : Init → Branche3 → Fusion = 1 + 4 + 1 = 6.

**Intérêt** : tester la compréhension du parallélisme. Le LLM doit identifier que le goulot d'étranglement est Branche3, pas la somme des trois branches.

### 3. Diamant (5 tâches, optimal = 10)

Deux chemins divergent puis convergent.

```
Début(2) ──┬── Gauche(3) ──┬── Jonction(2) ── Fin(1)
            └── Droite(5) ──┘
```

Chemin critique : Début → Droite → Jonction → Fin = 2 + 5 + 2 + 1 = 10.

**Intérêt** : même structure de fourche mais avec un point de convergence. Le modèle doit comprendre que Jonction attend **les deux** branches, pas seulement la plus rapide.

### 4. Construction maison (6 tâches, optimal = 13)

Un problème réaliste avec des métiers qui dépendent de la structure.

```
Fondations(3) → Murs(5) ──┬── Toiture(2) ──────┐
                           ├── Électricité(3) ───┼── Finitions(2)
                           └── Plomberie(2) ─────┘
```

Chemin critique : Fondations → Murs → Électricité → Finitions = 3 + 5 + 3 + 2 = 13.

**Intérêt** : introduction de noms réalistes qui ancrent le raisonnement dans un domaine concret. Le modèle pourrait utiliser ses connaissances du domaine de la construction.

### 5. Déploiement logiciel (7 tâches, optimal = 10)

Un pipeline CI/CD avec des étapes parallèles et des portes de validation.

```
Compilation(2) ──┬── Tests unitaires(3) ───┐
                 ├── Tests intégration(4) ──┼── Build image(1) ──┐
                 └── Revue de code(2) ──────┴── Déploiement staging(2) ── Déploiement prod(1)
```

Chemin critique : Compilation → Tests intégration → Build image → Staging → Prod = 2 + 4 + 1 + 2 + 1 = 10.

**Intérêt** : graphe de dépendances plus complexe avec des nœuds à dépendances multiples. Build image attend les tests unitaires **et** d'intégration ; le staging attend Build image **et** la revue de code.

### 6. Organisation événement (7 tâches, optimal = 6)

Deux chaînes indépendantes qui convergent.

```
Réserver salle(1) ──┬── Invitations(1) ─────────────────────┐
                    └── Décoration(3) ── Installer sono(1) ──┼── Accueillir invités(1)
Choisir traiteur(2) ── Préparer menu(2) ────────────────────┘
```

Chemin critique : Réserver salle → Décoration → Sono → Accueil = 1 + 3 + 1 + 1 = 6.

**Intérêt** : deux sous-projets indépendants (salle et traiteur) démarrent en parallèle. Le modèle doit identifier les deux chaînes et trouver la plus longue.

### 7. Projet web fullstack (8 tâches, optimal = 12)

Un projet logiciel complet avec frontend, backend et infrastructure.

```
Specs(2) ──┬── Design UI(3) ── Frontend(4) ──────┐
           ├── Setup infra(2) ── Base de données(3) ──┼── Intégration(3) ── Recette(2)
           └── Backend API(5) ───────────────────┘
```

Chemin critique : Specs → Backend API → Intégration → Recette = 2 + 5 + 3 + 2 = 12.

**Intérêt** : trois branches parallèles de longueurs différentes avec un point d'intégration. Le modèle doit gérer 8 tâches et identifier que le backend, pas le frontend ni l'infra, est le chemin critique.

### 8. Pipeline data ETL (9 tâches, optimal = 12)

Trois sources de données extraites, nettoyées, puis fusionnées.

```
Extraction A(2) ── Nettoyage A(2) ──┐
Extraction B(3) ── Nettoyage B(3) ──┼── Fusion(2) ── Agrégation(3) ── Rapport(1)
Extraction C(1) ── Nettoyage C(1) ──┘
```

Chemin critique : Extraction B → Nettoyage B → Fusion → Agrégation → Rapport = 3 + 3 + 2 + 3 + 1 = 12.

**Intérêt** : structure symétrique (3 pipelines parallèles) suivie d'un goulot d'étranglement. Le modèle doit comprendre que les trois sources sont indépendantes mais que la fusion attend les trois.

### 9. Rénovation appartement (10 tâches, optimal = 18)

Un projet de rénovation avec des corps de métier interdépendants.

```
Démolition(3) ──┬── Évacuation gravats(1) ──────────────────┐
                ├── Plomberie(4) ──┬── Carrelage(2) ────────┼── Nettoyage final(1)
                └── Électricité(3) ┼── Isolation(2) ── Plâtre(3) ── Peinture(2) ── Menuiserie(3) ──┘
                                   └───────────────┘
```

Chemin critique : Démolition → Plomberie → Isolation → Plâtre → Peinture → Menuiserie → Nettoyage = 3 + 4 + 2 + 3 + 2 + 3 + 1 = 18.

**Intérêt** : graphe de dépendances dense avec des nœuds ayant 2-3 prédécesseurs. Isolation dépend de Plomberie **et** Électricité ; le Nettoyage final dépend de Menuiserie **et** Carrelage. C'est sur ce problème que le benchmark a montré les premiers résultats positifs (flash-lite passe de 0.0 à 1.0 avec MCTS).

### 10. Lancement produit (12 tâches, optimal = 25)

Le problème le plus complexe : un lancement de produit complet.

```
Étude marché(3) ──┬── Prototype(5) ── Tests utilisateurs(3) ── Design final(2) ──┬── Développement(6) ── Tests QA(3) ──┐
                  │                                                                ├── Rédaction doc(2) ────────────────┼── Formation(2) ── Lancement(1)
                  ├── Revue légale(3) ────────────────────────────────────────────────────────────────────────────────────┤
                  └── Stratégie marketing(2) ── Création contenu(4) ─────────────────────────────────────────────────────┘
```

Chemin critique : Étude → Prototype → Tests → Design → Développement → QA → Formation → Lancement = 3 + 5 + 3 + 2 + 6 + 3 + 2 + 1 = 25.

**Intérêt** : 12 tâches avec un graphe de dépendances ramifié et profond. Plusieurs chemins quasi-critiques coexistent. Le modèle doit gérer la complexité combinatoire sans s'y perdre — c'est le test le plus difficile du benchmark.

## Progression de la difficulté

| # | Problème | Tâches | Dépendances | Optimal | Difficulté |
|---|----------|--------|-------------|---------|------------|
| 1 | Chaîne linéaire | 4 | 3 | 8 | Facile |
| 2 | Fourche parallèle | 5 | 4 | 6 | Facile |
| 3 | Diamant | 5 | 4 | 10 | Facile |
| 4 | Construction maison | 6 | 5 | 13 | Moyen |
| 5 | Déploiement logiciel | 7 | 8 | 10 | Moyen |
| 6 | Organisation événement | 7 | 7 | 6 | Moyen |
| 7 | Projet web fullstack | 8 | 8 | 12 | Difficile |
| 8 | Pipeline data ETL | 9 | 6 | 12 | Difficile |
| 9 | Rénovation appartement | 10 | 9 | 18 | Difficile |
| 10 | Lancement produit | 12 | 14 | 25 | Très difficile |

La difficulté ne dépend pas seulement du nombre de tâches : la densité du graphe de dépendances, le nombre de chemins quasi-critiques et la présence de « pièges » (branches courtes qui semblent critiques) jouent un rôle important.

## Pourquoi ces problèmes spécifiques ?

Chaque problème a été choisi pour tester un aspect particulier du raisonnement :

- **Problèmes 1-3** : bases (séquence, parallélisme, convergence). Un modèle qui échoue ici ne comprend pas les dépendances.
- **Problèmes 4-6** : complexité intermédiaire avec des noms réalistes. Le modèle peut-il mobiliser ses connaissances du domaine (construction, CI/CD, événementiel) ?
- **Problèmes 7-9** : graphes denses. Le modèle doit tenir en mémoire de travail 8-10 tâches et leurs interactions.
- **Problème 10** : test de saturation. 12 tâches avec un chemin critique de 8 étapes — au-delà de ce que la plupart des petits modèles gèrent en one-shot.

## Lien avec le benchmark

Le package `benchmark/problems` expose ces 10 problèmes via la fonction `All()`. Chaque problème fournit `FormatPrompt()` pour générer automatiquement la description en langage naturel envoyée au LLM. Le makespan `Optimal` sert de référence pour le scoring par le juge.
