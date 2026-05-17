package wardley_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/owulveryck/alphazego/exp/wardley"
)

const parseTestWTG2 = `title: Test Map
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
	proposer := sampleProposer()
	s, err := wardley.ParseWTG2(strings.NewReader(parseTestWTG2), 5, proposer, context.Background())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("Title:", s.Title())
	fmt.Println("Question:", s.Question())
	// Output:
	// Title: Test Map
	// Question: Should we evolve?
}

func TestParseWTG2PreservesText(t *testing.T) {
	proposer := sampleProposer()
	s, err := wardley.ParseWTG2(strings.NewReader(parseTestWTG2), 5, proposer, context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if s.WTG2Text() != parseTestWTG2 {
		t.Error("ParseWTG2 should preserve the raw WTG2 text")
	}
}

func TestParseWTG2ExtractsMetadata(t *testing.T) {
	proposer := sampleProposer()
	s, err := wardley.ParseWTG2(strings.NewReader(parseTestWTG2), 5, proposer, context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if s.Title() != "Test Map" {
		t.Errorf("Title = %q, want %q", s.Title(), "Test Map")
	}
	if s.Question() != "Should we evolve?" {
		t.Errorf("Question = %q, want %q", s.Question(), "Should we evolve?")
	}
}

func TestParseWTG2EmptyInput(t *testing.T) {
	proposer := sampleProposer()
	s, err := wardley.ParseWTG2(strings.NewReader(""), 5, proposer, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s.WTG2Text() != "" {
		t.Errorf("expected empty WTG2 text, got %q", s.WTG2Text())
	}
}
