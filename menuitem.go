package remotray

type MenuItem struct {
	tray    *SysTray
	id      int
	title   string
	tooltip string
}

func (mi MenuItem) OnClick(f func(item MenuItem)) {
	mi.tray.lock.Lock()
	defer mi.tray.lock.Unlock()
	mi.tray.clickEventCallback[mi.id] = f
}

func (mi MenuItem) Title() string {
	return mi.title
}

func (mi MenuItem) Tooltip() string {
	return mi.tooltip
}
