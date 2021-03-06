package queue

import (
	"fmt"
	"sync"

	"github.com/alash3al/exeq/internals/config"
)

type Driver interface {
	Open(*config.QueueConfig) (Driver, error)
	Enqueue(*Job) error
	List() ([]Job, error)
	Stats() (Stats, error)
	Err() <-chan error
	ListenAndConsume() error
}

type Stats struct {
	Pending   int64
	Running   int64
	Succeeded int64
	Failed    int64
	Retries   int64
}

var (
	drivers      = map[string]Driver{}
	driversMutex = &sync.RWMutex{}
)

func Register(name string, d Driver) error {
	driversMutex.Lock()
	defer driversMutex.Unlock()

	if _, found := drivers[name]; found {
		return fmt.Errorf("duplicate driver name %s", name)
	}

	drivers[name] = d

	return nil
}

func Open(name string, cfg *config.QueueConfig) (Driver, error) {
	driversMutex.RLock()
	defer driversMutex.RUnlock()

	driver, found := drivers[name]
	if !found {
		return nil, fmt.Errorf("driver %s not found", name)
	}

	return driver.Open(cfg)
}
