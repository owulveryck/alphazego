package wardley_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/owulveryck/alphazego/exp/wardley"
)

func TestRenderSVG(t *testing.T) {
	wtg2 := `title: Test Map
stages: Genesis, Custom, Product, Commodity

App : III.5
DB : II.5

App -> DB
`
	s := wardley.NewState(wtg2, "Test Map", "question?", 5, sampleProposer(), context.Background())

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

func TestRenderSVGMinimal(t *testing.T) {
	wtg2 := `title: Empty
stages: Genesis, Custom, Product, Commodity
`
	s := wardley.NewState(wtg2, "Empty", "", 5, sampleProposer(), context.Background())

	var buf bytes.Buffer
	err := wardley.RenderSVG(&buf, s)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "<svg") {
		t.Error("even minimal map should produce valid SVG")
	}
}
