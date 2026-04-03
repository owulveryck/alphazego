package reasoning

import (
	"fmt"
	"strings"
)

// formatGeneratePrompt construit le prompt pour le Generator.
// Il inclut la question, les étapes précédentes, le critère de succès,
// et l'instruction de conclure avec ConclusionPrefix quand le raisonnement
// est abouti.
func formatGeneratePrompt(question, criterion string, steps []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Question : %s\n\n", question))
	b.WriteString(fmt.Sprintf("Critère de succès : %s\n\n", criterion))

	if len(steps) > 0 {
		b.WriteString("Raisonnement en cours :\n")
		for i, step := range steps {
			b.WriteString(fmt.Sprintf("  Étape %d : %s\n", i+1, step))
		}
		b.WriteString("\n")
	}

	b.WriteString("Propose la prochaine étape de raisonnement. ")
	b.WriteString(fmt.Sprintf("Si tu as assez d'éléments pour conclure, commence ta réponse par \"%s\".", ConclusionPrefix))

	return b.String()
}

// formatJudgePrompt construit le prompt pour évaluer la qualité d'un
// chemin de raisonnement. Le Judge doit retourner un score entre 0 et 1.
func formatJudgePrompt(question, criterion string, steps []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Question : %s\n\n", question))
	b.WriteString(fmt.Sprintf("Critère de succès : %s\n\n", criterion))

	b.WriteString("Raisonnement :\n")
	for i, step := range steps {
		b.WriteString(fmt.Sprintf("  Étape %d : %s\n", i+1, step))
	}

	b.WriteString("\nÉvalue la qualité de ce raisonnement sur une échelle de 0 à 1. ")
	b.WriteString("0 = hors sujet ou incorrect. 1 = raisonnement complet et correct.")

	return b.String()
}

// formatValuePrompt construit le prompt pour estimer la valeur de l'état
// courant (progression vers la solution). Le Judge doit retourner un score
// entre 0 et 1.
func formatValuePrompt(question, criterion string, steps []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Question : %s\n\n", question))
	b.WriteString(fmt.Sprintf("Critère de succès : %s\n\n", criterion))

	if len(steps) == 0 {
		b.WriteString("Aucune étape de raisonnement n'a encore été effectuée.\n")
	} else {
		b.WriteString("Raisonnement en cours :\n")
		for i, step := range steps {
			b.WriteString(fmt.Sprintf("  Étape %d : %s\n", i+1, step))
		}
	}

	b.WriteString("\nEstime la probabilité que ce raisonnement mène à une réponse correcte. ")
	b.WriteString("0 = aucune chance. 1 = réponse certaine.")

	return b.String()
}
