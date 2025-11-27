//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"
)

// setupSignalHandler 设置信号处理器（Unix/Linux/macOS）
func setupSignalHandler(sigChan chan os.Signal) {
	signal.Notify(sigChan, syscall.SIGWINCH)
}
