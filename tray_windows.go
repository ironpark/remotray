package remotray

import (
	_ "embed"
)

//go:embed tray.exe
var trayBinary []byte
