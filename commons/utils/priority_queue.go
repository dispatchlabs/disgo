package utils

import (
	"container/heap"
	"sync"
)

var singleton sync.Once
var instance *PriorityQueue

type PriorityQueue []*Item

func GetQueue() *PriorityQueue {
	singleton.Do(func() {
		initial := make(PriorityQueue, 0)
		instance = &initial
		heap.Init(instance)
	})
	return instance
}

// An Item is something we manage in a priority queue.
type Item struct {
	Data        interface{}
	Priority 	int64    // The priority of the item in the queue.
	Index 		int 	// The index of the item in the heap.
	// The index is needed by update and is maintained by the heap.Interface methods.
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) Update(item *Item, value *interface{}, priority int64) {
	item.Data = value
	item.Priority = priority
	heap.Fix(pq, item.Index)
}

