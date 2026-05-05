package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
)

const (
	serviceName = "demo_app"
	maxReplicas = 10
	minReplicas = 1
)

var (
	mu      sync.Mutex
	pending string // "up" | "down" | ""
	timer   *time.Timer
)

func getReplicas() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return 0, err
	}
	defer cli.Close()
	res, err := cli.ServiceInspect(ctx, serviceName, client.ServiceInspectOptions{})
	if err != nil {
		return 0, err
	}
	return *res.Service.Spec.Mode.Replicated.Replicas, nil
}

func setReplicas(n uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return err
	}
	defer cli.Close()
	res, err := cli.ServiceInspect(ctx, serviceName, client.ServiceInspectOptions{})
	if err != nil {
		return err
	}
	spec := res.Service.Spec
	spec.Mode.Replicated = &swarm.ReplicatedService{Replicas: &n}
	_, err = cli.ServiceUpdate(ctx, serviceName, client.ServiceUpdateOptions{
		Version: res.Service.Meta.Version,
		Spec:    spec,
	})
	return err
}

func schedule(direction string) {
	mu.Lock()
	defer mu.Unlock()

	if pending == direction {
		log.Printf("[SCHEDULE] already pending %s, timer reset", direction)
		timer.Reset(30 * time.Second)
		return
	}

	if timer != nil {
		timer.Stop()
	}
	pending = direction
	log.Printf("[SCHEDULE] will scale-%s in 30s", direction)

	timer = time.AfterFunc(30*time.Second, func() {
		mu.Lock()
		dir := pending
		pending = ""
		mu.Unlock()

		cur, err := getReplicas()
		if err != nil {
			log.Printf("[ERROR] get replicas: %v", err)
			return
		}
		var next uint64
		if dir == "up" {
			if cur >= maxReplicas {
				log.Printf("[SKIP] already at maximum replicas (%d)", maxReplicas)
				return
			}
			next = cur + 1
		} else {
			if cur <= minReplicas {
				log.Printf("[SKIP] already at minimum replicas (%d)", minReplicas)
				return
			}
			next = cur - 1
		}
		log.Printf("[SCALE] %s: %d → %d", dir, cur, next)
		if err := setReplicas(next); err != nil {
			log.Printf("[ERROR] scale-%s: %v", dir, err)
		}
	})
}

func main() {
	log.Println("Scaler starting on :3619")
	log.Println("Succesfully create scheduler")
	http.HandleFunc("/scale-up", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[WEBHOOK] /scale-up")
		schedule("up")
		fmt.Fprint(w, "Scale-up scheduled in 30s")
	})

	http.HandleFunc("/scale-down", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[WEBHOOK] /scale-down")
		schedule("down")
		fmt.Fprint(w, "Scale-down scheduled in 30s")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})

	http.ListenAndServe(":3619", nil)
}
