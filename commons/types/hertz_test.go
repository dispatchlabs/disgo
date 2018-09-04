package types

import (
	"testing"
)

//var c *cache.Cache
//var db *badger.DB
//var dbPath = "." + string(os.PathSeparator) + "testdb"


////init
//func init()  {
//	c = cache.New(CacheTTL, CacheTTL*2)
//	utils.Info("opening DB...")
//	opts := badger.DefaultOptions
//	opts.Dir = dbPath
//	opts.ValueDir = dbPath
//	db, _ = badger.Open(opts)
//}

func TestHerz(t *testing.T) {
	hertz := Hertz{Db: db}
	hertz.Merge()
}