package wardley_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/owulveryck/alphazego/exp/wardley"
)

func ExampleSerializeWTG2() {
	comps := []wardley.Component{
		{Name: "App", Phase: wardley.Product, Type: "build"},
		{Name: "DB", Phase: wardley.Custom, Type: "buy"},
	}
	edges := []wardley.Edge{
		{From: "App", To: "DB"},
	}
	s := wardley.NewState("Mon projet", "Faut-il évoluer ?", comps, edges, 5)
	fmt.Print(wardley.SerializeWTG2(s))
	// Output:
	// title: Mon projet
	// question: "Faut-il évoluer ?"
	// stages: Genesis, Custom-Built, Product, Commodity
	//
	// App : III.5 (build)
	// DB : II.5 (buy)
	//
	// App -> DB
}

func ExampleSerializeWTG2_withGameplays() {
	comps := []wardley.Component{
		{Name: "Core", Phase: wardley.Product, Gameplays: []string{"open-source", "ILC"}},
	}
	s := wardley.NewState("test", "", comps, nil, 5)
	fmt.Print(wardley.SerializeWTG2(s))
	// Output:
	// title: test
	// stages: Genesis, Custom-Built, Product, Commodity
	//
	// Core : III.5
	//
	// gameplay open-source on Core
	// gameplay ILC on Core
}

func TestSerializeContainsAllComponents(t *testing.T) {
	s := wardley.NewState("t", "q", sampleComponents(), sampleEdges(), 5)
	output := wardley.SerializeWTG2(s)

	for _, c := range sampleComponents() {
		if !strings.Contains(output, c.Name) {
			t.Errorf("output missing component %q", c.Name)
		}
	}
}

func TestSerializeContainsEdges(t *testing.T) {
	s := wardley.NewState("t", "q", sampleComponents(), sampleEdges(), 5)
	output := wardley.SerializeWTG2(s)

	if !strings.Contains(output, "App -> DB") {
		t.Error("output missing edge App -> DB")
	}
	if !strings.Contains(output, "DB -> Cloud") {
		t.Error("output missing edge DB -> Cloud")
	}
}

func TestPlaygroundURL(t *testing.T) {
	comps := []wardley.Component{
		{Name: "App", Phase: wardley.Product, Type: "build"},
		{Name: "DB", Phase: wardley.Custom, Type: "buy"},
	}
	s := wardley.NewState("Test", "question?", comps, nil, 5)

	url, err := wardley.PlaygroundURL(s)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(url, "https://owulveryck.github.io/wardleyToGo/?wtg2=") {
		t.Errorf("unexpected URL prefix: %s", url)
	}

	// L'URL ne doit pas contenir de padding '='
	encoded := strings.TrimPrefix(url, "https://owulveryck.github.io/wardleyToGo/?wtg2=")
	if strings.Contains(encoded, "=") {
		t.Error("encoded string should not contain padding")
	}
	if len(encoded) == 0 {
		t.Error("encoded string should not be empty")
	}
}

func TestSerializeEdgeWithLabel(t *testing.T) {
	edges := []wardley.Edge{{From: "A", To: "B", Label: "data flow"}}
	comps := []wardley.Component{
		{Name: "A", Phase: wardley.Custom},
		{Name: "B", Phase: wardley.Product},
	}
	s := wardley.NewState("t", "q", comps, edges, 5)
	output := wardley.SerializeWTG2(s)

	if !strings.Contains(output, "A -[data flow]-> B") {
		t.Errorf("expected labeled edge, got:\n%s", output)
	}
}
