# Comment ajouter un nouveau problème au benchmark

Ce guide explique comment ajouter un problème d'ordonnancement au package `exp/benchmark/problems` pour l'inclure dans les benchmarks MCTS + LLM.

## Prérequis

- Connaissance de base de Go
- Avoir lu la [description des problèmes existants](../explanation/problemes-benchmark.md)

## 1. Concevoir le problème

Un bon problème de benchmark a les propriétés suivantes :

- **Chemin critique calculable** — vous devez pouvoir déterminer le makespan optimal à la main
- **Au moins un piège** — une branche qui semble critique mais ne l'est pas, ou une dépendance subtile facile à oublier
- **Noms réalistes** — les tâches doivent évoquer un domaine concret pour ancrer le raisonnement du LLM

### Calculer le chemin critique

Le chemin critique est la plus longue chaîne de dépendances. Pour le trouver :

1. Dessiner le graphe de dépendances
2. Pour chaque tâche, calculer la **date de fin au plus tôt** (earliest finish) :
   - Si pas de dépendances : `EF = durée`
   - Sinon : `EF = max(EF des dépendances) + durée`
3. Le makespan optimal est le `max(EF)` de toutes les tâches

Exemple pour un pipeline simple :

```
A(2) → C(3)      EF(A) = 2, EF(C) = 2+3 = 5
B(4) → C(3)      EF(B) = 4, EF(C) = max(2,4)+3 = 7   ← chemin critique
                  Optimal = 7
```

## 2. Ajouter le problème dans `All()`

Ouvrir `exp/benchmark/problems/problems.go` et ajouter une entrée dans le slice retourné par `All()` :

```go
// 11. Description courte (N tâches) — Niveau de difficulté
{
    Name: "Nom du problème",
    Tasks: []Task{
        {"Tâche A", 3, nil},                        // pas de dépendance
        {"Tâche B", 2, []string{"Tâche A"}},        // dépend de A
        {"Tâche C", 4, []string{"Tâche A"}},        // dépend de A (parallèle à B)
        {"Tâche D", 1, []string{"Tâche B", "Tâche C"}}, // attend B et C
    },
    Optimal: 8, // chemin critique : A(3) → C(4) → D(1) = 8
},
```

### Règles à respecter

| Règle | Pourquoi |
|-------|----------|
| Les noms de tâches sont des `string` uniques | `FormatPrompt()` les affiche tels quels, et les tests vérifient l'unicité |
| Les dépendances référencent des noms existants | Le test `TestAllProblems_DependenciesExist` échouera sinon |
| Pas d'auto-dépendance | Le test `TestAllProblems_NoCyclicSelfDep` vérifie cela |
| `Optimal > 0` | Le test `TestAllProblems_OptimalPositive` vérifie cela |
| `Optimal >= durée max` d'une tâche | Le test `TestAllProblems_OptimalAtLeastSumCriticalPath` vérifie cela |
| Maintenir l'ordre de difficulté croissante | Le nombre de tâches ne doit pas diminuer fortement par rapport au problème précédent |

### Conseils pour des noms de tâches efficaces

- Utiliser des noms en français, cohérents avec les problèmes existants
- Préférer des noms courts mais évocateurs (« Fondations », pas « Couler les fondations en béton armé »)
- Les accents sont acceptés et encouragés (« Électricité », « Rénovation »)

## 3. Mettre à jour le commentaire de la fonction `All()`

Le commentaire godoc de `All()` mentionne le nombre de problèmes :

```go
// All retourne les 11 problèmes du benchmark, de difficulté croissante.
func All() []Problem {
```

Mettre à jour le nombre si nécessaire.

## 4. Mettre à jour les tests

### Ajuster le compteur dans `problems_test.go`

Le test `TestAllProblems_Count` vérifie le nombre exact de problèmes :

```go
func TestAllProblems_Count(t *testing.T) {
    probs := All()
    if got := len(probs); got != 11 { // ← mettre à jour
        t.Errorf("All() retourne %d problèmes, attendu 11", got)
    }
}
```

### Ajuster l'exemple dans `example_test.go`

L'exemple `ExampleAll` affiche le nombre de problèmes et le dernier :

```go
func ExampleAll() {
    probs := problems.All()
    fmt.Printf("%d problèmes\n", len(probs))
    fmt.Printf("Premier: %s (%d tâches, optimal=%d)\n",
        probs[0].Name, len(probs[0].Tasks), probs[0].Optimal)
    fmt.Printf("Dernier: %s (%d tâches, optimal=%d)\n",
        probs[len(probs)-1].Name, len(probs[len(probs)-1].Tasks), probs[len(probs)-1].Optimal)
    // Output:
    // 11 problèmes                                    ← mettre à jour
    // Premier: Chaîne linéaire (4 tâches, optimal=8)
    // Dernier: Nom du problème (N tâches, optimal=X)  ← mettre à jour
}
```

## 5. Vérifier

Lancer les tests depuis la racine du projet :

```bash
go test ./exp/benchmark/problems/...
```

Tous les tests existants valident automatiquement la cohérence du nouveau problème :

- `TestAllProblems_OptimalPositive` — optimal > 0
- `TestAllProblems_HasTasks` — au moins une tâche
- `TestAllProblems_DependenciesExist` — toutes les dépendances pointent vers des tâches existantes
- `TestAllProblems_NoCyclicSelfDep` — aucune tâche ne dépend d'elle-même
- `TestAllProblems_OptimalAtLeastSumCriticalPath` — optimal >= durée de la plus longue tâche
- `TestAllProblems_DifficultyIncreasing` — le nombre de tâches ne décroît pas fortement
- `TestFormatPrompt_ContainsTaskNames` — tous les noms apparaissent dans le prompt généré

## 6. Documenter le problème

Ajouter une section dans `docs/explanation/problemes-benchmark.md` décrivant :

- Le graphe de dépendances (en ASCII art)
- Le chemin critique avec le calcul
- L'intérêt du problème (quel aspect du raisonnement il teste)

Mettre à jour le tableau récapitulatif en fin de document.

## Exemple complet

Voici un exemple de problème testant les dépendances croisées entre équipes :

```go
// 11. Développement microservices (8 tâches) — Difficile
{
    Name: "Microservices",
    Tasks: []Task{
        {"Définir API contracts", 2, nil},
        {"Service auth", 4, []string{"Définir API contracts"}},
        {"Service paiement", 5, []string{"Définir API contracts"}},
        {"Service notification", 3, []string{"Définir API contracts"}},
        {"Gateway API", 2, []string{"Service auth"}},
        {"Intégration paiement-auth", 2, []string{"Service paiement", "Service auth"}},
        {"Tests E2E", 3, []string{"Gateway API", "Intégration paiement-auth", "Service notification"}},
        {"Mise en production", 1, []string{"Tests E2E"}},
    },
    Optimal: 14,
    // Chemin critique : Contracts(2) → Paiement(5) → Intégration(2) → ... non
    // Contracts(2) → Auth(4) → Intégration(2) → Tests E2E(3) → Prod(1) = 12 ? Non.
    // Contracts(2) → Paiement(5) → Intégration(2) → Tests(3) → Prod(1) = 13
    // Vérification : EF(Contracts)=2, EF(Auth)=6, EF(Paiement)=7,
    //   EF(Notif)=5, EF(Gateway)=8, EF(Intég)=max(7,6)+2=9,
    //   EF(Tests)=max(8,9,5)+3=12, EF(Prod)=13
    // → Optimal = 13, pas 14 ! Toujours vérifier son calcul.
},
```

**Attention** : comme le montre cet exemple, il est facile de se tromper sur le chemin critique. Toujours vérifier en calculant les dates de fin au plus tôt pour **chaque** tâche.
