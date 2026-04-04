// Package problems définit les problèmes d'ordonnancement de tâches
// utilisés par les benchmarks MCTS + IA générative.
//
// Chaque problème modélise un projet concret (construction, déploiement
// logiciel, pipeline data, etc.) sous forme de tâches avec des durées et
// des contraintes de précédence. L'objectif est de trouver l'ordonnancement
// qui minimise le temps total d'exécution (makespan), en respectant toutes
// les dépendances et en exploitant le parallélisme possible.
//
// # Domaine : ordonnancement sous contraintes
//
// L'ordonnancement de tâches est un problème classique de recherche
// opérationnelle. Étant donné un ensemble de tâches avec des durées et
// des relations de précédence (« A doit finir avant que B commence »),
// il s'agit de déterminer l'ordre d'exécution optimal.
//
// La solution optimale correspond au chemin critique : le plus long chemin
// dans le graphe de dépendances. Les tâches hors du chemin critique peuvent
// s'exécuter en parallèle sans affecter le makespan.
//
// Exemple (construction maison) :
//
//	Fondations (3j) ──→ Murs (5j) ──→ Toiture (2j)
//	                               ├─→ Électricité (3j) ──→ Finitions (2j)
//	                               └─→ Plomberie (2j) ────↗
//
//	Chemin critique : Fondations → Murs → Électricité → Finitions = 13 jours
//
// # Types
//
// [Task] représente une tâche individuelle avec un nom, une durée en jours,
// et une liste de dépendances (noms des tâches prérequises).
//
// [Problem] regroupe un ensemble de tâches et le makespan optimal calculé
// manuellement. La méthode [Problem.FormatPrompt] génère une description
// en langage naturel adaptée à un LLM.
//
// # Les 10 problèmes
//
// Le benchmark comprend 10 problèmes de difficulté croissante :
//
//	| #  | Problème               | Tâches | Optimal | Difficulté   |
//	|----|------------------------|--------|---------|--------------|
//	| 1  | Chaîne linéaire        | 4      | 8j      | Facile       |
//	| 2  | Fourche parallèle      | 5      | 6j      | Facile       |
//	| 3  | Diamant                | 5      | 10j     | Facile       |
//	| 4  | Construction maison    | 6      | 13j     | Moyen        |
//	| 5  | Déploiement logiciel   | 7      | 10j     | Moyen        |
//	| 6  | Organisation événement | 7      | 6j      | Moyen        |
//	| 7  | Projet web fullstack   | 8      | 12j     | Difficile    |
//	| 8  | Pipeline data ETL      | 9      | 12j     | Difficile    |
//	| 9  | Rénovation appartement | 10     | 18j     | Difficile    |
//	| 10 | Lancement produit      | 12     | 25j     | Très diff.   |
//
// Les problèmes faciles (1-3) ont des structures simples (chaîne, fourche,
// diamant). Les problèmes moyens (4-6) introduisent des dépendances multiples.
// Les problèmes difficiles (7-10) combinent branches parallèles, dépendances
// croisées et un grand nombre de tâches.
//
// # Pourquoi ces problèmes
//
// L'ordonnancement est un bon terrain d'évaluation pour le couplage
// MCTS + LLM car :
//
//   - Le raisonnement one-shot est fragile : un petit modèle peut oublier
//     une dépendance, mal calculer le chemin critique, ou ne pas voir
//     que certaines tâches sont parallélisables.
//   - Le MCTS peut explorer différentes stratégies de planification et
//     identifier la meilleure via l'évaluation.
//   - La solution optimale est vérifiable algorithmiquement (chemin critique),
//     ce qui permet un scoring objectif.
//
// # Utilisation
//
//	probs := problems.All()
//	for _, p := range probs {
//	    fmt.Printf("%s (%d tâches, optimal=%d jours)\n",
//	        p.Name, len(p.Tasks), p.Optimal)
//	    prompt := p.FormatPrompt()
//	    // envoyer prompt à un LLM...
//	}
package problems
