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
package docs

import "embed"

//go:embed *.md
var Files embed.FS
