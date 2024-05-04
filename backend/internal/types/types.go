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
	LastUpdatedDttm  time.Time
}

type Item struct {
	ItemKey     int
	ItemId      string
	AccessToken string
	Cursor      string
}

type Transaction struct {
	TransactionId  string
	AccountId      string
	Amount         float32
	Category       string
	AuthorizedDttm time.Time
}

type Budget struct {
	BudgetKey int
	UserId    string
	Amount    float32
}
