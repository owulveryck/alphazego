package wardley_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/owulveryck/alphazego/exp/wardley"
)

func TestRenderSVG(t *testing.T) {
	comps := []wardley.Component{
		{Name: "App", Phase: wardley.Product, Visibility: 80, Type: "build"},
		{Name: "DB", Phase: wardley.Custom, Visibility: 50, Type: "buy"},
	}
	edges := []wardley.Edge{
		{From: "App", To: "DB"},
	}
	s := wardley.NewState("Test Map", "question?", comps, edges, 5)

	var buf bytes.Buffer
	err := wardley.RenderSVG(&buf, s)
	if err != nil {
		t.Fatal(err)
	}

	svg := buf.String()
	if !strings.Contains(svg, "<svg") {
		t.Error("output should contain <svg tag")
	}
	if !strings.Contains(svg, "App") {
		t.Error("output should contain component name App")
	}
	if !strings.Contains(svg, "DB") {
		t.Error("output should contain component name DB")
	}
}

func TestRenderSVGEmpty(t *testing.T) {
	s := wardley.NewState("Empty", "", nil, nil, 5)

	var buf bytes.Buffer
	err := wardley.RenderSVG(&buf, s)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "<svg") {
		t.Error("even empty map should produce valid SVG")
	}
}
