# Systray for Wails

POC

```go
func RunTray(){
    tray, err := systray.Run("<unix-domain-socket | pipe name for IPC>",
        systray.WithTitle("TITTLE"),)
    if err != nil {
        panic(err)
    }
    item, _ := tray.AddMenuItem("item title.1", "Tooltop")
    item.OnClick(func(item MenuItem) {
        fmt.Println("onclick", item)
    })
    item, _ = tray.AddMenuItem("quit", "Tooltop")
    item.OnClick(func(item MenuItem) {
        tray.Quit()
    })
}

func main()  {
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs
}
```