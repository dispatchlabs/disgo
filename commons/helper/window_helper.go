package helper

import (
	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/patrickmn/go-cache"

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
		CalcSlopeForWindow(cache, window)
		//CalcRollingAverageHertzForWindow(cache, window)

	} else {
		window = val.(*types.Window)
		window.AddHertz(cache, hertz)
	}
	return window
}

func persistPreviousWindow(txn *badger.Txn, cache *cache.Cache, id int64) {
	window, ok := types.ToWindowFromCache(cache, id)
	if !ok {
		return
	}
	window.Persist(txn)
}


func CalcSlopeForWindow(cache *cache.Cache, window *types.Window) {
	points := make([]utils.Point, 0)

	found :=0
	for i := window.Id - types.AvgWindowSize - 1; i < window.Id; i++ {
		win, ok := types.ToWindowFromCache(cache, i)
		if !ok {
			continue
		}
		found++
		points = append(points, utils.Point{X: float64(found), Y: float64(win.Sum),})

		utils.Info("XXXX Found", found)
		utils.Info("XXXX sum", win.Sum)

		//sum += win.Sum

		utils.Debug("Calc for minute: ", i )
	}
	if(found > 0) {
		window.Slope, _ = utils.LinearRegression(&points)
	} else {
		window.Slope = 0
	}
}

func populateCache(txn *badger.Txn, cache *cache.Cache) {
	currentWindow := types.NewWindow()
	for i := currentWindow.Id; i > (currentWindow.Id - types.AvgWindowSize); i-- {
		window, err := types.ToWindowFromKey(txn, i)
		if err != nil {
			utils.Debug("ID: ", i, err)
			continue
		}
		if window.Sum > 0 {
			utils.Debug("Add to cache --> ", window.String())
			window.Cache(cache)
		}
	}
}