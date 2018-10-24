package helper

import (
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dgraph-io/badger"
	"github.com/patrickmn/go-cache"
	"github.com/dispatchlabs/disgo/commons/utils"
	"time"
)

var cacheLoaded bool

func AddHertz(txn *badger.Txn, cache *cache.Cache, hertz uint64) *types.Window {
	epoch := time.Unix(0, types.DispatchEpoch)
	minutesSinceEpoch := int64(time.Now().Sub(epoch).Minutes())

	if !cacheLoaded {
		populateCache(txn, cache)
		cacheLoaded = true
	}
	var window *types.Window
	val, ok := cache.Get(types.GetWindowKey(minutesSinceEpoch))
	if !ok {
		window = types.NewWindow()
		persistPreviousWindow(txn, cache, minutesSinceEpoch-1)

		window.AddHertz(cache, hertz)
		CalcRollingAverageHertzForWindow(cache, window)

	} else {
		window = val.(*types.Window)
		window.AddHertz(cache, hertz)
	}
	window.Persist(txn)
	return window
}

func persistPreviousWindow(txn *badger.Txn, cache *cache.Cache, id int64) {
	window, ok := types.ToWindowFromCache(cache, id)
	if !ok {
		return
	}
	window.Persist(txn)
}

func CalcRollingAverageHertzForWindow(cache *cache.Cache, window *types.Window) {

	sum := window.Sum
	found :=0
	for i := window.Id-1; i > (window.Id - types.AvgWindowSize); i-- {
		win, ok := types.ToWindowFromCache(cache, i)
		if !ok {
			continue
		}
		found++
		sum += win.Sum
		utils.Debug("Calc for minute: ", i )
	}
	if(found > 0) {
		window.RollingAverage = uint64(sum / uint64(types.AvgWindowSize))
	} else {
		window.RollingAverage = uint64(0)
	}
}

func populateCache(txn *badger.Txn, cache *cache.Cache) {
	utils.Info("populateCache for rate limiting")
	currentWindow := types.NewWindow()
	for i := currentWindow.Id; i > (currentWindow.Id - types.AvgWindowSize); i-- {
		window, err := types.ToWindowFromKey(txn, i)
		if err != nil {
			utils.Debug("ID: ", i, err)
			continue
		}
		if window.Sum > 0 {
			utils.Info("Add to cache --> ", window.String())
			window.Cache(cache)
		}
	}
}