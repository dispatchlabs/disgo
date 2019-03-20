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
package services

import (
	"github.com/dgraph-io/badger"
	badgerOptions "github.com/dgraph-io/badger/options"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/patrickmn/go-cache"
	"os"
	"sync"
)

var dbServiceInstance *DbService
var dbServiceOnce sync.Once

// GetDbService
func GetDbService() *DbService {
	dbServiceOnce.Do(func() {
		dbServiceInstance = &DbService{running: false, kmutex: utils.NewKmutex(), cache: cache.New(types.CacheTTL, types.CacheTTL*2)}
		dbServiceInstance.openDb()
	})
	return dbServiceInstance
}

// DbService
type DbService struct {
	running bool
	db      *badger.DB
	kmutex  *utils.Kmutex
	cache   *cache.Cache
}

// IsRunning
func (this *DbService) IsRunning() bool {
	return this.running
}

// Close
func (this *DbService) Close() {
	this.db.Close()
}

// Go
func (this *DbService) Go() {
	this.running = true
	utils.Events().Raise(types.Events.DbServiceInitFinished)


	//on boot up, start running garbage collection every minute
	//ticker := time.NewTicker(5 * time.Minute)
	//defer ticker.Stop()
	//for range ticker.C {
	//again:
	//	utils.Info("looping garbage collection")
	//	err := this.db.RunValueLogGC(0.1)
	//	if err == nil {
	//		goto again
	//	}
	//}
}

// openDb
func (this *DbService) openDb() {
	fileName := "." + string(os.PathSeparator) + "db" + string(os.PathSeparator) + "LOCK"
	if utils.Exists(fileName) {
		err := os.Remove(fileName)
		if err != nil {
			utils.Fatal(err)
		}
	}

	utils.Info("opening DB...")
	opts := badger.DefaultOptions
	opts.Dir = "." + string(os.PathSeparator) + "db"
	opts.ValueDir = "." + string(os.PathSeparator) + "db"
	opts.ValueLogLoadingMode = badgerOptions.FileIO // https://github.com/dgraph-io/badger/issues/246
	db, err := badger.Open(opts)
	if err != nil {
		utils.Fatal(err)
	}
	this.db = db

	utils.Info("starting garbage collection")
	for i := 0; i < 1009; i++ {
	again:
		utils.Info("looping garbage collection")
		err = this.db.RunValueLogGC(0.01)
		if err == nil {
			goto again
		} else {
			utils.Error(err)
		}
}


}

// GetCache
func GetCache() *cache.Cache {
	return GetDbService().cache
}

// GetDb
func GetDb() *badger.DB {
	return GetDbService().db
}

// NewTxn
func NewTxn(update bool) *badger.Txn {
	return GetDbService().db.NewTransaction(update)
}

// Lock
func Lock(key interface{}) {
	GetDbService().kmutex.Lock(key)
}

// Unlock
func Unlock(key interface{}) {
	GetDbService().kmutex.Unlock(key)
}
