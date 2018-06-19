/*
 *    This file is part of Disgo-Commons library.
 *
 *    The Disgo-Commons library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo-Commons library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo-Commons library.  If not, see <http://www.gnu.org/licenses/>.
 */
package utils

import "sync"

// Kmutex
type Kmutex struct {
	cond   *sync.Cond
	locker sync.Locker
	keys   map[interface{}]struct{}
}

// NewKmutex
func NewKmutex() *Kmutex {
	locker := sync.Mutex{}
	return &Kmutex{cond: sync.NewCond(&locker), locker: &locker, keys: make(map[interface{}]struct{})}
}

// Lock
func (this *Kmutex) Lock(key interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()
	for this.locked(key) {
		this.cond.Wait()
	}
	this.keys[key] = struct{}{}
	return
}

// Unlock
func (this *Kmutex) Unlock(key interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()
	delete(this.keys, key)
	this.cond.Broadcast()
}

// locked
func (this *Kmutex) locked(key interface{}) (ok bool) { _, ok = this.keys[key]; return }
