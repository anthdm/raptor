package storage

import (
	"sync"

	"github.com/bananabytelabs/wazero"
	"github.com/google/uuid"
)

type ModCacher interface {
	Put(uuid.UUID, wazero.CompilationCache)
	Get(uuid.UUID) (wazero.CompilationCache, bool)
	Delete(uuid.UUID) error
}

type DefaultModCache struct {
	mu    sync.RWMutex
	cache map[uuid.UUID]wazero.CompilationCache
}

func NewDefaultModCache() *DefaultModCache {
	return &DefaultModCache{
		cache: make(map[uuid.UUID]wazero.CompilationCache, 0),
	}
}

func (c *DefaultModCache) Put(id uuid.UUID, mod wazero.CompilationCache) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[id] = mod
}

func (c *DefaultModCache) Get(id uuid.UUID) (wazero.CompilationCache, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	mod, ok := c.cache[id]
	return mod, ok
}

func (c *DefaultModCache) Delete(id uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, id)
	return nil
}
