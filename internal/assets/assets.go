package assets

import "embed"

//go:embed all:templates all:static
var fs embed.FS

var Templates = fs
var Static = fs
