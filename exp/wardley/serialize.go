package wardley

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"strings"
)

// phaseToWTG2 convertit une Phase en position WTG2 (centre de la phase).
func phaseToWTG2(p Phase) string {
	switch p {
	case Genesis:
		return "I.5"
	case Custom:
		return "II.5"
	case Product:
		return "III.5"
	case Commodity:
		return "IV.5"
	default:
		return "II.5"
	}
}

// SerializeWTG2 génère le texte WTG2 correspondant à l'état de la carte.
func SerializeWTG2(s *State) string {
	var b strings.Builder

	if s.Title() != "" {
		fmt.Fprintf(&b, "title: %s\n", s.Title())
	}
	if s.Question() != "" {
		fmt.Fprintf(&b, "question: %q\n", s.Question())
	}
	b.WriteString("stages: Genesis, Custom-Built, Product, Commodity\n\n")

	components := s.Components()
	edges := s.Edges()

	for _, c := range components {
		pos := phaseToWTG2(c.Phase)
		if c.Type != "" {
			fmt.Fprintf(&b, "%s : %s (%s)\n", c.Name, pos, c.Type)
		} else {
			fmt.Fprintf(&b, "%s : %s\n", c.Name, pos)
		}
	}

	if len(edges) > 0 {
		b.WriteString("\n")
		for _, e := range edges {
			if e.Label != "" {
				fmt.Fprintf(&b, "%s -[%s]-> %s\n", e.From, e.Label, e.To)
			} else {
				fmt.Fprintf(&b, "%s -> %s\n", e.From, e.To)
			}
		}
	}

	hasGameplays := false
	for _, c := range components {
		if len(c.Gameplays) > 0 {
			hasGameplays = true
			break
		}
	}
	if hasGameplays {
		b.WriteString("\n")
		for _, c := range components {
			for _, gp := range c.Gameplays {
				fmt.Fprintf(&b, "gameplay %s on %s\n", gp, c.Name)
			}
		}
	}

	annotations := s.Annotations()
	if len(annotations) > 0 {
		b.WriteString("\n")
		for _, a := range annotations {
			fmt.Fprintf(&b, "%s %q on %s\n", a.Kind, a.Text, a.Target)
		}
	}

	return b.String()
}

const playgroundBase = "https://owulveryck.github.io/wardleyToGo/?wtg2="

// PlaygroundURL retourne l'URL du playground wardleyToGo pour visualiser la carte.
// Le texte WTG2 est compressé en gzip puis encodé en base64url sans padding.
func PlaygroundURL(s *State) (string, error) {
	text := SerializeWTG2(s)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(text)); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buf.Bytes())
	return playgroundBase + encoded, nil
}
