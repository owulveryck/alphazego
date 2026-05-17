package wardley

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/owulveryck/wardleyToGo/parser/wtg2"
)

// ParseWTG2 lit un fichier WTG2 et retourne un [State] prêt pour le MCTS.
// Le texte WTG2 est stocké tel quel comme représentation canonique.
func ParseWTG2(r io.Reader, maxDepth int, proposer Proposer, ctx context.Context, opts ...Option) (*State, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("wardley: erreur lecture: %w", err)
	}
	text := string(data)

	p, err := wtg2.NewParser(strings.NewReader(text))
	if err != nil {
		return nil, fmt.Errorf("wardley: erreur création parser: %w", err)
	}

	doc, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("wardley: erreur parsing WTG2: %w", err)
	}

	return NewState(text, doc.Title, doc.Question, maxDepth, proposer, ctx, opts...), nil
}

// validateWTG2 vérifie qu'un texte WTG2 est syntaxiquement valide.
func validateWTG2(text string) error {
	p, err := wtg2.NewParser(strings.NewReader(text))
	if err != nil {
		return fmt.Errorf("validation WTG2: %w", err)
	}
	_, err = p.Parse()
	if err != nil {
		return fmt.Errorf("validation WTG2: %w", err)
	}
	return nil
}

// nodeNames extrait les noms des noeuds d'un texte WTG2.
// Utilisé pour valider les cibles d'annotations.
func nodeNames(wtg2Text string) map[string]bool {
	p, err := wtg2.NewParser(strings.NewReader(wtg2Text))
	if err != nil {
		return nil
	}
	doc, err := p.Parse()
	if err != nil {
		return nil
	}
	names := make(map[string]bool, len(doc.Nodes))
	for _, n := range doc.Nodes {
		names[n.Name] = true
	}
	return names
}
