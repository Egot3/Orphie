package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	pb "newsgetter/contracts"
	"newsgetter/internal/manager"
	"newsgetter/internal/reqresp"
	"newsgetter/internal/types"

	"google.golang.org/grpc"
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

	dataChan := make(chan *pb.OrphieDataResponse)

	workerManager := manager.NewWorkerManager(mgr, dataChan)
	log.Println("Setting onReload")
	mgr.OnReload = func(old, new *types.ServiceStruct) {
		workerManager.Reconcile(old, new)
	}
	log.Println("Set onReload")
	mgr.Load()

	port := fmt.Sprintf(":%v", mgr.Get().Service.Port)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Panicln("couldn't start a listener: ", err) //may recover
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrphieServiceServer(grpcServer, reqresp.NewOrphieStreamServer(dataChan))

	log.Println("listenin'n'serving ", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Panicln("Couldn't serve a listener: ", err)
	}

	select {}
}
