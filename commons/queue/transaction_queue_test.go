package queue

import (
	"testing"
	"fmt"
	"sync"

)

func TestTxQueue(t *testing.T) {
	testTime := 3
	txq := NewTransactionQueue(testTime)

	for i := 1; i <= 5; i++ {
		tx := GetMockTransaction(int64(i))
		//tx.Time = int64(rand.Intn(100))
		txq.Push(tx)
	}
	fmt.Printf("Queue length = %d", txq.Queue.Len())

	var wg sync.WaitGroup
	wg.Add(1)
	//testStart := utils.ToMilliSeconds(time.Now())
	for {
		if(txq.HasAvailable()) {
			fmt.Printf("\n***** There is a record in the queue *****\n")
			tx := txq.Pop()
			fmt.Printf(tx.String())
		} else {
			//if utils.ToMilliSeconds(time.Now()) - testStart > int64(5000) {
				break
			//}
		}
	}
}
