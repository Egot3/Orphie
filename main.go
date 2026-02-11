package main

import (
	"log"
	"newsgetter/internal/manager"
	"newsgetter/internal/types"
)

func main() {
	const configPath = "config.toml"

	workerManager := manager.NewWorkerManager()

	mgr, err := manager.NewManager(configPath, func(old, new *types.ServiceStruct) {
		workerManager.Reconcile(old, new)
	})
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}
	defer mgr.Stop()
	select {}
}
