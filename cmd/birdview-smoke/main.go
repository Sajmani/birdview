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

// Command birdview-smoke checks a running BirdView deployment after a deploy
// (tech.smokeCheckAfterDeploy):
//
//	go run ./cmd/birdview-smoke https://birdview.site \
//	    https://www.birdview.site https://birdview-570018079643.us-central1.run.app
//
// The first argument is the deployment's canonical origin; any others are
// origins that must redirect to it (tech.oneCanonicalHost). It prints one line
// per check and exits 1 if any failed.
package main

import (
	"fmt"
	"os"

	"birdview/internal/smoke"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: birdview-smoke <canonical origin> [origin that redirects to it ...]")
		os.Exit(2)
	}
	failed := 0
	for _, r := range smoke.Run(os.Args[1], os.Args[2:], nil) {
		if r.Err != nil {
			failed++
			fmt.Printf("FAIL  %s: %v\n", r.Name, r.Err)
		} else {
			fmt.Printf("ok    %s\n", r.Name)
		}
	}
	if failed > 0 {
		fmt.Printf("%d check(s) failed\n", failed)
		os.Exit(1)
	}
}
