package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	mapconverter "orphie/internal/mapConverter"
	"orphie/internal/types"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	banye "github.com/Egot3/Banye"
	pb "github.com/Egot3/Yidhari/contracts"
	"github.com/Egot3/Zhao/pub"
	"github.com/Egot3/Zhao/queues"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WorkerManager struct {
	cancelFuncs map[string]context.CancelFunc
	mu          sync.Mutex
	cfgMgr      *Manager
	client      *banye.Client
	publisher   *pub.Publisher
}

func NewWorkerManager(cfgMgr *Manager, client *banye.Client, publisher *pub.Publisher) *WorkerManager {
	return &WorkerManager{
		cancelFuncs: make(map[string]context.CancelFunc, 20),
		cfgMgr:      cfgMgr,
		client:      client,
		publisher:   publisher,
	}
}

func (wm *WorkerManager) Reconcile(oldCfg, newCfg *types.Config) {
	log.Println("Started reconciling")
	log.Printf("wm.publisher.Ch.IsClosed(): %v\n", wm.publisher.Ch.IsClosed())

	wm.mu.Lock()
	defer wm.mu.Unlock()

	newEndpoints := make(map[string]types.Endpoint)
	newQueues := make(map[string]types.Queue)
	newExchanges := make(map[string]types.Exchange)
	newBindings := make(map[string]types.Binding)

	for _, ep := range newCfg.Service.Endpoints {
		err := ep.ParsePathVariables()
		if err != nil {
			log.Printf("Couldn't parse path %v", ep.Path)
		}
		key := ep.ParsedPath + "|" + ep.Method
		newEndpoints[key] = ep
	}
	for _, q := range newCfg.RabbitMQ.Queues {
		newQueues[q.Name] = q
	}
	for _, e := range newCfg.RabbitMQ.Exchanges {
		newExchanges[e.Name] = e
	}
	for _, b := range newCfg.RabbitMQ.Bindings {
		newBindings[b.Exchange+"|"+b.QueueName+"|"+b.RoutingKey] = b
	}

	oldQueues := make(map[string]types.Queue)
	oldExchanges := make(map[string]types.Exchange)
	oldBindings := make(map[string]types.Binding)
	oldEndpoints := make(map[string]types.Endpoint)

	if oldCfg != nil {
		for _, ep := range oldCfg.Service.Endpoints {
			key := ep.ParsedPath + "|" + ep.Method
			oldEndpoints[key] = ep
		}
		for _, q := range newCfg.RabbitMQ.Queues {
			oldQueues[q.Name] = q
		}
		for _, e := range newCfg.RabbitMQ.Exchanges {
			oldExchanges[e.Name] = e
		}
		for _, b := range newCfg.RabbitMQ.Bindings {
			oldBindings[b.Exchange+"|"+b.QueueName+"|"+b.RoutingKey] = b
		}
	}

	for _, self := range oldCfg.RabbitMQ.Queues {
		_, existsInNew := newQueues[self.Name]
		newQ := self.Canonical()

		if !existsInNew {
			err := queues.DeleteQueue(wm.publisher.Ch, newQ)
			if err != nil {
				log.Println(err)
				continue
			}
			delete(newQueues, self.Name)
			continue
		}

		oldQ := oldQueues[self.Name]

		if !self.Enabled {
			err := queues.DeleteQueue(wm.publisher.Ch, newQ)
			if err != nil {
				log.Println(err)
				continue
			}
			delete(newQueues, self.Name)
			continue
		}
		if oldQ.Name == self.Name && !types.Equal(oldQ, self) {
			err := queues.DeleteQueue(wm.publisher.Ch, newQ)
			if err != nil {
				log.Println(err)
				continue
			}
			delete(newQueues, self.Name)
			continue
		}
	}
	for _, self := range oldCfg.RabbitMQ.Bindings {
		name := self.Exchange + "|" + self.QueueName + "|" + self.RoutingKey
		_, existsInNew := newBindings[name]
		newB := self.Canonical(newQueues[self.QueueName].Canonical(), newExchanges[self.Exchange].Canonical())

		if !existsInNew {
			wm.publisher.Unbind(&newB)
			delete(newBindings, name)
			continue
		}

		if !self.Enabled {
			wm.publisher.Unbind(&newB)
			delete(newBindings, name)
			continue
		}
	}
	for key, cancel := range wm.cancelFuncs {
		_, existsInNew := newEndpoints[key]
		if !existsInNew {
			cancel()
			delete(wm.cancelFuncs, key)
			log.Printf("Stopped endpoint %s (rem)", key)
			continue
		}

		newEp := newEndpoints[key]
		oldEp := oldEndpoints[key]

		if !newEp.Enabled {
			cancel()
			delete(wm.cancelFuncs, key)
			log.Printf("Stopped endpoint %s (dis)", key)
			continue
		}
		if oldEp.Enabled && (oldEp.ParsedPath != newEp.ParsedPath ||
			oldEp.Method != newEp.Method ||
			oldEp.Timeout != newEp.Timeout) {
			cancel()
			delete(wm.cancelFuncs, key)
			log.Printf("Stopped endpoint %s (crit cfg change)", key)
		}
	}

	log.Printf("started creating wm.publisher.Ch.IsClosed(): %v\n", wm.publisher.Ch.IsClosed())

	r, err := wm.client.MakeRequest(context.TODO(), "GET", fmt.Sprintf("http://%v:15672/api/queues/%v", os.Getenv("RABBIT_HOST"), "%2F"),
		map[string]string{"username": "guest", "password": "guest"})
	if err != nil {
		log.Printf("bad news with queues: %v", err)
	}
	var runQ []types.Queue
	_ = json.Unmarshal(r.Body, &runQ)
	for key, q := range newQueues {
		if !q.Enabled {
			continue
		}

		running := false
		for _, queue := range runQ {
			if queue.Name == key {
				running = true
				break
			}
		}
		if !running {
			target := "localhost:9130"
			log.Printf("Dialing %v", target)
			conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("Failed to make a grpc client!: %v", err)
				continue
			}
			defer conn.Close()

			client := pb.NewQueueServiceClient(conn)

			args, err := mapconverter.ConvertMap(q.Args)
			if err != nil {
				log.Printf("Unable to serialize a map: %v", err)
			}

			resp, err := client.CreateQueue(context.Background(), &pb.Queue{
				Name:         q.Name,
				Durable:      &q.Durable,
				DeleteUnused: &q.DeleteOnUnused,
				Exclusive:    &q.Exclusive,
				NoWait:       &q.NoWait,
				Args:         args,
			})
			if err != nil {
				log.Printf("Error: %v", err)
			}
			log.Println(*resp.Error)
		}
	}

	r, err = wm.client.MakeRequest(context.TODO(), "GET", fmt.Sprintf("http://%v:15672/api/exchanges/%v", os.Getenv("RABBIT_HOST"), "%2F"),
		map[string]string{"username": "guest", "password": "guest"})
	if err != nil {
		log.Printf("Problem in exchange: %v", err)
	}
	var runE []types.Exchange
	_ = json.Unmarshal(r.Body, &runE)
	for key, e := range newExchanges {
		if !e.Enabled {
			continue
		}

		running := false
		for _, exchange := range runE {
			if exchange.Name == key {
				running = true
				break
			}
		}
		if !running {
			conn, err := grpc.NewClient("localhost:9130", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("Failed to make a grpc client!: %v", err)
				continue
			}
			defer conn.Close()

			client := pb.NewExchangeServiceClient(conn)

			args, err := mapconverter.ConvertMap(e.Args)
			if err != nil {
				log.Printf("Unable to serialize a map: %v", err)
			}

			resp, err := client.CreateExchange(context.Background(), &pb.Exchange{
				Name:        e.Name,
				Type:        e.Type,
				Durable:     &e.Durable,
				AutoDeleted: &e.AutoDeleted,
				Internal:    &e.AutoDeleted,
				NoWait:      &e.NoWait,
				Args:        args,
			})
			if err != nil {
				log.Printf("Error: %v", err)
			}
			log.Println(resp.Error)
		}
	}

	r, _ = wm.client.MakeRequest(context.TODO(), "GET", fmt.Sprintf("http://%v:15672/api/bindings/%v", os.Getenv("RABBIT_HOST"), "%2F"),
		map[string]string{"username": "guest", "password": "guest"})
	var runB []types.Binding
	_ = json.Unmarshal(r.Body, &runB)
	for key, b := range newBindings {
		if !b.Enabled {
			continue
		}

		running := false
		for _, binding := range runB {
			if binding.Exchange+"|"+binding.QueueName+"|"+binding.RoutingKey == key {
				running = true
				break
			}
		}
		if !running {
			target := "localhost:9130"
			log.Printf("Dialing %v", target)
			conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("Failed to make a grpc client!: %v", err)
				continue
			}
			defer conn.Close()

			client := pb.NewBindingServiceClient(conn)
			resp, err := client.Bind(context.Background(), &pb.Binding{
				Queue:      b.QueueName,
				Exchange:   b.Exchange,
				RoutingKey: b.RoutingKey,
			})
			if err != nil {
				log.Printf("Error: %v", err)
			}
			log.Println(resp.Error)
		}
	}

	log.Printf("created all queues etc. wm.publisher.Ch.IsClosed(): %v\n", wm.publisher.Ch.IsClosed())

	for key, ep := range newEndpoints {
		if !ep.Enabled {
			continue
		}
		if _, running := wm.cancelFuncs[key]; !running {
			c, cancel := context.WithCancel(context.Background())
			wm.cancelFuncs[key] = cancel

			if ep.BenchmarkPath != "" {
				benchmarkResp, err := wm.client.MakeRequest(context.TODO(), ep.Method, ep.BenchmarkPath, nil)
				if err != nil {
					log.Printf("Error in benchmark request: %v", err)
				}
				ep.BenchmarkResponseHash = benchmarkResp.Hash()
				log.Printf("BMRH: %v", ep.BenchmarkResponseHash)
			}

			go wm.runEndpoint(c, ep)
			log.Printf("Started endpoint %s %s", ep.Method, ep.ParsedPath)
		}
	}
}

func (wm *WorkerManager) runEndpoint(c context.Context, ep types.Endpoint) {
	interval, err := time.ParseDuration(ep.Timeout)
	if err != nil {
		log.Printf("Endpoint %s: bad timeout %s, using 30m default",
			ep.ParsedPath, ep.Timeout)
		interval = 30 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Done():
			log.Printf("Endpoint %s is shutting down", ep.ParsedPath)
			return
		case <-ticker.C:
			log.Println("making a request to ", ep.ParsedPath)

			resp, err := wm.client.MakeRequest(context.TODO(), ep.Method, ep.ParsedPath, nil)
			if err != nil {
				log.Printf("Error in request %v %v : %v",
					ep.Method, ep.ParsedPath, err)
			} //else if resp != nil {
			// 	log.Println(*resp)
			// }

			respHash := resp.Hash()
			log.Printf("respHash: %v\nbenchHash: %v", respHash, ep.BenchmarkResponseHash)

			if respHash != ep.BenchmarkResponseHash {

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				pack := pub.PublishingPackage{
					Exchange:  ep.Exchange,
					Key:       ep.RoutingKey,
					Mandatory: false,
					Immediate: false,
					Message: amqp.Publishing{
						ContentType: "text/plain",
						Body:        resp.Body,
					},
				}
				err = wm.publisher.Publish(ctx, pack)
				if err != nil {
					log.Panicf("Couldn't publish: %v", err)
				}
				log.Printf("sent %#v", pack)

				if len(ep.Params) > 0 &&
					resp.StatusCode >= 200 && resp.StatusCode < 300 {

					value := int(ep.Params[ep.ParsedVariables()[0]].(int64))
					currConf := *wm.cfgMgr.Get()

					err = types.SwitchParams(&currConf,
						ep.Method+"|"+ep.ParsedPath,
						ep.ParsedVariables()[0],
						value+1)

					f, _ := os.Create("config.toml")
					toml.NewEncoder(f).Encode(currConf)
					defer f.Close()
				}
			}
		}
	}
}
