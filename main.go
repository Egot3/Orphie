package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"newsgetter/internal/manager"
	"newsgetter/internal/types"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	const configPath = "config.toml"

	log.Println("creating cfgmgr")
	mgr, err := manager.NewManager(configPath, nil)
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}
	defer mgr.Stop()
	log.Println("created cfgmgr")

	workerManager := manager.NewWorkerManager(mgr)
	log.Println("Setting onReload")
	mgr.OnReload = func(old, new *types.ServiceStruct) {
		workerManager.Reconcile(old, new)
	}
	log.Println("Set onReload")
	mgr.Load()

	select {}
}
