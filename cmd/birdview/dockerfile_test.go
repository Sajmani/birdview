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
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// dockerStage is one FROM ... AS name block of the Dockerfile: the paths it
// copies in from the build context, and the commands it builds.
type dockerStage struct {
	name   string
	copies []string // context paths, cleaned; not COPY --from
	builds []string // ./cmd/... packages named by go build
}

var (
	fromLine  = regexp.MustCompile(`(?i)^FROM\s+\S+(?:\s+AS\s+(\S+))?`)
	copyLine  = regexp.MustCompile(`(?i)^COPY\s+(.*)$`)
	goBuildOf = regexp.MustCompile(`go build[^&;]*?\s(\./cmd/[\w-]+)`)
)

// parseDockerfile splits the Dockerfile into stages. Line continuations are
// joined first, so a go build split across lines is still seen.
func parseDockerfile(t *testing.T, file string) []dockerStage {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("reading %s: %v", file, err)
	}
	text := strings.ReplaceAll(string(b), "\\\n", " ")
	var stages []dockerStage
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if m := fromLine.FindStringSubmatch(line); m != nil {
			stages = append(stages, dockerStage{name: m[1]})
			continue
		}
		if len(stages) == 0 {
			continue
		}
		st := &stages[len(stages)-1]
		if m := copyLine.FindStringSubmatch(line); m != nil {
			args := strings.Fields(m[1])
			if len(args) > 0 && strings.HasPrefix(args[0], "--from") {
				continue
			}
			for _, a := range args[:len(args)-1] { // the last is the destination
				if !strings.HasPrefix(a, "--") {
					st.copies = append(st.copies, path.Clean(a))
				}
			}
			continue
		}
		for _, m := range goBuildOf.FindAllStringSubmatch(line, -1) {
			st.builds = append(st.builds, path.Clean(m[1]))
		}
	}
	return stages
}

// moduleDeps returns the module-relative directory of pkg and of every package
// in this module it imports, directly or not. Every non-test file is read,
// whatever its build tags, so the set can only be too large, never too small.
// Embedded files need nothing extra: go:embed cannot reach outside its
// package's directory.
func moduleDeps(t *testing.T, root, module, pkg string) []string {
	t.Helper()
	seen := map[string]bool{}
	var visit func(dir string)
	visit = func(dir string) {
		if seen[dir] {
			return
		}
		seen[dir] = true
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			t.Fatalf("reading package %s: %v", dir, err)
		}
		fset := token.NewFileSet()
		for _, e := range entries {
			n := e.Name()
			if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, filepath.Join(root, dir, n), nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parsing %s/%s: %v", dir, n, err)
			}
			for _, imp := range f.Imports {
				p, _ := strconv.Unquote(imp.Path.Value)
				if rest, ok := strings.CutPrefix(p, module+"/"); ok {
					visit(rest)
				}
			}
		}
	}
	visit(strings.TrimPrefix(pkg, "./"))
	var dirs []string
	for d := range seen {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	return dirs
}

// covered reports whether dir lies at or beneath one of the copied paths.
func covered(dir string, copies []string) bool {
	for _, c := range copies {
		if c == "." || dir == c || strings.HasPrefix(dir, c+"/") {
			return true
		}
	}
	return false
}

// TestDockerStagesCopyWhatTheyBuild asserts that each Dockerfile stage that
// builds a command copies in every package of this module the command
// imports, and go.mod and go.sum, so the image builds from a clean checkout.
// It reads the Go sources rather than running docker, which the tests must
// not need; what that leaves unchecked is in the criterion.
//
// Verifies: tech.containerImageBuilds. Criterion: acceptance.dockerStagesCopyWhatTheyBuild.
func TestDockerStagesCopyWhatTheyBuild(t *testing.T) {
	root := filepath.Join("..", "..")
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^module\s+(\S+)`).FindSubmatch(mod)
	if m == nil {
		t.Fatal("go.mod names no module")
	}
	module := string(m[1])

	stages := parseDockerfile(t, filepath.Join(root, "Dockerfile"))
	built := map[string]bool{}
	for _, st := range stages {
		if len(st.builds) == 0 {
			continue
		}
		for _, f := range []string{"go.mod", "go.sum"} {
			if !covered(f, st.copies) {
				t.Errorf("Dockerfile stage %q builds %v but does not copy %s", st.name, st.builds, f)
			}
		}
		for _, cmd := range st.builds {
			built[cmd] = true
			for _, dir := range moduleDeps(t, root, module, cmd) {
				if !covered(dir, st.copies) {
					t.Errorf("Dockerfile stage %q builds %s, which needs %s/%s, but copies only %v",
						st.name, cmd, module, dir, st.copies)
				}
			}
		}
	}
	// Fail loudly if the parse found nothing to check: a Dockerfile this test
	// can no longer read must not pass as one that copies everything.
	for _, want := range []string{"cmd/birdview", "cmd/birdview-ingest"} {
		if !built[want] {
			t.Errorf("found no Dockerfile stage that builds ./%s; stages parsed: %+v", want, stages)
		}
	}
}
