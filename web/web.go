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
