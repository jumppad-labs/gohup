package gohup

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var opts = Options{
	Path: "/usr/bin/tail",
	Args: []string{
		"-f",
		"/dev/null",
	},
}

func setup() (*LocalProcess, Options) {
	lp := &LocalProcess{}

	return lp, opts
}

func start(t *testing.T, options *Options) (int, string, error) {
	lp, o := setup()

	if options != nil {
		o = *options
	}

	pid, pidfile, err := lp.Start(o)
	require.NoError(t, err)
	require.Greater(t, pid, 1)

	// check the process is running
	p, err := os.FindProcess(pid)
	require.NoError(t, err)

	t.Cleanup(func() {
		p.Kill()
		os.Remove(pidfile)
	})

	return pid, pidfile, err
}

func tempPidfile(t *testing.T) string {
	testFolder, _ := ioutil.TempDir("", "")
	t.Cleanup(func() {
		os.RemoveAll(testFolder)
	})

	return path.Join(testFolder, "mypid.pid")
}

func Test_StartsAProcessInBackground(t *testing.T) {
	pid, _, err := start(t, nil)

	require.NoError(t, err)
	require.Greater(t, pid, 1)
}

func Test_StartsAProcessInBackgroundAndLogOutput(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")

	o := Options{
		Path:    "echo",
		Args:    []string{"Hello World"},
		Logfile: path.Join(dir, "file.log"),
	}

	lp := &LocalProcess{}
	pid, _, err := lp.Start(o)
	time.Sleep(10 * time.Millisecond)

	require.NoError(t, err)
	require.Greater(t, pid, 1)

	d, err := ioutil.ReadFile(o.Logfile)
	require.NoError(t, err)
	require.Equal(t, "Hello World\n", string(d))
}

func Test_StartsAProcessAndCreatesPIDFileWithDefaults(t *testing.T) {
	pid, pidfile, err := start(t, nil)

	d, err := ioutil.ReadFile(pidfile)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%d", pid), string(d))
}

func Test_StartsAProcessAndCreatesPIDFileWithCustom(t *testing.T) {
	pidfile := tempPidfile(t)
	options := opts
	options.Pidfile = pidfile

	pid, _, err := start(t, &options)

	d, err := ioutil.ReadFile(pidfile)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%d", pid), string(d))
}

func Test_StopsAProcessAndDeletesPidFile(t *testing.T) {
	pidfile := tempPidfile(t)
	options := opts
	options.Pidfile = pidfile

	pid, _, err := start(t, &options)
	require.NoError(t, err)

	lp := &LocalProcess{}
	err = lp.Stop(options.Pidfile)
	require.NoError(t, err)

	_, err = os.FindProcess(pid)
	require.NoError(t, err)

	_, err = os.Stat(options.Pidfile)
	require.Error(t, err)
}

func Test_StatusRunningWhenRunning(t *testing.T) {
	pidfile := tempPidfile(t)
	options := opts
	options.Pidfile = pidfile

	_, _, err := start(t, &options)
	require.NoError(t, err)

	lp := &LocalProcess{}
	st, err := lp.QueryStatus(options.Pidfile)

	require.NoError(t, err)
	require.Equal(t, StatusRunning, st)
}

func Test_StatusStoppedWhenStopped(t *testing.T) {
	pidfile := tempPidfile(t)
	options := opts
	options.Pidfile = pidfile
	ioutil.WriteFile(pidfile, []byte("99999"), os.ModePerm)

	lp := &LocalProcess{}
	st, err := lp.QueryStatus(options.Pidfile)

	require.NoError(t, err)
	require.Equal(t, StatusStopped, st)
}
