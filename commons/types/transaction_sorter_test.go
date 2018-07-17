package types

import (
	"time"
	"github.com/dispatchlabs/disgo/commons/utils"
	"testing"
)

func getMockTransaction(value int64) *Transaction {
	tx, err := NewTransferTokensTransaction(
		"0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a",
		"3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c",
		"d70613f93152c84050e7826c4e2b0cc02c1c3b99",
		value,
		0,
		utils.ToMilliSeconds(time.Now()),
	)
	if err != nil {
		panic(err)
	}
	return tx
}

func TestSortingTransactions(t *testing.T) {
	txs := make([]*Transaction, 0)
	var i int64
	for i = 1; i <= 5; i++ {
		tx := getMockTransaction(i)
		txs = append(txs, tx)
		time.Sleep(time.Second)
	}
	SortByTime(txs, false)
	var lastTime int64
	lastTime = 0;
	for _, tx := range txs {
		if lastTime == 0 {
			lastTime = tx.Time
		} else {
			if tx.Time > lastTime {
				t.Fatalf("transactions are not sorted")
			}
		}
		//t.Logf("transaction %d has value %d and last value was %d", i, tx.Time, lastTime)
		lastTime = tx.Time
	}
}