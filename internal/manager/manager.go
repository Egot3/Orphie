package manager

import (
	"log"
	"newsgetter/internal/types"
	"sync/atomic"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
)

type Manager struct {
	config   atomic.Value
	path     string
	watcher  *fsnotify.Watcher
	done     chan struct{}
	OnReload func(old, new *types.ServiceStruct)
}

func NewManager(path string, onReload func(old, new *types.ServiceStruct)) (*Manager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	m := &Manager{
		path:     path,
		watcher:  watcher,
		done:     make(chan struct{}),
		OnReload: onReload,
	}

	m.config.Store(&types.Config{})

	if err := m.Load(); err != nil {
		return nil, err
	}
	if err := watcher.Add(path); err != nil {
		return nil, err
	}

	go m.watch()
	return m, nil
}

func (m *Manager) Load() error {
	var cfg types.Config
	_, err := toml.DecodeFile(m.path, &cfg)
	if err != nil {
		return err
	}

	// var rawCfg map[string]interface{}
	// _, err = toml.DecodeFile(m.path, &rawCfg)
	// if err != nil {
	// 	return err
	// }

	// service := rawCfg["service"].(map[string]interface{})
	// endpoints := service["endpoints"].([]map[string]interface{})

	// for i := range cfg.Service.Endpoints {
	// 	cfg.Service.Endpoints[i].Params = make(map[string]interface{})
	// 	for key, val := range endpoints[i] {
	// 		if key != "path" && key != "method" && key != "timeout" && key != "enabled" && key != "updateLogPath" && key != "lastUpdate" {
	// 			cfg.Service.Endpoints[i].Params[key] = val
	// 		}
	// 	}
	// } <- mental illness

	oldCfg := m.Get()
	m.config.Store(&cfg)

	if m.OnReload != nil {
		m.OnReload(&oldCfg.Service, &cfg.Service)
	}
	return nil
}

func (m *Manager) watch() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Chmod) != 0 {
				if err := m.Load(); err != nil {
					log.Printf("failed to reload config: %v", err)
				} else {
					log.Println("Config reloaded!")
				}
			}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher's error: %v", err)
		case <-m.done:
			return
		}
	}
}

func (m Manager) Get() *types.Config {
	cfg, ok := m.config.Load().(*types.Config)
	if !ok {
		log.Println("couldn't Load a config")
		return nil
	}
	return cfg
}

func (m *Manager) Stop() error {
	close(m.done)
	return m.watcher.Close()
}
