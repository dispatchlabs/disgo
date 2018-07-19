package utils

import (
	"container/heap"
	"time"
)

// heapPopChanMsg - the message structure for a pop chan
type heapPopChanMsg struct {
	h      heap.Interface
	result chan interface{}
}

// heapPushChanMsg - the message structure for a push chan
type heapPushChanMsg struct {
	h heap.Interface
	x interface{}
}

var (
	quitChan chan bool
	// heapPushChan - push channel for pushing to a heap
	heapPushChan = make(chan heapPushChanMsg)
	// heapPopChan - pop channel for popping from a heap
	heapPopChan = make(chan heapPopChanMsg)
)

// HeapPush - safely push item to a heap interface
func HeapPush(h heap.Interface, x interface{}) {
	time.Sleep(5 * time.Second)
	heapPushChan <- heapPushChanMsg{
		h: h,
		x: x,
	}
}

// HeapPop - safely pop item from a heap interface
func HeapPop(h heap.Interface) interface{} {
	var result = make(chan interface{})
	heapPopChan <- heapPopChanMsg{
		h:      h,
		result: result,
	}
	return <-result
}

//stopWatchHeapOps - stop watching for heap operations
func stopWatchHeapOps() {
	quitChan <- true
}

// watchHeapOps - watch for push/pops to our heap, and serializing the operations
// with channels
func watchHeapOps() chan bool {
	var quit = make(chan bool)
	go func() {
		for {
			select {
			case <-quit:
				// TODO: update to quit gracefully
				// TODO: maybe need to dump state somewhere?
				return
			case popMsg := <-heapPopChan:
				popMsg.result <- heap.Pop(popMsg.h)
			case pushMsg := <-heapPushChan:
				heap.Push(pushMsg.h, pushMsg.x)
			}
		}
	}()
	return quit
}
