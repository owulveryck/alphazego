package wardley

import (
	"image"
	"io"

	wardleyToGo "github.com/owulveryck/wardleyToGo"
	"github.com/owulveryck/wardleyToGo/components/wardley"
	svgmap "github.com/owulveryck/wardleyToGo/encoding/svg"
)

// phaseToX convertit une Phase en coordonnée X (0-100).
func phaseToX(p Phase) int {
	switch p {
	case Genesis:
		return 12
	case Custom:
		return 37
	case Product:
		return 62
	case Commodity:
		return 87
	default:
		return 50
	}
}

// RenderSVG génère un rendu SVG de la carte à partir de l'état.
func RenderSVG(w io.Writer, s *State) error {
	m := buildMap(s)

	box := image.Rect(30, 50, 1070, 850)
	canvas := image.Rect(0, 0, 1000, 800)

	enc, err := svgmap.NewEncoder(w, box, canvas)
	if err != nil {
		return err
	}

	if err := enc.Encode(m); err != nil {
		return err
	}

	enc.Close()
	return nil
}

func buildMap(s *State) *wardleyToGo.Map {
	m := wardleyToGo.NewMap(0)

	components := s.Components()
	edges := s.Edges()

	nodeByName := make(map[string]*wardley.Component, len(components))

	for i, c := range components {
		wc := wardley.NewComponent(int64(i + 1))
		wc.Label = c.Name
		wc.Placement = image.Pt(phaseToX(c.Phase), c.Visibility)

		switch c.Type {
		case "build":
			wc.Type = wardley.BuildComponent
		case "buy":
			wc.Type = wardley.BuyComponent
		case "outsource":
			wc.Type = wardley.OutsourceComponent
		}

		for _, gp := range c.Gameplays {
			wc.Gameplays = append(wc.Gameplays, wardley.ComponentGameplay{
				Type: gp,
			})
		}

		nodeByName[c.Name] = wc
		_ = m.AddComponent(wc)
	}

	for _, e := range edges {
		from, okFrom := nodeByName[e.From]
		to, okTo := nodeByName[e.To]
		if !okFrom || !okTo {
			continue
		}
		collab := &wardley.Collaboration{
			F:     from,
			T:     to,
			Label: e.Label,
		}
		_ = m.SetCollaboration(collab)
	}

	return m
}
