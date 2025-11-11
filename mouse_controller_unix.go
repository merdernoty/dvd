//go:build !windows

package main

import "github.com/go-vgo/robotgo"

func getScreenSize() (int, int) {
	return robotgo.GetScreenSize()
}

func moveMouse(x, y int) {
	robotgo.MoveMouse(x, y)
}

func getMousePos() (int, int) {
	return robotgo.GetMousePos()
}
