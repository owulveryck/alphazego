package wardley

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed skill.md
var skillContext string

// formatProposalPrompt construit le prompt pour demander au LLM de proposer
// N modifications stratégiques de la carte WTG2.
func formatProposalPrompt(wtg2Text, question string, history []string, n int) string {
	var b strings.Builder

	b.WriteString(skillContext)
	b.WriteString("\n\n---\n\n")
	b.WriteString("Tu es un expert en stratégie Wardley. Propose des modifications stratégiques.\n\n")
	fmt.Fprintf(&b, "Carte Wardley actuelle :\n```wtg2\n%s```\n\n", wtg2Text)
	fmt.Fprintf(&b, "Question stratégique : %s\n\n", question)

	if len(history) > 0 {
		b.WriteString("Modifications déjà appliquées :\n")
		for i, h := range history {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, h)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "Propose %d modifications stratégiques distinctes.\n", n)
	b.WriteString("Pour chaque candidat, fournis :\n")
	b.WriteString("1. Une description de la modification (quoi et pourquoi)\n")
	b.WriteString("2. Le texte WTG2 COMPLET résultant\n")
	b.WriteString("3. Un score de confiance entre 0 et 1\n\n")
	b.WriteString("IMPORTANT :\n")
	b.WriteString("- Modifie SEULEMENT les lignes pertinentes à ta modification\n")
	b.WriteString("- Garde TOUTES les positions, notes, groupes, pipelines et edges ")
	b.WriteString("inchangés sauf ceux directement affectés\n")
	b.WriteString("- Chaque modification doit être un pas stratégique cohérent\n")
	b.WriteString("- Les modifications possibles incluent : évoluer un composant ")
	b.WriteString("(changer sa position), ajouter un gameplay, ajouter une note/signal, ")
	b.WriteString("ajouter un nouveau composant ou edge, créer un groupe\n\n")
	b.WriteString("Réponds uniquement par un tableau JSON :\n")
	b.WriteString(`[{"description":"...","wtg2":"...","confidence":0.8},...]`)
	b.WriteString("\n")

	return b.String()
}

// formatValuePrompt construit le prompt pour estimer la progression de la
// séquence stratégique vers une réponse à la question.
func formatValuePrompt(wtg2Text, question string, history []string) string {
	var b strings.Builder

	b.WriteString(skillContext)
	b.WriteString("\n\n---\n\n")
	b.WriteString("Tu es un expert en stratégie Wardley. Évalue la progression de cette séquence stratégique.\n\n")
	fmt.Fprintf(&b, "Carte Wardley :\n```wtg2\n%s```\n\n", wtg2Text)
	fmt.Fprintf(&b, "Question stratégique : %s\n\n", question)

	if len(history) > 0 {
		b.WriteString("Modifications effectuées :\n")
		for i, h := range history {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, h)
		}
		b.WriteString("\n")
	}

	b.WriteString("Estime la probabilité que cette séquence stratégique mène à une bonne ")
	b.WriteString("réponse à la question. Utilise le framework OODA et le Value Flywheel ")
	b.WriteString("pour juger la progression.\n")
	b.WriteString("0 = stratégie incohérente ou contre-productive.\n")
	b.WriteString("1 = stratégie optimale répondant clairement à la question.\n")
	b.WriteString("\nRéponds uniquement par un nombre décimal entre 0 et 1.\n")

	return b.String()
}

// formatAnnotationPrompt construit le prompt pour générer des notes explicatives
// sur la carte résultante. Le LLM doit retourner un tableau JSON d'annotations.
func formatAnnotationPrompt(wtg2Text, question string, history []string) string {
	var b strings.Builder

	b.WriteString(skillContext)
	b.WriteString("\n\n---\n\n")
	b.WriteString("Tu es un expert en stratégie Wardley. Annote la carte suivante ")
	b.WriteString("avec des notes explicatives pour aider un décideur à comprendre ")
	b.WriteString("la stratégie proposée.\n\n")
	fmt.Fprintf(&b, "Carte Wardley résultante :\n```wtg2\n%s```\n\n", wtg2Text)
	fmt.Fprintf(&b, "Question stratégique : %s\n\n", question)

	if len(history) > 0 {
		b.WriteString("Séquence de décisions stratégiques appliquées :\n")
		for i, h := range history {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, h)
		}
		b.WriteString("\n")
	}

	b.WriteString("Génère des notes explicatives pour les composants clés de la carte. ")
	b.WriteString("Chaque note doit expliquer :\n")
	b.WriteString("- Pourquoi un composant a été évolué ou pourquoi un gameplay a été appliqué\n")
	b.WriteString("- Les implications stratégiques (opportunités, risques, dépendances)\n")
	b.WriteString("- Les signaux climatiques ou violations de doctrine pertinents\n\n")
	b.WriteString("Ajoute aussi des notes sur les composants non modifiés si leur position ")
	b.WriteString("actuelle mérite une explication stratégique.\n\n")
	b.WriteString("Réponds uniquement par un tableau JSON d'objets avec les champs ")
	b.WriteString("\"kind\" (\"note\" ou \"warning\"), \"text\" et \"target\" (nom exact du composant).\n")
	b.WriteString("Exemple : [{\"kind\":\"note\",\"text\":\"Évolué vers Product pour standardiser\",\"target\":\"API\"}]\n")

	return b.String()
}
