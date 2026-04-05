package problems

import (
	"fmt"
	"strings"
)

// Task représente une tâche individuelle dans un problème d'ordonnancement.
// Une tâche a un nom unique, une durée en jours, et éventuellement des
// dépendances vers d'autres tâches (identifiées par leur nom).
//
// Une tâche sans dépendances peut commencer immédiatement ;
// une tâche avec dépendances ne peut commencer qu'après la fin
// de toutes ses dépendances.
type Task struct {
	// Name est le nom unique de la tâche dans le problème.
	Name string
	// Duration est la durée d'exécution en jours.
	Duration int
	// Dependencies est la liste des noms de tâches prérequises.
	// Nil ou vide signifie que la tâche peut commencer immédiatement.
	Dependencies []string
}

// Problem représente un problème d'ordonnancement avec sa solution de référence.
// Il contient un ensemble de [Task] liées par des contraintes de précédence,
// et le makespan optimal (durée du chemin critique) calculé manuellement.
//
// Le makespan optimal sert de ground truth pour évaluer la qualité des
// réponses d'un LLM : une réponse correcte doit trouver exactement cette valeur.
type Problem struct {
	// Name est le nom descriptif du problème (ex: "Construction maison").
	Name string
	// Tasks est l'ensemble des tâches à ordonnancer, dans l'ordre de déclaration.
	Tasks []Task
	// Optimal est le makespan optimal en jours, correspondant à la longueur
	// du chemin critique dans le graphe de dépendances.
	Optimal int
}

// FormatPrompt retourne la description du problème en langage naturel,
// prête à être envoyée à un LLM. Le prompt liste les tâches avec leurs
// durées et dépendances, puis les règles d'ordonnancement (respect des
// dépendances, parallélisme possible, objectif de minimisation du makespan).
func (p Problem) FormatPrompt() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Problème d'ordonnancement : %s\n\n", p.Name)
	b.WriteString("Tâches à planifier :\n")
	for _, t := range p.Tasks {
		deps := "aucune"
		if len(t.Dependencies) > 0 {
			deps = strings.Join(t.Dependencies, ", ")
		}
		fmt.Fprintf(&b, "  - %s (durée: %d jours, dépendances: %s)\n", t.Name, t.Duration, deps)
	}
	b.WriteString("\nRègles :\n")
	b.WriteString("  - Une tâche ne peut commencer que lorsque toutes ses dépendances sont terminées\n")
	b.WriteString("  - Les tâches sans dépendances mutuelles peuvent s'exécuter en parallèle\n")
	b.WriteString("  - L'objectif est de minimiser le temps total (makespan)\n")
	return b.String()
}

// All retourne les 10 problèmes du benchmark, ordonnés par difficulté
// croissante : facile (1-3), moyen (4-6), difficile (7-9) et très difficile (10).
// Le nombre de tâches va de 4 à 12, et le makespan optimal de 6 à 25 jours.
//
// Les problèmes faciles ont des topologies simples (chaîne, fourche, diamant).
// Les problèmes moyens et difficiles combinent branches parallèles et
// dépendances croisées, rendant le chemin critique moins évident.
func All() []Problem {
	return []Problem{
		// 1. Linéaire (4 tâches) — Facile
		{
			Name: "Chaîne linéaire",
			Tasks: []Task{
				{"A", 2, nil},
				{"B", 3, []string{"A"}},
				{"C", 1, []string{"B"}},
				{"D", 2, []string{"C"}},
			},
			Optimal: 8, // 2+3+1+2
		},
		// 2. Parallèle (5 tâches) — Facile
		{
			Name: "Fourche parallèle",
			Tasks: []Task{
				{"Init", 1, nil},
				{"Branche1", 3, []string{"Init"}},
				{"Branche2", 2, []string{"Init"}},
				{"Branche3", 4, []string{"Init"}},
				{"Fusion", 1, []string{"Branche1", "Branche2", "Branche3"}},
			},
			Optimal: 6, // 1+4+1 (chemin critique via Branche3)
		},
		// 3. Diamant (5 tâches) — Facile
		{
			Name: "Diamant",
			Tasks: []Task{
				{"Début", 2, nil},
				{"Gauche", 3, []string{"Début"}},
				{"Droite", 5, []string{"Début"}},
				{"Jonction", 2, []string{"Gauche", "Droite"}},
				{"Fin", 1, []string{"Jonction"}},
			},
			Optimal: 10, // 2+5+2+1 (chemin critique via Droite)
		},
		// 4. Construction maison (6 tâches) — Moyen
		{
			Name: "Construction maison",
			Tasks: []Task{
				{"Fondations", 3, nil},
				{"Murs", 5, []string{"Fondations"}},
				{"Toiture", 2, []string{"Murs"}},
				{"Électricité", 3, []string{"Murs"}},
				{"Plomberie", 2, []string{"Murs"}},
				{"Finitions", 2, []string{"Toiture", "Électricité", "Plomberie"}},
			},
			Optimal: 13, // 3+5+3+2 (chemin critique via Électricité)
		},
		// 5. Déploiement logiciel (7 tâches) — Moyen
		{
			Name: "Déploiement logiciel",
			Tasks: []Task{
				{"Compilation", 2, nil},
				{"Tests unitaires", 3, []string{"Compilation"}},
				{"Tests intégration", 4, []string{"Compilation"}},
				{"Revue de code", 2, []string{"Compilation"}},
				{"Build image", 1, []string{"Tests unitaires", "Tests intégration"}},
				{"Déploiement staging", 2, []string{"Build image", "Revue de code"}},
				{"Déploiement prod", 1, []string{"Déploiement staging"}},
			},
			Optimal: 10, // 2+4+1+2+1 (via Tests intégration)
		},
		// 6. Organisation événement (7 tâches) — Moyen
		{
			Name: "Organisation événement",
			Tasks: []Task{
				{"Réserver salle", 1, nil},
				{"Choisir traiteur", 2, nil},
				{"Envoyer invitations", 1, []string{"Réserver salle"}},
				{"Décoration", 3, []string{"Réserver salle"}},
				{"Préparer menu", 2, []string{"Choisir traiteur"}},
				{"Installer sono", 1, []string{"Décoration"}},
				{"Accueillir invités", 1, []string{"Envoyer invitations", "Préparer menu", "Installer sono"}},
			},
			Optimal: 6, // 1+3+1+1 (via Décoration→Sono→Accueil)
		},
		// 7. Projet web (8 tâches) — Difficile
		{
			Name: "Projet web fullstack",
			Tasks: []Task{
				{"Specs", 2, nil},
				{"Design UI", 3, []string{"Specs"}},
				{"Setup infra", 2, []string{"Specs"}},
				{"Backend API", 5, []string{"Specs"}},
				{"Frontend", 4, []string{"Design UI"}},
				{"Base de données", 3, []string{"Setup infra"}},
				{"Intégration", 3, []string{"Frontend", "Backend API", "Base de données"}},
				{"Recette", 2, []string{"Intégration"}},
			},
			Optimal: 12, // 2+5+3+2 (via Backend API→Intégration→Recette)
		},
		// 8. Pipeline data (9 tâches) — Difficile
		{
			Name: "Pipeline data ETL",
			Tasks: []Task{
				{"Extraction source A", 2, nil},
				{"Extraction source B", 3, nil},
				{"Extraction source C", 1, nil},
				{"Nettoyage A", 2, []string{"Extraction source A"}},
				{"Nettoyage B", 3, []string{"Extraction source B"}},
				{"Nettoyage C", 1, []string{"Extraction source C"}},
				{"Fusion", 2, []string{"Nettoyage A", "Nettoyage B", "Nettoyage C"}},
				{"Agrégation", 3, []string{"Fusion"}},
				{"Rapport", 1, []string{"Agrégation"}},
			},
			Optimal: 12, // 3+3+2+3+1 (via source B)
		},
		// 9. Rénovation complète (10 tâches) — Difficile
		{
			Name: "Rénovation appartement",
			Tasks: []Task{
				{"Démolition", 3, nil},
				{"Évacuation gravats", 1, []string{"Démolition"}},
				{"Plomberie", 4, []string{"Démolition"}},
				{"Électricité", 3, []string{"Démolition"}},
				{"Isolation", 2, []string{"Plomberie", "Électricité"}},
				{"Plâtre", 3, []string{"Isolation"}},
				{"Carrelage", 2, []string{"Plomberie", "Évacuation gravats"}},
				{"Peinture", 2, []string{"Plâtre"}},
				{"Menuiserie", 3, []string{"Peinture"}},
				{"Nettoyage final", 1, []string{"Menuiserie", "Carrelage"}},
			},
			Optimal: 18, // 3+4+2+3+2+3+1 (Démolition→Plomberie→Isolation→Plâtre→Peinture→Menuiserie→Nettoyage)
		},
		// 10. Lancement produit (12 tâches) — Très difficile
		{
			Name: "Lancement produit",
			Tasks: []Task{
				{"Étude marché", 3, nil},
				{"Prototype", 5, []string{"Étude marché"}},
				{"Tests utilisateurs", 3, []string{"Prototype"}},
				{"Design final", 2, []string{"Tests utilisateurs"}},
				{"Développement", 6, []string{"Design final"}},
				{"Rédaction doc", 2, []string{"Design final"}},
				{"Revue légale", 3, []string{"Étude marché"}},
				{"Stratégie marketing", 2, []string{"Étude marché"}},
				{"Création contenu", 4, []string{"Stratégie marketing", "Design final"}},
				{"Tests QA", 3, []string{"Développement"}},
				{"Formation équipe", 2, []string{"Rédaction doc", "Tests QA"}},
				{"Lancement", 1, []string{"Formation équipe", "Création contenu", "Revue légale"}},
			},
			Optimal: 25, // Étude(3)→Prototype(5)→Tests(3)→Design(2)→Dév(6)→QA(3)→Formation(2)→Lancement(1)
		},
	}
}
