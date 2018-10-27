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


