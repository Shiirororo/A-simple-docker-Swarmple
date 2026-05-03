package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
)

func scaleService(serviceName string, targetReplicas uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cli, err := client.New(client.FromEnv)
	if err != nil {
		return fmt.Errorf("Error Docker Client %v", err)
	}
	defer cli.Close()

	log.Printf("Check service: %s", serviceName)
	res, err := cli.ServiceInspect(ctx, serviceName, client.ServiceInspectOptions{})
	if err != nil {
		return fmt.Errorf("Service not found '%s' : %v", serviceName, err)
	}
	currReplicas := *res.Service.Spec.Mode.Replicated.Replicas
	log.Printf("Current replicas: %d, Target: %d", currReplicas, targetReplicas)
	if currReplicas == targetReplicas {
		log.Printf("Replicas is maxed")
		return nil
	}
	newSpec := res.Service.Spec
	newSpec.Mode.Replicated = &swarm.ReplicatedService{
		Replicas: &targetReplicas,
	}
	log.Printf("Sending request to scale to %d replicas...", targetReplicas)
	_, err = cli.ServiceUpdate(ctx, serviceName, client.ServiceUpdateOptions{
		Version: res.Service.Meta.Version,
		Spec:    newSpec,
	})
	if err != nil {
		return fmt.Errorf("Scale failed: %v", err)
	}
	return nil
}
func autoScaler()

func cadVisor()

func main() {
	log.Println("Scaler is starting...")

	http.HandleFunc("/scale-up", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[WEBHOOK] received request /scale-up")
		err := scaleService("demo_app", 5)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
		log.Println("[SUCCESS] scaled up to 5 replicas!")
		fmt.Fprint(w, "Scale up sucessfully!")

	})

	http.HandleFunc("/scale-down", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[WEBHOOK] received request /scale-down")
		err := scaleService("demo_app", 1)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
		log.Println("[SUCCESS] scaled down to 1 replica!")
		fmt.Fprint(w, "Scale down sucessfully!")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	})
	fmt.Println("Listening on :3619")
	http.ListenAndServe(":3619", nil)
}
