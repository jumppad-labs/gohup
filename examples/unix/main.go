package main

import (
	"fmt"
	"time"

	"github.com/shipyard-run/gohup"
)

func main() {
	lp := &gohup.LocalProcess{}
	o := gohup.Options{
		Path: "sleep",
		Args: []string{
			"5",
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
