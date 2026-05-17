package wardley_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/owulveryck/alphazego/exp/wardley"
)

const sampleWTG2 = `title: Test Map
question: "Should we evolve?"
stages: Genesis, Custom, Product, Commodity

anchor User : IV.5

App : III.5 {
  type: build
}

DB : II.3 (buy)

Cloud : IV.8 (outsource)

Sensor : I.2 !! >> II.5

User -> App
App -> DB
DB -> Cloud

gameplay open-source on App
gameplay ILC on DB
`

func ExampleParseWTG2() {
	s, err := wardley.ParseWTG2(strings.NewReader(sampleWTG2), 5)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("Title:", s.Title())
	fmt.Println("Question:", s.Question())
	fmt.Println("Components:", len(s.Components()))
	// Output:
	// Title: Test Map
	// Question: Should we evolve?
	// Components: 5
}

func TestParseWTG2Phases(t *testing.T) {
	s, err := wardley.ParseWTG2(strings.NewReader(sampleWTG2), 5)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]wardley.Phase{
		"User":   wardley.Commodity,
		"App":    wardley.Product,
		"DB":     wardley.Custom,
		"Cloud":  wardley.Commodity,
		"Sensor": wardley.Genesis,
	}

	for _, c := range s.Components() {
		want, ok := expected[c.Name]
		if !ok {
			t.Errorf("composant inattendu: %q", c.Name)
			continue
		}
		if c.Phase != want {
			t.Errorf("%s: phase = %v, want %v", c.Name, c.Phase, want)
		}
	}
}

func TestParseWTG2Types(t *testing.T) {
	s, err := wardley.ParseWTG2(strings.NewReader(sampleWTG2), 5)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"App":   "build",
		"DB":    "buy",
		"Cloud": "outsource",
	}

	for _, c := range s.Components() {
		if want, ok := expected[c.Name]; ok && c.Type != want {
			t.Errorf("%s: type = %q, want %q", c.Name, c.Type, want)
		}
	}
}

func TestParseWTG2Inertia(t *testing.T) {
	s, err := wardley.ParseWTG2(strings.NewReader(sampleWTG2), 5)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range s.Components() {
		if c.Name == "Sensor" && c.Inertia != 2 {
			t.Errorf("Sensor inertia = %d, want 2", c.Inertia)
		}
	}
}

func TestParseWTG2Gameplays(t *testing.T) {
	s, err := wardley.ParseWTG2(strings.NewReader(sampleWTG2), 5)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string][]string{
		"App": {"open-source"},
		"DB":  {"ILC"},
	}

	for _, c := range s.Components() {
		if want, ok := expected[c.Name]; ok {
			if len(c.Gameplays) != len(want) {
				t.Errorf("%s: %d gameplays, want %d", c.Name, len(c.Gameplays), len(want))
				continue
			}
			for i, gp := range c.Gameplays {
				if gp != want[i] {
					t.Errorf("%s: gameplay[%d] = %q, want %q", c.Name, i, gp, want[i])
				}
			}
		}
	}
}

func TestParseWTG2Edges(t *testing.T) {
	s, err := wardley.ParseWTG2(strings.NewReader(sampleWTG2), 5)
	if err != nil {
		t.Fatal(err)
	}

	edges := s.Edges()
	if len(edges) != 3 {
		t.Fatalf("got %d edges, want 3", len(edges))
	}

	edgeSet := make(map[string]bool)
	for _, e := range edges {
		edgeSet[e.From+"->"+e.To] = true
	}

	for _, want := range []string{"User->App", "App->DB", "DB->Cloud"} {
		if !edgeSet[want] {
			t.Errorf("missing edge %s", want)
		}
	}
}
