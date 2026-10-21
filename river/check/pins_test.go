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

package check

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestVendoredHashesMatch checks every committed input against the SHA-256
// its provenance record holds, so a vendored file cannot change without its
// record changing with it.
//
// Scope: committed inputs only. wiley is quoted rather than vendored and has
// no file to hash; aba is fetched at build and verified by the ingest instead
// (acceptance.abaFileHashMatches).
//
// Verifies: tech.recordBeforeConsuming. Criterion: acceptance.vendoredHashesMatch.
func TestVendoredHashesMatch(t *testing.T) {
	// Table rows of the form | `//path` | origin | `sha256` |
	row := regexp.MustCompile("\\|\\s*`//([^`]+)`\\s*\\|[^|]*\\|\\s*`([0-9a-f]{64})`\\s*\\|")
	records, err := filepath.Glob(filepath.Join("..", "requirements", "imported", "*", "*", "PROVENANCE.md"))
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, record := range records {
		b, err := os.ReadFile(record)
		if err != nil {
			t.Fatal(err)
		}
		source := filepath.Base(filepath.Dir(record))
		for _, m := range row.FindAllStringSubmatch(string(b), -1) {
			rel, want := m[1], m[2]
			data, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(rel)))
			if err != nil {
				t.Errorf("%s: missing file %s", source, rel)
				continue
			}
			checked++
			sum := sha256.Sum256(data)
			if got := hex.EncodeToString(sum[:]); got != want {
				t.Errorf("%s: %s\n    recorded %s\n    actual   %s", source, rel, want, got)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no vendored files found to check; the provenance tables or the layout have changed")
	}
	t.Logf("%d vendored file(s) match their recorded hashes", checked)
}
