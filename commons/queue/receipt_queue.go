package queue

/*
 *  The Receipt Queue is a wrapper of the Priority queue heap implementation
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

type ReceiptQueue struct {
	Queue 		*PriorityQueue
	ExistsMap 	map[string]bool
	LockTime    int64
}

// need to figure out best way to implement the lock time
func NewReceiptQueue(lockTimeInSeconds int) *ReceiptQueue {
	rtq := make(PriorityQueue, 0)
	heap.Init(&rtq)
	WatchHeapOps()
	exists := make(map[string]bool)
	return &ReceiptQueue{&rtq, exists, int64(lockTimeInSeconds * 1000)}
}

// - Push onto the queue and then resort (latest to earliest) also add to fast Exists map for quick checks
func (rtq *ReceiptQueue) Push(receipt *types.Receipt) {
	itm := Item{receipt, receipt.Created.UnixNano(), rtq.Queue.Len()+1}
	HeapPush(rtq.Queue, &itm)
	if(rtq.Queue.Len() > 0) {
		sort.Sort(rtq.Queue)
	}
	rtq.ExistsMap[receipt.TransactionHash] = true
}

// - Push onto the queue and then resort (latest to earliest) also add to fast Exists map for quick checks
func (rtq *ReceiptQueue) Pop() *types.Receipt {
	itm := HeapPop(rtq.Queue).(*Item)
	receipt := itm.Data.(*types.Receipt)
	delete(rtq.ExistsMap, receipt.TransactionHash)
	return receipt
}

// - Check to see if there is an item in the queue that is older than the LockTime
func (rtq ReceiptQueue) HasAvailable() bool {
	timestamp := rtq.Queue.Peek()
	if timestamp != -1 && utils.ToMilliSeconds(time.Now()) - timestamp > rtq.LockTime {
		return true
	}
	return false
}

// - Check to see if the current transaction Hash is in the exists map
func (rtq ReceiptQueue) Exists(key string) bool {
	return rtq.ExistsMap[key] == true
}

// - Dump returns the contents of the queue from oldest to newest (the order they are constantly sorted to)
func (rtq ReceiptQueue) Dump() []*types.Receipt {

	receiptList := make([]*types.Receipt, 0)
	for _, itm := range *rtq.Queue {
		receiptList = append(receiptList, itm.Data.(*types.Receipt))
	}
	return receiptList
}