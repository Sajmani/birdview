// Package spec embeds the normative requirement documents so BirdView can
// serve them, the way package docs does for the user-facing guides.
//
// **Only the top-level documents.** `*.md` does not match `sources/`, and
// that is deliberate rather than incidental: that directory holds vendored
// snapshots of other people's terms — the ABA checklist's, Wiley's,
// iNaturalist's — kept for hashing and provenance. They are quoted under
// fair dealing for a record, not ours to republish, and serving them would
// be exactly the redistribution T-004 exists to prevent.
package spec

import "embed"

//go:embed *.md
var Files embed.FS
