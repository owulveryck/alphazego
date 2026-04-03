package taquin

import (
	"fmt"
	"strings"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	yellow = "\033[33m"
)

// String retourne une représentation lisible du taquin avec des bordures
// Unicode. La case vide est affichée en espace, les tuiles en gras jaune.
// Le nombre de coups joués est affiché en en-tête.
func (t *Taquin) String() string {
	size := t.rows * t.cols
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Steps: %d/%d\n", t.steps, t.maxSteps))

	// Ligne du haut
	b.WriteString(" " + topBorder(t.cols, size))

	for r := 0; r < t.rows; r++ {
		b.WriteString(" │")
		for c := 0; c < t.cols; c++ {
			val := t.board[r*t.cols+c]
			if val == 0 {
				b.WriteString(fmt.Sprintf(" %*s ", cellWidth(size), ""))
			} else {
				b.WriteString(fmt.Sprintf(" %s%s%*d%s ", reset, bold+yellow, cellWidth(size), val, reset))
			}
			b.WriteString("│")
		}
		b.WriteString("\n")
		if r < t.rows-1 {
			b.WriteString(" " + midBorder(t.cols, size))
		}
	}

	// Ligne du bas
	b.WriteString(" " + botBorder(t.cols, size))
	return b.String()
}

// cellWidth retourne la largeur d'affichage d'une cellule en fonction
// du nombre total de cases (1 chiffre pour ≤9, 2 pour ≤99).
func cellWidth(size int) int {
	if size <= 10 {
		return 1
	}
	return 2
}

func topBorder(cols, size int) string { return border("┌", "┬", "┐", cols, size) }
func midBorder(cols, size int) string { return border("├", "┼", "┤", cols, size) }
func botBorder(cols, size int) string { return border("└", "┴", "┘", cols, size) }

func border(left, mid, right string, cols, size int) string {
	seg := strings.Repeat("─", cellWidth(size)+2)
	var b strings.Builder
	b.WriteString(left)
	for c := 0; c < cols; c++ {
		b.WriteString(seg)
		if c < cols-1 {
			b.WriteString(mid)
		}
	}
	b.WriteString(right + "\n")
	return b.String()
}
