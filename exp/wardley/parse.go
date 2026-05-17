package wardley

import (
	"fmt"
	"io"

	"github.com/owulveryck/wardleyToGo/parser/wtg2"
)

// ParseWTG2 parse un fichier WTG2 et retourne un [State] prêt pour le MCTS.
// maxDepth limite la profondeur de l'arbre de recherche.
func ParseWTG2(r io.Reader, maxDepth int) (*State, error) {
	p, err := wtg2.NewParser(r)
	if err != nil {
		return nil, fmt.Errorf("wardley: erreur création parser: %w", err)
	}

	doc, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("wardley: erreur parsing WTG2: %w", err)
	}

	return buildStateFromDocument(doc, maxDepth)
}

func buildStateFromDocument(doc *wtg2.Document, maxDepth int) (*State, error) {
	gameplaysByNode := make(map[string][]string)
	for _, gp := range doc.Gameplays {
		gameplaysByNode[gp.Target] = append(gameplaysByNode[gp.Target], gp.Type)
	}

	var components []Component
	for _, node := range doc.Nodes {
		phase, err := positionToPhase(node.Evolution)
		if err != nil {
			return nil, fmt.Errorf("wardley: composant %q: %w", node.Name, err)
		}

		visibility := 50
		if node.Visibility >= 0 {
			visibility = int(node.Visibility * 100)
		}

		c := Component{
			Name:       node.Name,
			Phase:      phase,
			Visibility: visibility,
			Type:       node.Type,
			Inertia:    node.Inertia,
		}
		if gps, ok := gameplaysByNode[node.Name]; ok {
			c.Gameplays = make([]string, len(gps))
			copy(c.Gameplays, gps)
		}
		components = append(components, c)
	}

	var edges []Edge
	for _, e := range doc.Edges {
		edges = append(edges, Edge{
			From:  e.From,
			To:    e.To,
			Label: e.Label,
		})
	}

	return NewState(doc.Title, doc.Question, components, edges, maxDepth), nil
}

// positionToPhase convertit une position WTG2 (ex: "III.5") en Phase.
func positionToPhase(evolution string) (Phase, error) {
	if evolution == "" {
		return Genesis, nil
	}

	pos, err := wtg2.ParsePosition(evolution)
	if err != nil {
		return Genesis, fmt.Errorf("position invalide %q: %w", evolution, err)
	}

	switch {
	case pos < 25:
		return Genesis, nil
	case pos < 50:
		return Custom, nil
	case pos < 75:
		return Product, nil
	default:
		return Commodity, nil
	}
}
