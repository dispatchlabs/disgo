package queue

import (
	"container/heap"
	"sort"
	"github.com/dispatchlabs/disgo/commons/utils"
)

type PriorityQueue []*Item

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
	utils.Debug("Swap() ", i, j)
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

func (pq *PriorityQueue) DumpHighToLow() []*Item {
	sort.Sort(pq)
	return *pq
}

func (pq *PriorityQueue) DumpLowToHigh() []*Item {
	sort.Sort(sort.Reverse(pq))
	return *pq
}

// - Get the top Priority to support making decision on priority from calling code
func (pq PriorityQueue) Peek() int64 {
	length := pq.Len()
	utils.Debug("Peek --> %d", length)
	if length > 0 {
		return pq[length-1].Priority
	}
	return -1
}