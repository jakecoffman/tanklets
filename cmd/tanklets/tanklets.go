package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"github.com/jakecoffman/tanklets/client"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	runtime.LockOSThread()

	// Dumps goroutines when ctl-c to help figure out what is wrong
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
			os.Exit(1)
		}
	}()

	client.Loop()
}
