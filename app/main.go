package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func count(wg *sync.WaitGroup, start, end int) {
	defer wg.Done()
	for i := start; i <= end; i++ {
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	const total = 10_000_000
	const goroutines = 5
	chunk := total / goroutines

	t := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		start := i*chunk + 1
		end := start + chunk - 1
		go count(&wg, start, end)
	}
	wg.Wait()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\nExecution time: %s\n", time.Since(t))
	log.Printf("OK, Execution time: %s\n", time.Since(t))
}

func main() {
	containerID, _ := os.Hostname()
	log.SetFlags(log.LstdFlags)
	log.SetPrefix(fmt.Sprintf("[%s] ", containerID))

	http.HandleFunc("/count", handler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		fmt.Fprintf(w, "Handled by: %s\n", hostname)
		fmt.Fprint(w, "OK")
	})
	log.Println("Listening on :3618")
	http.ListenAndServe(":3618", nil)
}
