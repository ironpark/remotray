package remotray

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed tray.exe
var trayBinary []byte
var installPath = filepath.Join(os.TempDir(), "remotray.exe")
