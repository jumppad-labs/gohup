//go:build windows

package gohup

import (
	"os"
	"syscall"
)

func SetSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}

func Kill(pid *os.Process) {
	pid.Kill()
}
