package controller

import (
	"sync"
)

// contains the key of the currently processing finalizer
// it's concurrency safe
type mapFinalizer struct {
	keys map[string]bool
	lock *sync.Mutex
}

func NewMapFinalizer() *mapFinalizer {
	return &mapFinalizer{
		keys: make(map[string]bool),
		lock: &sync.Mutex{},
	}
}

func (f *mapFinalizer) IsAlreadyProcessing(key string) bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	_, ok := f.keys[key]
	return ok
}

func (f *mapFinalizer) Add(key string) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.keys[key] = true
}

func (f *mapFinalizer) Delete(key string) {
	f.lock.Lock()
	defer f.lock.Unlock()
	delete(f.keys, key)
}
