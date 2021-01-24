/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
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

type CtxWithCancel struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}
