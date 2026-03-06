// Package skilldata embeds the dtasks-cli SKILL.md for in-binary distribution.
package skilldata

import _ "embed"

//go:embed SKILL.md
var Content []byte
