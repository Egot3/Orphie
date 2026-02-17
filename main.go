package main

import (
	"log"
	_ "net/http/pprof"
	diacon "newsgetter/dialynConnection"
	"newsgetter/internal/manager"
	"newsgetter/internal/types"
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

	go diacon.Connect(mgr.Get().Service.RabbitMQPort)

	workerManager := manager.NewWorkerManager(mgr)
	log.Println("Setting onReload")
	mgr.OnReload = func(old, new *types.ServiceStruct) {
		workerManager.Reconcile(old, new)
	}
	log.Println("Set onReload")

	mgr.Load()

	select {}
}
