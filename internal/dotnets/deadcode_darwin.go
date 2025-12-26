//go:build darwin

package dotnets

import _ "embed"

//go:embed deadcode_darwin_amd64
var analyzerBinary []byte
