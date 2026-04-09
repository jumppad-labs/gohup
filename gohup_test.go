package gohup

import (
	"fmt"
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
	testFolder, _ := os.MkdirTemp("", "")
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

func Test_QueryStatusReturnsNotFoundWhenPidNotExist(t *testing.T) {
	status, err := (&LocalProcess{}).QueryStatus("/tmp/doesnotexist.pid")

	require.NoError(t, err)
	require.Equal(t, status, StatusNotFound)
}

func Test_StartsAProcessInBackgroundAndLogOutput(t *testing.T) {
	dir := t.TempDir()

	o := Options{
		Path:    "echo",
		Args:    []string{"Hello World"},
		Logfile: path.Join(dir, "file.log"),
	}

	lp := &LocalProcess{}
	pid, pidfile, err := lp.Start(o)

	require.Eventuallyf(t, func() bool {
		s, err := lp.QueryStatus(pidfile)
		fmt.Println(s, err)
		return s == StatusStopped
	}, 5*time.Second, 1*time.Second, "process failed to stop")

	require.NoError(t, err)
	require.Greater(t, pid, 1)

	d, err := os.ReadFile(o.Logfile)
	require.NoError(t, err)
	require.Equal(t, "Hello World\n", string(d))
}

func Test_StartsAProcessAndCreatesPIDFileWithDefaults(t *testing.T) {
	pid, pidfile, err := start(t, nil)
	require.NoError(t, err)

	d, err := os.ReadFile(pidfile)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%d", pid), string(d))
}

func Test_StartsAProcessAndCreatesPIDFileWithCustom(t *testing.T) {
	pidfile := tempPidfile(t)
	options := opts
	options.Pidfile = pidfile

	pid, _, err := start(t, &options)
	require.NoError(t, err)

	d, err := os.ReadFile(pidfile)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%d", pid), string(d))
}

func Test_StartsAProcessAndCreatesPIDFileInPidDir(t *testing.T) {
	pidDir := t.TempDir()
	options := opts
	options.PidDir = pidDir

	pid, pidfile, err := start(t, &options)
	require.NoError(t, err)

	expected := path.Join(pidDir, fmt.Sprintf("%d.pid", pid))
	require.Equal(t, expected, pidfile)

	d, err := os.ReadFile(pidfile)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%d", pid), string(d))
}

func Test_StartsAProcessWithPidDirAndRelativePidfile(t *testing.T) {
	pidDir := t.TempDir()
	options := opts
	options.PidDir = pidDir
	options.Pidfile = "myprocess.pid"

	pid, pidfile, err := start(t, &options)
	require.NoError(t, err)

	expected := path.Join(pidDir, "myprocess.pid")
	require.Equal(t, expected, pidfile)

	d, err := os.ReadFile(pidfile)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%d", pid), string(d))
}

func Test_StartsAProcessWithPidDirAndAbsolutePidfile(t *testing.T) {
	pidDir := t.TempDir()
	absDir := t.TempDir()
	absPidfile := path.Join(absDir, "absolute.pid")

	options := opts
	options.PidDir = pidDir
	options.Pidfile = absPidfile

	pid, pidfile, err := start(t, &options)
	require.NoError(t, err)

	// absolute Pidfile should be used as-is, PidDir ignored
	require.Equal(t, absPidfile, pidfile)

	d, err := os.ReadFile(pidfile)
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
	os.WriteFile(pidfile, []byte("99999"), os.ModePerm)

	lp := &LocalProcess{}
	st, err := lp.QueryStatus(options.Pidfile)

	require.NoError(t, err)
	require.Equal(t, StatusStopped, st)
}
