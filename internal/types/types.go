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
	Type             string
	LastUpdatedDttm  time.Time
}

type Item struct {
	ItemKey     int
	ItemId      string
	AccessToken string
	Cursor      string
}

type Transaction struct {
	TransactionId    string
	AccountId        string
	AccountName      string
	Amount           float32
	CategoryPrimary  string
	CategoryDetailed string
	AuthorizedDttm   string
}

type Budget struct {
	BudgetKey int
	UserId    string
	Amount    float32
}
