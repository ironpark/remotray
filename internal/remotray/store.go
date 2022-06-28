package main

import (
	"fyne.io/systray"
	"remotray/internal"
	"sync"
)

type Store struct {
	menuItemIdIncrement int
	lock                sync.RWMutex
	sysMenu             map[int]*systray.MenuItem
	menu                map[int]internal.MenuItem
}

func (s *Store) S() {

}
