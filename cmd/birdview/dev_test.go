package main

import (
	"strings"
	"testing"
)

// TestEmbedIsTheDefault guards the one way -dev could do harm: shipping a
// binary that reads assets off the deploy host's disk instead of its own
// embed. The flag defaults empty, and devDir empty means the embed.
func TestEmbedIsTheDefault(t *testing.T) {
	if devDir != "" {
		t.Fatalf("devDir is %q at rest; a deployed binary would read from disk", devDir)
	}
	s, err := newServer(testDataset())
	if err != nil {
		t.Fatal(err)
	}
	if s.devDir != "" {
		t.Errorf("a server built without -dev has devDir %q", s.devDir)
	}
	// And the embed actually carries the templates, so the default works.
	b, err := assetFS.(interface {
		ReadFile(string) ([]byte, error)
	}).ReadFile("static/index.html")
	if err != nil {
		t.Fatalf("the default asset source has no index.html: %v", err)
	}
	if !strings.Contains(string(b), "AviList") {
		t.Error("the embedded index.html does not credit AviList")
	}
}
