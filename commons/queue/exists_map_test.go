package queue

import (
	"testing"
	"fmt"
	"time"
)

func TestExistsMapConcurrency(t *testing.T) {

	em := NewExistsMap()
	go addToQueue(em)
	go existInQueue(em)
	go deleteFromQueue(em)
	time.Sleep(time.Second * 10)
}

func addToQueue(em *ExistsMap) {
	for i := 0; i < 1000; i++ {
		value := fmt.Sprintf("value-%d", i)
		em.Put(value)
		//fmt.Printf("Put: %s\n", value)
	}
}

func existInQueue(em *ExistsMap) {
	for i := 0; i < 1000; i++ {
		value := fmt.Sprintf("value-%d", i)
		em.Exists(fmt.Sprintf(value))
		//fmt.Printf("Exists: %s = %v\n", value, exists)
	}
}

func deleteFromQueue(em *ExistsMap) {
	for i := 0; i < 1000; i++ {
		value := fmt.Sprintf("value-%d", i)
		em.Delete(value)
		//fmt.Printf("Delete: %s\n", value)
	}
}