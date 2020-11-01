# GoHUP

[![PkgGoDev](https://pkg.go.dev/badge/github.com/shipyard-run/gohup)](https://pkg.go.dev/github.com/shipyard-run/gohup)

GoHup allows you to run a long running command in a detached background process.
It does not a daemon runner which monitors the status of the process restarting when necessary,
GoHup only starts the process and returns the process id.

GoHup is similar to the Unix command `nohup`, except GoHup automatically creates a PID file and
redirects log output to a file. You can use GoHup in your own applications, commands started
with GoHup will continue to run after the application using the library has exited.

## Usage

To create and start a new process create a new instance of `gohup.LocalProcess` and call the 
`Start` method with the process options.

The following exanple will start the command `tail` in the background and return the pid
and the pidfile containing the pid.

```golang
lp := &gohup.LocalProcess{}
o := gohup.Options{
	Path: "/usr/bin/tail",
	Args: []string{
		"-f",
		"/dev/null",
	},
  Logfile: "./process.log",
}

pid, pidfile, err := lp.Start(o)
if err != nil {
	panic(err)
}

fmt.Printf("Started PID: %d, PID file: %s\n", pid, pidfile)
```

To query the status of a process, you can use the `QueryStatus` method 
passing the location of the pidfile. `QueryStatus` will return a string,
`gohup.StatusRunning` or `gohup.StatusStopped` depending on the state
of the process.

```golang
s, err := lp.QueryStatus(pidfile)
if err != nil {
	panic(err)
}

fmt.Println("Status:", s)
```

To stop a backgrounded process you can call the `Stop` method passing the location of a 
pidfile.

GoHup removes the pidfile if `Stop` is successful.

```golang
fmt.Println("Stopping process")
err = lp.Stop(pidfile)
if err != nil {
	panic(err)
}
```
