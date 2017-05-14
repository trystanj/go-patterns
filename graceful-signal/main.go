package main

import (
	"log"
	"os"
	"sync"
	"time"

	"os/signal"
	"syscall"
)

func spin(i int, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done() // tell the parent group this routine is done

	log.Println("Starting goroutine: ", i)

	for {
		select {
		case <-done:
			log.Println("Exiting goroutine: ", i)
			time.Sleep(time.Second * 1)
			return
		}
	}
}

func main() {

	// use an empty struct as our signal since they consume no storage: https://dave.cheney.net/2014/03/25/the-empty-struct
	done := make(chan struct{}, 1)

	// channel to listen for system signals
	sig := make(chan os.Signal, 1)

	// register our channel as the place we want signals sent
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// create a WaitGroup so we can track when all our goroutines have exited
	wg := &sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)            // add to our counter
		go spin(i, done, wg) // kick them off
	}

	log.Println("Waiting for signal...")
	for {
		select {
		case s := <-sig:
			log.Printf("Got signal: %v", s)
			log.Println("Closing done channel")
			close(done) // signal everyone

			wg.Wait() // wait for them to tell us they're truly done before continuing
			log.Println("Exiting main")
			return
		}
	}
}
