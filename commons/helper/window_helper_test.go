package helper

import (
	"testing"
	"github.com/dgraph-io/badger"
	"github.com/patrickmn/go-cache"

	"os"
	"github.com/dispatchlabs/disgo/commons/utils"
	"time"
	"github.com/dispatchlabs/disgo/commons/types"
	"fmt"
)

var c *cache.Cache
var db *badger.DB
var dbPath = "." + string(os.PathSeparator) + "testdb"

//init
func init()  {
	c = cache.New(types.CacheTTL, types.CacheTTL*2)
	utils.Info("opening DB...")
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, _ = badger.Open(opts)
}


func TestWindow(t *testing.T) {
	txn := db.NewTransaction(true)
	defer txn.Discard()
	window := AddHertz(txn, c, uint64(utils.Random(0, 1000)))
	fmt.Printf("%s\n", window.ToPrettyJson())
	time.Sleep(2 * time.Second)
	txn.Commit(nil)
}

func TestCalcSlopeForWindow(t *testing.T) {
	for i := 241 - types.AvgWindowSize - 1; i < 241; i++ {
		window := types.NewWindow()
		window.Id = int64(i)
		window.AddHertz(c, 1000)
	}

	// test for a slope of zero
	window := types.NewWindow()
	window.Id = 242
	CalcSlopeForWindow(c, window)
	if window.Slope != 0 {
		t.Errorf("Window slope is not zero when it should be, instead it is %f", window.Slope)
	}

	for i := 441 - types.AvgWindowSize - 1; i < 441; i++ {
		window := types.NewWindow()
		window.Id =	int64(i)
		window.AddHertz(c, uint64(i))
	}

	// test for a slope of one
	window = types.NewWindow()
	window.Id = 442
	CalcSlopeForWindow(c, window)
	if window.Slope != 1 {
		t.Errorf("Window slope is not one when it should be, instead it is %f", window.Slope)
	}

	for i := 641 - types.AvgWindowSize - 1; i < 641; i++ {
		window := types.NewWindow()
		window.Id =	int64(i)
		window.AddHertz(c, uint64(641 - i))
	}

	// test for a slope of negative one
	window = types.NewWindow()
	window.Id = 642
	CalcSlopeForWindow(c, window)
	if window.Slope != -1 {
		t.Errorf("Window slope is not one when it should be, instead it is %f", window.Slope)
	}

	for i := 841 - types.AvgWindowSize - 1; i < 841; i++ {
		window := types.NewWindow()
		window.Id =	int64(i)
		window.AddHertz(c, uint64(i * 86400))
	}

	// Test for a 24 hour slope
	window = types.NewWindow()
	window.Id = 842
	CalcSlopeForWindow(c, window)
	if window.Slope != 86400 {
		t.Errorf("Window slope is not 86400 when it should be, instead it is %f", window.Slope)
	}
}
