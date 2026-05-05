package config

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type Subscriber func(cfg Config)

type Manager struct {
	path        string
	configValue atomic.Value

	mu          sync.RWMutex
	subscribers []Subscriber
}

func NewManager(path string) (*Manager, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	manager := &Manager{path: path}
	manager.configValue.Store(cfg)
	return manager, nil
}

func (m *Manager) Current() Config {
	return m.configValue.Load().(Config)
}

func (m *Manager) Subscribe(fn Subscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers = append(m.subscribers, fn)
}

func (m *Manager) Start(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer func() {
		err := watcher.Close()
		if err != nil {
			fmt.Println("failed to close watcher:", err)
		}
	}()

	dir := filepath.Dir(m.path)
	filename := filepath.Base(m.path)
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("watch config dir: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-watcher.Errors:
			if err != nil {
				return fmt.Errorf("watch config: %w", err)
			}
		case event := <-watcher.Events:
			if filepath.Base(event.Name) != filename {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			if err := m.reload(); err != nil {
				continue
			}
		}
	}
}

func (m *Manager) reload() error {
	cfg, err := Load(m.path)
	if err != nil {
		return err
	}

	m.configValue.Store(cfg)

	m.mu.RLock()
	subscribers := append([]Subscriber(nil), m.subscribers...)
	m.mu.RUnlock()

	for _, subscriber := range subscribers {
		subscriber(cfg)
	}

	return nil
}
