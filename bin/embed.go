package bin

import "embed"

//go:embed paqet_linux_amd64 paqet_linux_arm64
var PaqetBinFS embed.FS
