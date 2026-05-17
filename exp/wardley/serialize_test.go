package wardley_test

import (
	"context"
	"strings"
	"testing"

	"github.com/owulveryck/alphazego/exp/wardley"
)

func TestSerializeWTG2ReturnsRawText(t *testing.T) {
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 5, sampleProposer(), context.Background())
	output := wardley.SerializeWTG2(s)

	if output != sampleWTG2Text {
		t.Errorf("SerializeWTG2 should return raw WTG2 text:\ngot:  %q\nwant: %q", output, sampleWTG2Text)
	}
}

func TestSerializePreservesContent(t *testing.T) {
	s := wardley.NewState(sampleWTG2Text, "Test", "q", 5, sampleProposer(), context.Background())
	output := wardley.SerializeWTG2(s)

	if !strings.Contains(output, "App : III.5") {
		t.Error("output missing component App")
	}
	if !strings.Contains(output, "DB : II.5") {
		t.Error("output missing component DB")
	}
	if !strings.Contains(output, "App -> DB") {
		t.Error("output missing edge App -> DB")
	}
}

func TestPlaygroundURL(t *testing.T) {
	s := wardley.NewState(sampleWTG2Text, "Test", "question?", 5, sampleProposer(), context.Background())

	url, err := wardley.PlaygroundURL(s)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(url, "https://owulveryck.github.io/wardleyToGo/?wtg2=") {
		t.Errorf("unexpected URL prefix: %s", url)
	}

	encoded := strings.TrimPrefix(url, "https://owulveryck.github.io/wardleyToGo/?wtg2=")
	if strings.Contains(encoded, "=") {
		t.Error("encoded string should not contain padding")
	}
	if len(encoded) == 0 {
		t.Error("encoded string should not be empty")
	}
}
