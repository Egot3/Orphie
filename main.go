package main

import (
	"log"
	_ "net/http/pprof"
	"orphie/internal/manager"
	"orphie/internal/types"

	diacon "github.com/Egot3/Zhao"
	"github.com/Egot3/Zhao/pub"
	"github.com/Egot3/Zhao/queues"
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

	//go diacon.Connect(mgr.Get().Service.RabbitMQPort)

	cfg := diacon.RabbitMQConfiguration{
		URL:  "amqp://guest:guest@localhost",
		Port: "1130",
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

	qStruct := queues.QueueStruct{
		Name:           "RINGRINGRING",
		Durable:        false,
		DeleteOnUnused: false,
		Exclusive:      false,
		NoWait:         false,
		Args:           nil,
	}
	_, err = queues.NewQueue(publisher.Ch, qStruct)
	if err != nil {
		log.Panicf("Couldn't create a queue: %v", err)
	}

	workerManager := manager.NewWorkerManager(mgr, publisher)
	log.Println("Setting onReload")
	mgr.OnReload = func(old, new *types.ServiceStruct) {
		workerManager.Reconcile(old, new)
	}
	log.Println("Set onReload")

	mgr.Load()

	select {}
}
