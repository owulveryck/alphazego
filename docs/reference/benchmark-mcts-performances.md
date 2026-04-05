# Comparatif de performances BenchmarkRunMCTS_10000

## Contexte

Historique complet des performances de `BenchmarkRunMCTS_10000` (TicTacToe, MCTS pur)
depuis son introduction en v1.3.3 jusqu'à HEAD. Toutes les mesures ont été prises
sur la même machine (arm64, 8 cores), `go test -bench=BenchmarkRunMCTS_10000 -benchmem -count=3`.

## Tableau comparatif

| Commit | Version | Description | ns/op | B/op | allocs/op |
|--------|---------|-------------|------:|-----:|----------:|
| 55dbba3 | v1.3.3 | Baseline (introduction du benchmark) | 62 290 399 | 12 408 979 | 267 498 |
| db5d47f | v1.3.4 | Cache PossibleMoves sur les noeuds | 61 003 172 | 9 460 780 | 208 901 |
| f976bf4 | — | Expansion par index (suppression doublons) | 52 767 116 | 8 521 030 | 140 392 |
| 68e6ade | — | Audit perf round 1 (log cache, batch TicTacToe) | 33 264 438 | 6 022 402 | 69 933 |
| f520978 | — | RandomMover (simulate O(1)) | 21 418 416 | 3 174 313 | 53 710 |
| 5014598 | HEAD | Audit perf round 2 (inlining, batch alloc, cache) | 20 179 333 | 3 343 051 | 44 508 |

## Progression par commit (delta vs commit précédent)

| Commit | Description | Temps | Mémoire | Allocs |
|--------|-------------|------:|--------:|-------:|
| db5d47f | Cache PossibleMoves | -2.1% | -23.8% | -21.9% |
| f976bf4 | Expansion par index | -13.5% | -9.9% | -32.8% |
| 68e6ade | Audit perf round 1 | -36.9% | -29.3% | -50.2% |
| f520978 | RandomMover | -35.6% | -47.3% | -23.2% |
| 5014598 | Audit perf round 2 | -5.8% | +5.3% | -17.1% |

## Bilan global (v1.3.3 → HEAD)

| Métrique | v1.3.3 (baseline) | HEAD | Amélioration | Facteur |
|----------|-------------------:|-----:|-----------:|--------:|
| Temps (ns/op) | 62 290 399 | 20 179 333 | **-67.6%** | **3.1x** |
| Mémoire (B/op) | 12 408 979 | 3 343 051 | **-73.0%** | **3.7x** |
| Allocations | 267 498 | 44 508 | **-83.4%** | **6.0x** |

## Détail des optimisations par commit

### v1.3.3 → v1.3.4 : Cache PossibleMoves

- `getPossibleMoves()` met en cache le résultat de `PossibleMoves()` sur le noeud
- Évite les allocations répétées dans `isFullyExpanded()` et `expand()`
- Impact principal sur la mémoire (-23.8%) et les allocations (-21.9%)

### v1.3.4 → f976bf4 : Expansion par index

- Remplace la détection de doublons (boucle O(n) sur `children`) par un index incrémental
- `expandedIndex` sert de curseur dans `cachedMoves`
- Réduction massive des allocations (-32.8%)

### f976bf4 → 68e6ade : Audit perf round 1 (plus gros impact)

- Cache `logVisits` sur le noeud parent pour éviter `math.Log()` dans la boucle de sélection
- Suppression de l'inventory map dans `selectChildUCB`
- Batch allocation dans le TicTacToe (réduction allocations rollout)
- **Plus grosse amélioration en temps absolu : -19.5 ms/op**

### 68e6ade → f520978 : RandomMover

- Interface `RandomMover` pour génération O(1) d'un successeur aléatoire
- `simulate()` utilise `RandomMove(rng)` au lieu de `PossibleMoves()[rng(n)]`
- Élimine l'allocation du slice intermédiaire à chaque pas du rollout
- **Plus grosse amélioration en mémoire : -47.3%**

### f520978 → 5014598 : Audit perf round 2

- Inlining direct de `ucb1()` et `puct()` (suppression du pointeur de fonction `selectionFn`)
- Batch allocation des `mctsNode` par blocs de 256 (`allocNode()`)
- Cache `previousActor` sur le noeud pour éviter l'appel d'interface dans `backpropagate()`
- `logVisits` mis à jour incrémentalement dans `backpropagate()`
- La mémoire augmente légèrement (+5.3%) à cause du padding des batches de 256 noeuds

## Données brutes

### 55dbba3 (v1.3.3)

```
BenchmarkRunMCTS_10000-8   28   63263044 ns/op   12377251 B/op   266509 allocs/op
BenchmarkRunMCTS_10000-8   18   65812444 ns/op   12470679 B/op   268482 allocs/op
BenchmarkRunMCTS_10000-8   19   62795708 ns/op   12379008 B/op   266504 allocs/op
```

### db5d47f (v1.3.4)

```
BenchmarkRunMCTS_10000-8   19   61031316 ns/op   9477473 B/op   209200 allocs/op
BenchmarkRunMCTS_10000-8   24   61655465 ns/op   9463688 B/op   208990 allocs/op
BenchmarkRunMCTS_10000-8   18   60322735 ns/op   9441179 B/op   208512 allocs/op
```

### f976bf4

```
BenchmarkRunMCTS_10000-8   24   55035000 ns/op   8525363 B/op   140441 allocs/op
BenchmarkRunMCTS_10000-8   22   53578094 ns/op   8544661 B/op   140842 allocs/op
BenchmarkRunMCTS_10000-8   24   49688253 ns/op   8493066 B/op   139894 allocs/op
```

### 68e6ade

```
BenchmarkRunMCTS_10000-8   33   35494788 ns/op   6014073 B/op   69835 allocs/op
BenchmarkRunMCTS_10000-8   31   33539811 ns/op   6043973 B/op   70160 allocs/op
BenchmarkRunMCTS_10000-8   50   30758714 ns/op   6009159 B/op   69803 allocs/op
```

### f520978

```
BenchmarkRunMCTS_10000-8   54   22745497 ns/op   3173169 B/op   53701 allocs/op
BenchmarkRunMCTS_10000-8   64   20264870 ns/op   3173066 B/op   53679 allocs/op
BenchmarkRunMCTS_10000-8   72   21244880 ns/op   3176704 B/op   53751 allocs/op
```

### 5014598 (HEAD)

```
BenchmarkRunMCTS_10000-8   51   20004368 ns/op   3339242 B/op   44462 allocs/op
BenchmarkRunMCTS_10000-8   62   19836506 ns/op   3342339 B/op   44492 allocs/op
BenchmarkRunMCTS_10000-8   50   20697126 ns/op   3347571 B/op   44571 allocs/op
```
