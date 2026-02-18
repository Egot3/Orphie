package main

import (
	"context"
	"log"
	_ "net/http/pprof"
	"orphie/internal/manager"
	"orphie/internal/types"
	"time"

	diacon "github.com/Egot3/Zhao"
	"github.com/Egot3/Zhao/pub"
	"github.com/Egot3/Zhao/queues"
	"github.com/rabbitmq/amqp091-go"
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
	q, err := queues.NewQueue(publisher.Ch, qStruct)
	if err != nil {
		log.Panicf("Couldn't create a queue: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = publisher.Publish(ctx, "", q.Name, false, false, amqp091.Publishing{
		ContentType: "text/plain",
		Body:        []byte("DINGDINGDING"),
	})
	if err != nil {
		log.Panicf("Couldn't publish: %v", err)
	}

	workerManager := manager.NewWorkerManager(mgr)
	log.Println("Setting onReload")
	mgr.OnReload = func(old, new *types.ServiceStruct) {
		workerManager.Reconcile(old, new)
	}
	log.Println("Set onReload")

	mgr.Load()

	select {}
}
