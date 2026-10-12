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

// Package docs embeds the user-facing documentation so BirdView can serve
// it.
//
// The embed lives here for the same reason web's does: //go:embed only
// matches files at or below the embedding package's own directory, and
// these are not under web/.
//
// They are served rather than linked because there is no public repository
// to link to, and because T-002 requires the running server to be
// self-contained: a provenance link that needs the network is one that
// breaks (P-069).
//
// The SVGs are diagrams pre-rendered from the Mermaid source in the page of
// the same name, by render-diagrams.py, so drawing them needs no script from
// the network either.
package docs

import "embed"

//go:embed *.md *.svg
var Files embed.FS
