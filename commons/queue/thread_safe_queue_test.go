package queue


import (
	"github.com/dispatchlabs/disgo/commons/types"
	"time"
	"fmt"
	"sync"
	"testing"
	"container/heap"
	"github.com/dispatchlabs/disgo/commons/utils"
	"math/rand"
)

func CreateMockTransactions(pq *PriorityQueue, count int) {
	utils.Info(count)
	rand.Seed(100)
	txs := make([]*types.Transaction, 0)
	for i := 1; i <= count; i++ {
		tx := GetMockTransaction(int64(i))
		tx.Time = int64(rand.Intn(100))
		p := Item{tx, tx.Time, i}
		fmt.Printf("\n *** Adding tx %v\n", tx.Hash)
		HeapPush(pq, &p)
		txs = append(txs, tx)
		time.Sleep(time.Second)
	}
}

func GetMockTransaction(value int64) *types.Transaction {
	tx, err := types.NewTransferTokensTransaction(
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

func startListening(pq *PriorityQueue) {
	utils.Info("startListening")
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			if(pq.Len() > 0) {
				fmt.Printf("\n***** There is a record in the queue *****\n")
				itm := HeapPop(pq).(*Item)
				var tx *types.Transaction
				tx = itm.Data.(*types.Transaction)
				fmt.Printf(tx.String())
			}
		}
	}()
	wg.Wait()
}

func TestQueue(t *testing.T) {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	WatchHeapOps()

	CreateMockTransactions(&pq,5)
	items := pq.Dump()
	for _, itm := range items {
		var tx *types.Transaction
		tx = itm.Data.(*types.Transaction)
		fmt.Printf("Time: %v\n", tx.Time)
	}
	fmt.Printf("\nQueue length = %v\n", len(items))
	go startListening(&pq)
}
