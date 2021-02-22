package main

import (
	"fmt"
	"time"

	"github.com/nicholasjackson/gohup"
)

func main() {
	lp := &gohup.LocalProcess{}
	o := gohup.Options{
		Path: "powershell.exe",
		Args: []string{
			".\\test-script.ps1",
			"50000",
		},
	}

	pid, pidfile, err := lp.Start(o)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Started PID: %d, PID file: %s\n", pid, pidfile)

	fmt.Println("Stopping process in 30 seconds")
	for i := 0; i < 15; i++ {
		s, err := lp.QueryStatus(pidfile)
		if err != nil {
			panic(err)
		}

		fmt.Println("Status:", s)
		time.Sleep(2 * time.Second)
	}

	// Stop the running process
	fmt.Println("Stopping process")
	err = lp.Stop(pidfile)
	if err != nil {
		panic(err)
	}
}
