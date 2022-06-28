package remotray

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestRun(t *testing.T) {
	tray, err := Run("sssdddssdsdsdsdsd", WithTitle("ads"), WithTooltip("AAS"))
	tray.SetTitle("ASD")
	if err != nil {
		panic(err)
	}
	item, _ := tray.AddMenuItem("Quit", "s2s")
	item.OnClick(func(item MenuItem) {
		tray.Quit()
	})

	item, _ = tray.AddMenuItem("OH MY2", "s2s")
	item.OnClick(func(item MenuItem) {
		fmt.Println(item)
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
