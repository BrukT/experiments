package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/brukt/experiments/coap-client-mapper/twin"
)

func main() {
	terminate := false

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)
	terminateDevice, terminateTwin := make(chan bool, 1), make(chan bool, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go twin.Start(terminateTwin, &wg)

	for {
		select {
		case <-sigc:
			terminate = true
		case <-time.After(time.Second):
		}
		if terminate {
			terminateDevice <- true
			terminateTwin <- true
			wg.Wait() // wait the other modules before terminating
			break
		}
	}
}
