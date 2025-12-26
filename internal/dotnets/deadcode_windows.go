//go:build windows

package dotnets

import _ "embed"

//go:embed deadcode_windows_amd64.exe
var analyzerBinary []byte
