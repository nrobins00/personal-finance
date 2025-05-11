package customSort

import (
	"sort"
	"time"

	"github.com/nrobins00/personal-finance/types"
)

type By func(t1, t2 *types.Transaction) bool

func (by By) Sort(transactions []types.Transaction) {
	ts := &TransactionSorter{
		Transactions: transactions,
		By:           by,
	}
	sort.Sort(ts)
}

func (ts *TransactionSorter) Len() int {
	return len(ts.Transactions)
}

func (ts *TransactionSorter) Less(i, j int) bool {
	return ts.By(&ts.Transactions[i], &ts.Transactions[j])
}

func (ts *TransactionSorter) Swap(i, j int) {
	ts.Transactions[i], ts.Transactions[j] = ts.Transactions[j], ts.Transactions[i]
}

type TransactionSorter struct {
	Transactions []types.Transaction
	By           func(t1, t2 *types.Transaction) bool
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
