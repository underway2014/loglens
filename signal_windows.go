//go:build windows
// +build windows

package main

import (
	"os"
)

// setupSignalHandler 设置信号处理器（Windows）
// Windows 不支持 SIGWINCH，所以这里是空实现
func setupSignalHandler(sigChan chan os.Signal) {
	// Windows 不支持窗口大小变化信号
	// 这里保持空实现，不监听任何信号
}
