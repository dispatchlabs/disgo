package queue

/*
 *  The Gossip Queue is a wrapper of the Priority queue heap implementation
 *  It uses the thread safe queue functions for pushing and popping from the queue
 *
 *  The intent is to make it very simple to use and eliminate casting
 *  It includes a few helper functions to keep thing easy
 */
import (
	"container/heap"

	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
)

type GossipQueue struct {
	Queue     *PriorityQueue
	ExistsMap *ExistsMap
}

// need to figure out best way to implement the lock time
func NewGossipQueue() *GossipQueue {
	gq := make(PriorityQueue, 0)
	heap.Init(&gq)
	WatchHeapOps()
	return &GossipQueue{&gq, NewExistsMap()}
}

// - Push onto the queue and then resort (latest to earliest) also add to fast Exists map for quick checks
func (gq *GossipQueue) Push(gossip *types.Gossip) {
	utils.Debug("GossipQueue.Push --> ", gossip.Transaction.Hash)
	itm := Item{gossip, gossip.Transaction.Time, gq.Queue.Len() + 1}
	gq.ExistsMap.Put(gossip.Transaction.Hash)
	HeapPush(gq.Queue, &itm)
}

// - Push onto the queue and then resort (latest to earliest) also add to fast Exists map for quick checks
func (gq *GossipQueue) Pop() *types.Gossip {
	itm := HeapPop(gq.Queue).(*Item)
	gossip := itm.Data.(*types.Gossip)
	utils.Debug("GossipQueue.Pop --> ", gossip.Transaction.Hash)
	gq.ExistsMap.Delete(gossip.Transaction.Hash)
	return gossip
}

// - Check to see if there is an item in the queue that is older than the LockTime
func (gq GossipQueue) HasAvailable() bool {
	timestamp := gq.Queue.Peek()
	//if timestamp != -1 && utils.ToMilliSeconds(time.Now()) - timestamp > gq.LockTime {
	if timestamp != -1 {
		return true
	}
	return false
}

// - Check to see if the current Gossip Hash is in the exists map
func (gq GossipQueue) Exists(key string) bool {
	return gq.ExistsMap.Exists(key)
}

// - Dump returns the contents of the queue from oldest to newest (the order they are constantly sorted to)
func (gq GossipQueue) Dump() []*types.Gossip {

	gossipList := make([]*types.Gossip, 0)
	for _, itm := range *gq.Queue {
		gossipList = append(gossipList, itm.Data.(*types.Gossip))
	}
	return gossipList
}
