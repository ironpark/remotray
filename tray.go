package remotray

import (
	"bufio"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ironpark/remotray/internal"
	"github.com/ironpark/remotray/internal/ipc"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var installPath = filepath.Join(os.TempDir(), "remotray")

func install() {
	file, _ := os.OpenFile(installPath, os.O_RDWR|os.O_CREATE, 0755)
	file.Write(trayBinary)
	file.Close()
}

type Config struct {
	iconData []byte
	title    string
	tooltip  string
}

type SysTray struct {
	ipc                *ipc.Client
	cmd                *exec.Cmd
	lock               sync.Mutex
	clickEventCallback map[int]func(item MenuItem)
}

func (tray *SysTray) Quit() {
	tray.ipc.WriteMessage(internal.MsgTypeQuit, "")
	tray.cmd.Process.Wait()
}

func (tray *SysTray) SetTitle(title string) {
	tray.ipc.WriteMessage(internal.MsgTypeSetTitle, title)
}

func (tray *SysTray) SetIcon(data []byte) {
	tray.ipc.WriteMessage(internal.MsgTypeSetIcon, base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(data))
}

func (tray *SysTray) SetTooltip(tooltip string) {
	tray.ipc.WriteMessage(internal.MsgTypeSetTooltip, tooltip)
}

func (tray *SysTray) AddMenuItem(title, tooltip string) (MenuItem, error) {
	menuItem := internal.MenuItem{
		Title:   title,
		Tooltip: tooltip,
	}
	msgId, _ := tray.ipc.WriteMessage(internal.MsgTypeAddMenuItem, internal.MenuItem{
		Title:   title,
		Tooltip: tooltip,
	})
	err := tray.ipc.ReadReplyMessage(msgId, &menuItem)
	return MenuItem{
		tray:    tray,
		id:      menuItem.Id,
		title:   menuItem.Title,
		tooltip: menuItem.Tooltip,
	}, err
}

func Run(ipcName string, opt ...Option) (*SysTray, error) {
	install()
	cf := &Config{}
	for _, option := range opt {
		option(cf)
	}
	args := []string{"-ipc=" + ipcName}
	if cf.title != "" {
		args = append(args, fmt.Sprintf(`-title=%s`, cf.title))
	}
	if cf.tooltip != "" {
		args = append(args, fmt.Sprintf(`-tooltip=%s`, cf.tooltip))
	}
	if cf.iconData != nil {
		args = append(args, fmt.Sprintf(`-icon=%s`, base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(cf.iconData)))
	}
	cmd := exec.Command(installPath, args...)
	stdout, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	readyCh := make(chan bool)
	go func() {
		cmd.Wait()
		close(readyCh)
	}()
	go func(stdout io.ReadCloser) {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			message := scanner.Text()
			if strings.Contains(message, "Ready") {
				readyCh <- true
			}
		}
	}(stdout)

	if !<-readyCh {
		return nil, errors.New("fail")
	}
	client, err := ipc.NewClient(ipcName)
	if err != nil {
		return nil, err
	}

	systray := &SysTray{
		ipc:                client,
		cmd:                cmd,
		lock:               sync.Mutex{},
		clickEventCallback: map[int]func(item MenuItem){},
	}

	client.OnEvent(func(eventId int, data []byte) {
		switch eventId {
		case internal.MsgTypeOnClick:
			item := internal.MenuItem{}
			_ = json.Unmarshal(data, &item)
			systray.lock.Lock()
			ch := systray.clickEventCallback[item.Id]
			systray.lock.Unlock()
			if ch != nil {
				ch(MenuItem{
					tray:    systray,
					id:      item.Id,
					title:   item.Title,
					tooltip: item.Tooltip,
				})
			}
		}
	})

	return systray, nil
}
