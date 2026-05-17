package wardley

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed skill.md
var skillContext string

// formatPolicyPrompt construit le prompt pour évaluer la qualité stratégique
// d'un état enfant. Le Judge doit retourner un score entre 0 et 1.
func formatPolicyPrompt(wtg2Text, question string) string {
	var b strings.Builder

	b.WriteString(skillContext)
	b.WriteString("\n\n---\n\n")
	b.WriteString("Tu es un expert en stratégie Wardley. Évalue la carte suivante.\n\n")
	fmt.Fprintf(&b, "Carte Wardley actuelle :\n```wtg2\n%s```\n\n", wtg2Text)
	fmt.Fprintf(&b, "Question stratégique : %s\n\n", question)
	b.WriteString("Évalue la qualité stratégique de cette carte sur une échelle de 0 à 1.\n")
	b.WriteString("Utilise la Strategic Completeness Checklist et les concepts de ")
	b.WriteString("climat/doctrine/manœuvre pour ton évaluation. Considère :\n")
	b.WriteString("- Cohérence de la chaîne de valeur\n")
	b.WriteString("- Pertinence des gameplays par rapport au contexte\n")
	b.WriteString("- Alignement EVT (teams vs phases d'évolution)\n")
	b.WriteString("- Violations de doctrine (NIH, dispersion, strategy theatre...)\n")
	b.WriteString("- Gestion de l'inertie et des signaux climatiques\n")

	return b.String()
}

// phaseNames associe chaque Phase à son nom lisible.
var phaseNames = [...]string{"Genesis", "Custom", "Product", "Commodity"}

// moveDescription retourne une description enrichie du move.
// Pour un Evolve, elle inclut la transition de phase (ex: "Custom → Product").
func moveDescription(m Move, components []Component) string {
	switch m.Type {
	case Evolve:
		for _, c := range components {
			if c.Name == m.Component {
				from := phaseNames[c.Phase]
				to := phaseNames[c.Phase+1]
				return fmt.Sprintf("EVOLVE %q (%s → %s)", m.Component, from, to)
			}
		}
		return m.String()
	default:
		return m.String()
	}
}

// formatBatchPolicyPrompt construit un prompt unique pour scorer N moves d'un coup.
// Le Judge doit retourner un tableau JSON de N scores entre 0 et 1.
func formatBatchPolicyPrompt(wtg2Text, question string, moves []Move, components []Component) string {
	var b strings.Builder

	b.WriteString(skillContext)
	b.WriteString("\n\n---\n\n")
	b.WriteString("Tu es un expert en stratégie Wardley. Évalue chaque move candidat.\n\n")
	fmt.Fprintf(&b, "Carte Wardley actuelle :\n```wtg2\n%s```\n\n", wtg2Text)
	fmt.Fprintf(&b, "Question stratégique : %s\n\n", question)
	b.WriteString("Moves candidats :\n")
	for i, m := range moves {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, moveDescription(m, components))
	}
	b.WriteString("\nPour chaque move, évalue sa pertinence stratégique sur [0, 1].\n")
	b.WriteString("Considère : cohérence de la chaîne de valeur, doctrine, EVT, inertie.\n\n")
	b.WriteString("Réponds uniquement par un tableau JSON de scores, ex: [0.7, 0.3, 0.8]\n")

	return b.String()
}

// formatValuePrompt construit le prompt pour estimer la progression de la
// séquence stratégique vers une réponse à la question.
func formatValuePrompt(wtg2Text, question string, history []Move) string {
	var b strings.Builder

	b.WriteString(skillContext)
	b.WriteString("\n\n---\n\n")
	b.WriteString("Tu es un expert en stratégie Wardley. Évalue la progression de cette séquence stratégique.\n\n")
	fmt.Fprintf(&b, "Carte Wardley :\n```wtg2\n%s```\n\n", wtg2Text)
	fmt.Fprintf(&b, "Question stratégique : %s\n\n", question)

	if len(history) > 0 {
		b.WriteString("Moves effectués :\n")
		for i, m := range history {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, m.String())
		}
		b.WriteString("\n")
	}

	b.WriteString("Estime la probabilité que cette séquence stratégique mène à une bonne ")
	b.WriteString("réponse à la question. Utilise le framework OODA et le Value Flywheel ")
	b.WriteString("pour juger la progression.\n")
	b.WriteString("0 = stratégie incohérente ou contre-productive.\n")
	b.WriteString("1 = stratégie optimale répondant clairement à la question.\n")

	return b.String()
}
