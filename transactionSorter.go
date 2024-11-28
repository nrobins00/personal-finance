package main

import (
	"sort"
	"time"

	"github.com/nrobins00/personal-finance/internal/types"
)

type By func(t1, t2 *types.Transaction) bool

func (by By) Sort(transactions []types.Transaction) {
	ts := &TransactionSorter{
		transactions: transactions,
		by:           by,
	}
	sort.Sort(ts)
}

func (ts *TransactionSorter) Len() int {
	return len(ts.transactions)
}

func (ts *TransactionSorter) Less(i, j int) bool {
	return ts.by(&ts.transactions[i], &ts.transactions[j])
}

func (ts *TransactionSorter) Swap(i, j int) {
	ts.transactions[i], ts.transactions[j] = ts.transactions[j], ts.transactions[i]
}

type TransactionSorter struct {
	transactions []types.Transaction
	by           func(t1, t2 *types.Transaction) bool
}

func Amount(t1, t2 *types.Transaction) bool {
	return t1.Amount < t2.Amount
}

func Date(t1, t2 *types.Transaction) bool {
	t1Date, err := time.Parse(time.DateOnly, t1.AuthorizedDttm)
	if err != nil {
		panic(err)
	}
	t2Date, err := time.Parse(time.DateOnly, t2.AuthorizedDttm)
	if err != nil {
		panic(err)
	}
	return t1Date.Compare(t2Date) == -1
}

func AccountName(t1, t2 *types.Transaction) bool {
	return t1.AccountName < t2.AccountName
}

func Category(t1, t2 *types.Transaction) bool {
	return t1.CategoryDetailed < t2.CategoryDetailed
}
