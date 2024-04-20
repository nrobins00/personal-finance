package types

import (
	"time"

	"github.com/plaid/plaid-go/plaid"
)

type User struct {
	userId   int
	userName string
	//password?
}

type Account struct {
	Base    plaid.AccountBase
	ItemKey int
}

type Item struct {
	ItemKey   int
	ItemId    string
	AccessKey string
	Cursor    string
}

type BankTransaction struct {
	TransactionKey int
	TransactionId  int
	AccountKey     int
	Amount         float32
	CategoryId     int
	AuthorizedDttm time.Time
}
