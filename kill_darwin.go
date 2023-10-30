//go:build darwin
package gohup

import (
	"os"
	"syscall"
)

func SetSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

func Kill(pid *os.Process) {
	syscall.Kill(-pid.Pid, syscall.SIGKILL)
}