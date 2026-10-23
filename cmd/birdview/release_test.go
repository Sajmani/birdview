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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// dockerStageText splits the Dockerfile into its stages by name, each the raw
// text from its FROM line to the next, with line continuations joined.
func dockerStageText(t *testing.T) (map[string]string, []string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(b), "\\\n", " ")
	stages, order := map[string]string{}, []string(nil)
	name := ""
	for _, line := range strings.Split(text, "\n") {
		if m := fromLine.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			name = m[1]
			if name == "" {
				name = "final"
			}
			order = append(order, name)
			continue
		}
		if name != "" && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			stages[name] += line + "\n"
		}
	}
	return stages, order
}

// TestDockerTestStageGatesTheBuild asserts tech.testedBeforeDeploy in the
// Dockerfile: a test stage runs the tests, all but river/check, against the
// dataset the image will carry; SKIP_TESTS is its only bypass and says so;
// and the image cannot be built without it, because the server build copies
// from it and the final image copies from the server build. Without that
// copy, BuildKit would skip the test stage as unused.
//
// Verifies: tech.testedBeforeDeploy. Criterion: acceptance.dockerTestStageGates.
func TestDockerTestStageGatesTheBuild(t *testing.T) {
	stages, order := dockerStageText(t)
	test, ok := stages["test"]
	if !ok {
		t.Fatalf("no stage named test; stages are %v", order)
	}
	for _, want := range []string{
		"go test $(go list ./... | grep -v '/river/check$')",
		"COPY --from=ingest /out/species.json web/data/species.json",
		"if [ -e SKIP_TESTS ]",
		"WITHOUT running the tests",
		"touch /tested",
	} {
		if !strings.Contains(test, want) {
			t.Errorf("the test stage does not contain %q", want)
		}
	}
	if !strings.Contains(stages["builder"], "COPY --from=test /tested") {
		t.Error("the builder stage does not copy from the test stage, so nothing makes the tests run")
	}
	if !strings.Contains(stages["final"], "COPY --from=builder") {
		t.Error("the final image does not copy from the builder stage")
	}
	if i, j := indexOf(order, "test"), indexOf(order, "builder"); i < 0 || j < i {
		t.Errorf("the test stage must come before the builder that copies from it; order is %v", order)
	}
}

func indexOf(xs []string, x string) int {
	for i, v := range xs {
		if v == x {
			return i
		}
	}
	return -1
}

// TestChecksWorkflow asserts tech.checksOnEveryChange against the workflow
// file: it runs on pushes to main and on pull requests, never on
// pull_request_target; it is read-only and uses no secrets; every action is
// GitHub's own and pinned to a commit; it builds the dataset before testing,
// runs the four standing checks under a shell with pipefail, and fails on any
// skipped test.
//
// Verifies: tech.checksOnEveryChange. Criterion: acceptance.checksWorkflowIsSound.
func TestChecksWorkflow(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "checks.yml"))
	if err != nil {
		t.Fatal(err)
	}
	wf := string(b)
	for _, want := range []string{
		"  push:\n    branches: [main]\n", "  pull_request:\n",
		"permissions:\n  contents: read\n", "shell: bash",
		"go build ./...", "go vet ./...", "gofmt -l .", "go run ./cmd/birdview-ingest",
		"go test -v ./... 2>&1 | tee test.log", `grep -E '^\s*--- SKIP' test.log`,
	} {
		if !strings.Contains(wf, want) {
			t.Errorf("the workflow does not contain %q", want)
		}
	}
	for _, banned := range []string{"pull_request_target", "${{ secrets.", ": write"} {
		if strings.Contains(wf, banned) {
			t.Errorf("the workflow contains %q", banned)
		}
	}
	if strings.Index(wf, "go run ./cmd/birdview-ingest") > strings.Index(wf, "go test -v") {
		t.Error("the dataset is built after the tests run, so the real-data tests would skip")
	}
	uses := regexp.MustCompile(`(?m)^\s*-?\s*uses:\s*(\S+)`).FindAllStringSubmatch(wf, -1)
	if len(uses) == 0 {
		t.Fatal("the workflow uses no actions; the parse has gone wrong")
	}
	pinned := regexp.MustCompile(`^actions/[a-z-]+@[0-9a-f]{40}$`)
	for _, u := range uses {
		if !pinned.MatchString(u[1]) {
			t.Errorf("action %s is not one of GitHub's own pinned to a commit", u[1])
		}
	}
}
