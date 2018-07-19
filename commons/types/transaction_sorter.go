package types

import "sort"

// By is the type of a "less" function that defines the ordering of its Transaction arguments.
type By func(t1, t2 *Transaction) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(transactions []*Transaction) {
	ts := &TransactionSorter{
		transactions: transactions,
		by:      by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ts)
}

// TransactionSorter joins a By function and a slice of Transaction to be sorted.
type TransactionSorter struct {
	transactions 	[]*Transaction
	by      		func(t1, t2 *Transaction) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (ts *TransactionSorter) Len() int {
	return len(ts.transactions)
}

// Swap is part of sort.Interface.
func (ts *TransactionSorter) Swap(i, j int) {
	ts.transactions[i], ts.transactions[j] = ts.transactions[j], ts.transactions[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (ts *TransactionSorter) Less(i, j int) bool {
	return ts.by(ts.transactions[i], ts.transactions[j])
}

func SortByTime(txs []*Transaction, ascending bool) {
	timestamp := func(t1, t2 *Transaction) bool {
		if(ascending) {
			return t1.Time < t2.Time
		} else {
			return t2.Time < t1.Time
		}
	}
	By(timestamp).Sort(txs)
}

func SortByTimeHash(txs []*Transaction, ascending bool) {
	timestamp := func(t1, t2 *Transaction) bool {
		if(ascending) {
			if t1.Time == t2.Time {
				return t1.Hash < t2.Hash
			} else {
				return t1.Time < t2.Time
			}
		} else {
			if t2.Time == t1.Time {
				return t2.Hash < t1.Hash
			} else {
				return t2.Time < t1.Time
			}
		}
	}
	By(timestamp).Sort(txs)
}
