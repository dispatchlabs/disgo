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
	"fmt"
	"github.com/dgraph-io/badger"
	badgerOptions "github.com/dgraph-io/badger/options"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/patrickmn/go-cache"
	"github.com/robfig/cron"
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

	//set up cron routine to collect garbage in badgerdb
	CollectGarbage()

	c := cron.New()
	c.AddFunc("@every 1h", func() {CollectGarbage()})
	c.Start()
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

func CollectGarbage() {
	utils.Info("starting garbage collection")
	var cleaned = 0

	for i := 0; i < 100; i++ {
	again:
		fmt.Printf("\r%d log files cleaned", cleaned)
		err := GetDbService().db.RunValueLogGC(0.01)
		if err == nil {
			cleaned++
			fmt.Printf("%d log files cleaned", cleaned)
			goto again
		} else if err.Error() != "Value log GC attempt didn't result in any cleanup"{
			utils.Error(err)
		}
	}
}
