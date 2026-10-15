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

// Package spec embeds the normative requirement documents so BirdView can
// serve them, the way package docs does for the user-facing guides.
//
// **Only the top-level documents.** `*.md` does not match `sources/`, and
// that is deliberate rather than incidental: that directory holds vendored
// snapshots of other people's terms — the ABA checklist's, Wiley's,
// iNaturalist's — kept for hashing and provenance. They are quoted under
// fair dealing for a record, not ours to republish, and serving them would
// be exactly the redistribution T-004 exists to prevent.
package river

import "embed"

//go:embed *.md
var Files embed.FS
