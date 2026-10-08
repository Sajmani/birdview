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

// Package web embeds the served assets.
//
// The embed lives here rather than in cmd/birdview because //go:embed only
// matches files at or below the embedding package's own directory, and the
// assets are not under cmd/.
//
// species.json is deliberately NOT embedded: it is a build artifact carrying
// values derived from the ABA checklist, which is all rights reserved (T-005).
// The server reads it from disk at startup instead.
package web

import "embed"

//go:embed static
var Static embed.FS
