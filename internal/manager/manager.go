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
	onReload func(old, new *types.ServiceStruct)
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
		onReload: onReload,
	}

	m.config.Store(&types.ServiceStruct{})

	_ = m.load()

	if err := m.load(); err != nil {
		return nil, err
	}
	if err := watcher.Add(path); err != nil {
		return nil, err
	}

	go m.watch()
	return m, nil
}

func (m *Manager) load() error {
	var cfg types.Config
	if _, err := toml.DecodeFile(m.path, &cfg); err != nil {
		return err
	}

	oldCfg := m.Get()
	m.config.Store(&cfg.Service)

	if m.onReload != nil {
		m.onReload(oldCfg, &cfg.Service)
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
				if err := m.load(); err != nil {
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

func (m *Manager) Get() *types.ServiceStruct {
	serv, ok := m.config.Load().(*types.ServiceStruct)
	if !ok {
		log.Println("couldn't load a config")
		return nil
	}
	return serv
}

func (m *Manager) Stop() error {
	close(m.done)
	return m.watcher.Close()
}
