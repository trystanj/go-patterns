package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"context"
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

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	s := &http.Server{Addr: ":3000", Handler: nil}

	// kick off the server in another goroutine so we can listen for signals in main
	go func() {
		log.Fatal(s.ListenAndServe())
	}()

	log.Println("Waiting for signal...")

SigListen:
	for {
		select {
		case s := <-sig:
			log.Printf("Got signal: %v", s)
			log.Println("Closing done channel")

			close(done) // signal everyone
			wg.Wait()   // wait for them to tell us they're truly done before continuing

			log.Println("All goroutines done; exiting loop")
			break SigListen
		}
	}

	// This could be in the signal case above, but clean-up is more clear here
	log.Println("Sending shutdown (5s grace period)...")

	// Note that this doesn't help with "fancy" connections like Websockets; those are up to use to close up
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	s.Shutdown(ctx)

	log.Println("Server shutdown complete; terminating")
}
