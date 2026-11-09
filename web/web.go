// Package web embeds the web UI.
package web

import "embed"

//go:embed root.gohtml
var FS embed.FS
