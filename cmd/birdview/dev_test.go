// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	}).ReadFile("static/next.html")
	if err != nil {
		t.Fatalf("the default asset source has no next.html: %v", err)
	}
	if !strings.Contains(string(b), "AviList") {
		t.Error("the embedded next.html does not credit AviList")
	}
}
