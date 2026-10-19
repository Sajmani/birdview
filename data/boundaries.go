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

// Package data embeds the committed generated files the server serves as
// they are. The tables under this directory that only the ingest reads are
// read from disk and are not embedded.
package data

import _ "embed"

// Boundaries is region-boundaries.json, written by cmd/birdview-boundaries
// (tech.regionBoundaries).
//
//go:embed region-boundaries.json
var Boundaries []byte
