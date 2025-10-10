//go:build windows
// +build windows

package main

import (
	"os"
	"os/signal"
)

func setupSignalHandler() chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	return stop
}
