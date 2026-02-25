package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"orphie/internal/manager"
	"orphie/internal/middleware"
	"orphie/internal/types"
	"os"

	banye "github.com/Egot3/Banye"
	diacon "github.com/Egot3/Zhao"
	"github.com/Egot3/Zhao/pub"
)

func main() {

	const configPath = "config.toml"

	log.Println("creating cfgmgr")
	mgr, err := manager.NewManager(configPath, nil)
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}
	defer mgr.Stop()
	log.Println("created cfgmgr")

	mgr.Load()

	cfg := diacon.RabbitMQConfiguration{
		URL:  os.Getenv("RABBIT_URL"),
		Port: os.Getenv("RABBIT_PORT"),
	}
	conn, err := diacon.Connect(cfg)
	if err != nil {
		log.Panicf("Couldn't connect to Rabbitmq: %v", err)
	}
	defer conn.Close()

	publisher, err := pub.NewPublisher(conn)
	if err != nil {
		log.Panicf("Couldn't create a Publisher: %v", err)
	}
	defer publisher.Close()

	before, after := middleware.TraceTripperMiddleware()
	client := banye.NewClient(http.DefaultClient)
	client.UseTripper(before, after)

	workerManager := manager.NewWorkerManager(mgr, client, publisher)
	log.Println("Setting onReload")
	mgr.OnReload = func(old, new *types.Config) {
		workerManager.Reconcile(old, new)
	}

	mgr.Load()

	select {}
}
