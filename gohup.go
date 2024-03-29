package gohup

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	stats "github.com/guillermo/go.procstat"
	ps "github.com/mitchellh/go-ps"
)

// Status of the process
type Status string

// StatusRunning is returned when the process is Running
const StatusRunning Status = "StatusRunning"

// StatusStopped is returned when the process is Stopped
const StatusStopped Status = "StatusStopped"

// StatusError is returned when it is not possible to determine the status of a process
const StatusError Status = "StatusError"

// StatusNotFound is returned when the pid does not exist
const StatusNotFound Status = "StatusNotFound"

// Options to be used when starting a process
type Options struct {
	// Path of the process to start
	Path string
	// Arguments to pass when starting the process
	Args []string
	// Environments to pass when starting the process
	Env []string
	// Directory to start the process in, default is the current directory
	Dir string
	// File to write the running process ID to, if blank a temporary file
	// will be created.
	Pidfile string
	// File to write output from the started command to, if blank logs
	// will be disgarded
	Logfile string
}

// Process defines an interface which defines methods
// for managing a child process.
type Process interface {
	// Start a process in the background
	Start(options Options) (int, string, error)
	// Stop the current running process
	Stop(pidfile string, signal syscall.Signal) error
	// Return the status of the currently running process
	QueryStatus(pidfile string) (Status, error)
}

// LocalProcess is the implementation of the Process interface.
type LocalProcess struct {
	options Options
}

// Start a local process in the background with the given options
// if the process does not start an error is returned.

// Start returns the process ID and the pidfile for the running process.
// If a process can not be started then an error will be returned.
func (l *LocalProcess) Start(options Options) (int, string, error) {
	// Use the current directory unless Dir has been specified
	if options.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return -1, "", err
		}

		options.Dir = cwd
	}

	cmd := exec.Command(options.Path, options.Args...)
	cmd.Dir = options.Dir
	cmd.Env = options.Env
	cmd.SysProcAttr = SetSysProcAttr()

	// create a logfile and redirect std error and std out
	if options.Logfile != "" {
		f, err := os.Create(options.Logfile)
		if err != nil {
			return -1, "", fmt.Errorf("Unable to open log file: %s", err)
		}
		cmd.Stderr = f
		cmd.Stdout = f
	}

	// start the process
	err := cmd.Start()
	if err != nil {
		return -1, "", err
	}
	pid := cmd.Process.Pid
	cmd.Process.Release()

	// If no pidfile is specified create a pid file in the
	// temporary directory
	if options.Pidfile == "" {
		options.Pidfile = filepath.Join(os.TempDir(), fmt.Sprintf("%d.%s", pid, "pid"))
	}

	// write the pid file
	err = l.writePidFile(pid, options.Pidfile)
	if err != nil {
		return -1, "", err
	}

	return pid, options.Pidfile, nil
}

// Stop the process referenced by the PID in the given file.
func (l *LocalProcess) Stop(pidfile string) error {
	pid, err := l.readPidFile(pidfile)
	if err != nil {
		return err
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// kill all the processes in the process group
	Kill(p)

	err = os.Remove(pidfile)
	if err != nil {
		return err
	}

	return nil
}

// QueryStatus of the backgrounded process referenced by the PID in the given file.
//
// If the process is not running StatusStopped, and a nil error will be returned.
// If the process is running StatusRunning, and a nil error will be returned.
// If it is not possible to query the status of the process or if the pidfile
// is not readable. StatusError, and an error will be returned.
func (l *LocalProcess) QueryStatus(pidfile string) (Status, error) {
	pid, err := l.readPidFile(pidfile)
	if err != nil {
		return StatusNotFound, nil
	}

	p, err := ps.FindProcess(pid)
	if err != nil {
		return StatusError, err
	}

	if p == nil {
		return StatusStopped, nil
	}

	if runtime.GOOS == "linux" {
		// check the process is Zombie on Linux
		stats := stats.Stat{Pid: pid}
		err = stats.Update()
		if err != nil {
			return StatusError, err
		}

		if stats.State == 90 {
			return StatusStopped, nil
		}
	} else if runtime.GOOS == "darwin" {
		// check the process is Zombie on Darwin
		output, _ := exec.Command("ps", fmt.Sprintf("%d", p.Pid())).CombinedOutput()
		if strings.Contains(string(output), "<defunct>") {
			return StatusStopped, nil
		}
	}

	return StatusRunning, nil
}

func (l *LocalProcess) writePidFile(pid int, pidfile string) error {

	f, err := os.Create(pidfile)
	if err != nil {
		return fmt.Errorf("unable to create pid file: %s", err)
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("%d", pid))

	return nil
}

func (l *LocalProcess) readPidFile(pidfile string) (int, error) {
	d, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return -1, fmt.Errorf("error reading file: %s", err)
	}

	pid, err := strconv.ParseInt(string(d), 10, 64)
	if err != nil {
		return -1, fmt.Errorf("unable to cast pid to integer: %s", err)
	}

	return int(pid), nil
}
