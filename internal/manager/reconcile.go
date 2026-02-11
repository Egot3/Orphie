package manager

import (
	"context"
	"log"
	"newsgetter/internal/types"
	"newsgetter/internal/utils"
	"sync"
	"time"
)

type WorkerManager struct {
	cancelFuncs map[string]context.CancelFunc
	mu          sync.Mutex
}

func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		cancelFuncs: make(map[string]context.CancelFunc),
	}
}

func (wm *WorkerManager) Reconcile(oldCfg, newCfg *types.ServiceStruct) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	log.Println("amount of endpoints: ", len(newCfg.Endpoints))

	newEndpoints := make(map[string]types.Endpoint)
	for _, ep := range newCfg.Endpoints {
		key := ep.Path + "|" + ep.Method
		newEndpoints[key] = ep
	}

	oldEndpoints := make(map[string]types.Endpoint)
	if oldCfg != nil {
		for _, ep := range oldCfg.Endpoints {
			key := ep.Path + "|" + ep.Method
			oldEndpoints[key] = ep
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
		if oldEp.Enabled && (oldEp.Path != newEp.Path ||
			oldEp.Method != newEp.Method ||
			oldEp.Timeout != newEp.Timeout) {
			cancel()
			delete(wm.cancelFuncs, key)
			log.Printf("Stopped endpoint %s (crit cfg change)", key)
		}
	}

	for key, ep := range newEndpoints {
		if !ep.Enabled {
			continue
		}
		if _, running := wm.cancelFuncs[key]; !running {
			c, cancel := context.WithCancel(context.Background())
			wm.cancelFuncs[key] = cancel
			go wm.runEndpoint(c, ep)
			log.Printf("Started endpoint %s %s", ep.Method, ep.Path)
		}
	}
}

func (wm *WorkerManager) runEndpoint(c context.Context, ep types.Endpoint) {
	interval, err := time.ParseDuration(ep.Timeout)
	if err != nil {
		log.Printf("Endpoint %s: bad timeout %s, using 30m default", ep.Path, ep.Timeout)
		interval = 30 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Done():
			log.Printf("Endpoint %s is shutting down", ep.Path)
			return
		case <-ticker.C:
			log.Println("making a request to ", ep.Path)
			resp, err := utils.MakeRequest(ep.Method, ep.Path)
			if err != nil {
				log.Printf("Error in request %v %v : %v", ep.Method, ep.Path, err)
			} else if resp != nil {
				log.Println(*resp)
			}
		}
	}
}
