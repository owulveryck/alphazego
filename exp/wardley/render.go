package wardley

import (
	"fmt"
	"image"
	"io"
	"strings"

	svgmap "github.com/owulveryck/wardleyToGo/encoding/svg"
	"github.com/owulveryck/wardleyToGo/parser/wtg2"
)

// RenderSVG génère un rendu SVG de la carte à partir du texte WTG2.
func RenderSVG(w io.Writer, s *State) error {
	p, err := wtg2.NewParser(strings.NewReader(s.WTG2Text()))
	if err != nil {
		return fmt.Errorf("render: parser: %w", err)
	}

	doc, err := p.Parse()
	if err != nil {
		return fmt.Errorf("render: parse: %w", err)
	}

	result, err := wtg2.BuildMap(doc)
	if err != nil {
		return fmt.Errorf("render: build map: %w", err)
	}

	box := image.Rect(30, 50, 1070, 850)
	canvas := image.Rect(0, 0, 1000, 800)

	enc, err := svgmap.NewEncoder(w, box, canvas)
	if err != nil {
		return err
	}

	if err := enc.Encode(result.Map); err != nil {
		return err
	}

	enc.Close()
	return nil
}
