package queue

import "sync"

type ExistsMap struct {
	sync.RWMutex
	internal map[interface{}]bool
}

func NewExistsMap() *ExistsMap {
	return &ExistsMap{
		internal: make(map[interface{}]bool),
	}
}

func (em *ExistsMap) Exists(key interface{}) bool {
	em.RLock()
	_, ok := em.internal[key]
	em.RUnlock()
	return ok
}

func (em *ExistsMap) Delete(key interface{}) {
	em.Lock()
	delete(em.internal, key)
	em.Unlock()
}

func (em *ExistsMap) Put(key interface{}) {
	em.Lock()
	em.internal[key] = true
	em.Unlock()
}
