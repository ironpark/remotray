package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"fyne.io/systray"
	"github.com/ironpark/remotray/internal"
	"github.com/ironpark/remotray/internal/ipc"
	"sync"
)

var globalMenuItemId = 0
var menuClickCh = make(chan int, 10)
var menuItems = make(map[int]*systray.MenuItem)
var menuItemCache = make(map[int]internal.MenuItem)

var lock sync.RWMutex
var (
	ipcName = flag.String("ipc", "", "domain socket or pipe name for ipc")
	icon    = flag.String("icon", "", "base64 data")
	title   = flag.String("title", "", "a string")
	tooltip = flag.String("tooltip", "", "a string")
)

func main() {
	flag.Parse()
	systray.Run(onReady, onExit)
}

func incrementId() int {
	lock.Lock()
	defer lock.Unlock()
	globalMenuItemId++
	return globalMenuItemId
}

func addMenuItem(title, tooltip string) int {
	id := incrementId()
	sysItem := systray.AddMenuItem(title, tooltip)
	lock.Lock()
	menuItems[id] = sysItem
	menuItemCache[id] = internal.MenuItem{
		Id:      id,
		Title:   title,
		Tooltip: tooltip,
	}
	lock.Unlock()
	go func(click chan struct{}, itemId int) {
		for {
			<-click
			menuClickCh <- itemId
		}
	}(sysItem.ClickedCh, globalMenuItemId)
	return id
}

func onReady() {
	defer fmt.Println("Ready For IPC")
	if icon != nil && *icon != "" {
		iconData, _ := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(*icon)
		systray.SetIcon(iconData)
	}
	if title != nil && *title != "" {
		systray.SetTitle(*title)
	}
	if tooltip != nil && *tooltip != "" {
		systray.SetTooltip(*tooltip)
	}
	ipcServer, err := ipc.NewServer(*ipcName)
	if err != nil {
		panic(err)
	}
	ipcServer.SetMessageProcessor(internal.MsgTypeAddMenuItem, func(data []byte) (interface{}, error) {
		item := internal.MenuItem{}
		_ = json.Unmarshal(data, &item)
		item.Id = addMenuItem(item.Title, item.Tooltip)
		return item, nil
	})

	ipcServer.SetMessageProcessor(internal.MsgTypeSetIcon, func(data []byte) (interface{}, error) {
		var iconData string
		json.Unmarshal(data, &title)
		data, _ = base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(iconData)
		systray.SetIcon(data)
		return nil, nil
	})

	ipcServer.SetMessageProcessor(internal.MsgTypeSetTitle, func(data []byte) (interface{}, error) {
		var title string
		json.Unmarshal(data, &title)
		systray.SetTitle(title)
		return nil, nil
	})
	ipcServer.SetMessageProcessor(internal.MsgTypeSetTooltip, func(data []byte) (interface{}, error) {
		var title string
		json.Unmarshal(data, &title)
		systray.SetTooltip(title)
		return nil, nil
	})
	ipcServer.SetMessageProcessor(internal.MsgTypeQuit, func(data []byte) (interface{}, error) {
		systray.Quit()
		return nil, nil
	})

	go ipcServer.Run()
	go func(menuClickCh chan int) {
		for menuItemId := range menuClickCh {
			lock.RLock()
			ipcServer.EventEmit(internal.MsgTypeOnClick, menuItemCache[menuItemId])
			lock.RUnlock()
		}
	}(menuClickCh)
}

func onExit() {
	// clean up here
}
