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
	"time"
	"github.com/dispatchlabs/disgo/commons/utils"
	"sort"
)

type GossipQueue struct {
	Queue 		*PriorityQueue
	ExistsMap 	map[string]bool
	LockTime    int64
}

// need to figure out best way to implement the lock time
func NewGossipQueue(lockTimeInSeconds int) *GossipQueue {
	gq := make(PriorityQueue, 0)
	heap.Init(&gq)
	WatchHeapOps()
	exists := make(map[string]bool)
	return &GossipQueue{&gq, exists, int64(lockTimeInSeconds * 1000)}
}

// - Push onto the queue and then resort (latest to earliest) also add to fast Exists map for quick checks
func (gq *GossipQueue) Push(gossip *types.Gossip) {
	itm := Item{gossip, gossip.Transaction.Time, gq.Queue.Len()+1}
	HeapPush(gq.Queue, &itm)
	if(gq.Queue.Len() > 0) {
		sort.Sort(gq.Queue)
	}
	gq.ExistsMap[gossip.Transaction.Hash] = true
}

// - Push onto the queue and then resort (latest to earliest) also add to fast Exists map for quick checks
func (gq *GossipQueue) Pop() *types.Gossip {
	itm := HeapPop(gq.Queue).(*Item)
	gossip := itm.Data.(*types.Gossip)
	delete(gq.ExistsMap, gossip.Transaction.Hash)
	return gossip
}

// - Check to see if there is an item in the queue that is older than the LockTime
func (gq GossipQueue) HasAvailable() bool {
	timestamp := gq.Queue.Peek()
	if timestamp != -1 && utils.ToMilliSeconds(time.Now()) - timestamp > gq.LockTime {
		return true
	}
	return false
}

// - Check to see if the current Gossip Hash is in the exists map
func (gq GossipQueue) Exists(key string) bool {
	return gq.ExistsMap[key] == true
}

// - Dump returns the contents of the queue from oldest to newest (the order they are constantly sorted to)
func (gq GossipQueue) Dump() []*types.Gossip {

	gossipList := make([]*types.Gossip, 0)
	for _, itm := range *gq.Queue {
		gossipList = append(gossipList, itm.Data.(*types.Gossip))
	}
	return gossipList
}