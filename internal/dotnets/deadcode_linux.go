//go:build linux

package dotnets

import _ "embed"

//go:embed deadcode_linux_amd64
var analyzerBinary []byte
