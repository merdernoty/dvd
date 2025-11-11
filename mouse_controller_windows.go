//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procSetCursorPos     = user32.NewProc("SetCursorPos")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
)

const (
	smCXSCREEN = 0
	smCYSCREEN = 1
)

type point struct {
	X int32
	Y int32
}

func getScreenSize() (int, int) {
	width, _, _ := procGetSystemMetrics.Call(uintptr(smCXSCREEN))
	height, _, _ := procGetSystemMetrics.Call(uintptr(smCYSCREEN))
	return int(width), int(height)
}

func moveMouse(x, y int) {
	procSetCursorPos.Call(uintptr(x), uintptr(y))
}

func getMousePos() (int, int) {
	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y)
}
