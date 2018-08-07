package queue

/*
 *  The Transaction Queue is a wrapper of the Priority queue heap implementation
 *  It uses the thread safe queue functions for pushing and popping from the queue
 *
 *  The intent is to make it very simple to use and eliminate casting
 *  It includes a few helper functions to keep thing easy
 */
import (
	"container/heap"
	"github.com/dispatchlabs/disgo/commons/types"
	"time"
	"github.com/dispatchlabs/disgo/commons/utils"
	"sort"
)

type TransactionQueue struct {
	Queue 		*PriorityQueue
	ExistsMap 	*ExistsMap
	LockTime    int64
}

// need to figure out best way to implement the lock time
func NewTransactionQueue(lockTimeInSeconds int) *TransactionQueue {
	txq := make(PriorityQueue, 0)
	heap.Init(&txq)
	WatchHeapOps()
	return &TransactionQueue{&txq, NewExistsMap(), int64(lockTimeInSeconds * 1000)}
}

// - Push onto the queue and then resort (latest to earliest) also add to fast Exists map for quick checks
func (txq *TransactionQueue) Push(tx *types.Transaction) {
	itm := Item{tx, tx.Time, txq.Queue.Len()+1}
	HeapPush(txq.Queue, &itm)
	if(txq.Queue.Len() > 0) {
		sort.Sort(txq.Queue)
	}
	txq.ExistsMap.Put(tx.Hash)
}

// - Push onto the queue and then resort (latest to earliest) also add to fast Exists map for quick checks
func (txq *TransactionQueue) Pop() *types.Transaction {
	itm := HeapPop(txq.Queue).(*Item)
	tx := itm.Data.(*types.Transaction)
	txq.ExistsMap.Delete(tx.Hash)
	return tx
}

// - Check to see if there is an item in the queue that is older than the LockTime
func (txq TransactionQueue) HasAvailable() bool {
	timestamp := txq.Queue.Peek()
	if timestamp != -1 && utils.ToMilliSeconds(time.Now()) - timestamp > txq.LockTime {
		return true
	}
	return false
}

// - Check to see if the current transaction Hash is in the exists map
func (txq TransactionQueue) Exists(key string) bool {
	return txq.ExistsMap.Exists(key)
}

// - Dump returns the contents of the queue from oldest to newest (the order they are constantly sorted to)
func (txq TransactionQueue) Dump() []*types.Transaction {

	txList := make([]*types.Transaction, 0)
	for _, itm := range *txq.Queue {
		txList = append(txList, itm.Data.(*types.Transaction))
	}
	return txList
}