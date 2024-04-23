package types

import (
	"time"
)

type User struct {
	userId   int
	userName string
	//password?
}

type Account struct {
	AccountId        string
	AvailableBalance float32
	CurrentBalance   float32
	Mask             string
	Name             string
	ItemKey          int
	LastUpdatedDttm  time.Time
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
