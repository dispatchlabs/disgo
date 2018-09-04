package types

import (
	"github.com/dgraph-io/badger"
	"encoding/binary"
	"time"
	"fmt"
)

type Hertz struct {
	Amount 		int64
	Account 	string
	TxHash  	string
	Db      	*badger.DB
}


func (this *Hertz) Merge() error {
	key := []byte("merge")
	m := this.Db.GetMergeOperator(key, add, 200*time.Millisecond)
	defer m.Stop()

	m.Add(uint64ToBytes(1))
	m.Add(uint64ToBytes(2))
	m.Add(uint64ToBytes(3))

	res, err := m.Get() // res should have value 6 encoded
	if err != nil {
		return err
	}
	fmt.Println(bytesToUint64(res))
	return nil
}


func uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Merge function to add two uint64 numbers
func add(existing, new []byte) []byte {
	return uint64ToBytes(bytesToUint64(existing) + bytesToUint64(new))
}