export type { User, Account, Item, Transaction };

type User = {
	userId: number
	userName: string
	//password?
}

type Account = {
	AccountId: string
	AvailableBalance: number
	CurrentBalance: number
	Mask: string
	Name: string
	LastUpdatedDttm: Date
}

type Item = {
	ItemKey: number
	ItemId: string
	AccessToken: string
	Cursor: string
}

type Transaction = {
	TransactionId: string
	AccountId: string
	AccountName: string
	Amount: number
	CategoryPrimary: string
	CategoryDetailed: string
	AuthorizedDttm: Date
}

//type Budget = {
//	BudgetKey int
//	UserId    string
//	Amount    float32
//}
