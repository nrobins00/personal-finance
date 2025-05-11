package home

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"

	"github.com/gorilla/sessions"
	"github.com/nrobins00/personal-finance/platform/database"
	"github.com/nrobins00/personal-finance/platform/plaidActions"
	"github.com/nrobins00/personal-finance/types"
	"github.com/nrobins00/personal-finance/util/customSort"
	"github.com/nrobins00/personal-finance/web/templates"
)

func HomePage(db *database.DB, plaidClient plaidActions.PlaidClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		context := r.Context()
		session := context.Value("session").(*sessions.Session)
		fmt.Println("hit homepage")

		if session.Values["userId"] == nil {
			//TODO: redirect to new user or sign in page
			w.WriteHeader(404)
			return
		}

		userId := session.Values["userId"].(int64)

		query := r.URL.Query()

		allTransactionsStr := query.Get("allTransactions")
		allTransactions := false
		if allTransactionsStr == "true" {
			allTransactions = true
		}

		sortColStr := query.Get("sortCol")
		var sortCol customSort.By
		switch sortColStr {
		case "Amount":
			sortCol = customSort.Amount
		case "Date":
			sortCol = customSort.Date
		case "Category":
			sortCol = customSort.Category
		case "AccountName":
			sortCol = customSort.AccountName
		default:
			sortCol = nil
		}

		sortDirStr := query.Get("sortDir")

		transactionsLimit := 0
		if !allTransactions {
			transactionsLimit = 10
		}
		transactions, err := getTransactions(db, userId, transactionsLimit)
		if err != nil {
			// update link
			//panic(err)
			fmt.Println(err)
		}

		if sortCol != nil {
			transSorter := &customSort.TransactionSorter{
				Transactions: transactions,
				By:           sortCol,
			}
			if sortDirStr == "desc" {
				sort.Sort(sort.Reverse(transSorter))
			} else {
				sort.Sort(transSorter)
			}
		}

		spendings, err := db.GetSpendingsForLastMonth(int(userId))
		if err != nil {
			//panic(err)
			fmt.Println(err)
		}

		budget, err := db.GetBudget(int(userId))
		if err != nil && err != sql.ErrNoRows {
			fmt.Println(err)
		}

		w.WriteHeader(200)
		homeParams := templates.HomeParams{
			Spent:            spendings,
			Budget:           budget,
			Transactions:     transactions,
			UserId:           userId,
			Page:             "home",
			MoreTransactions: !allTransactions,
			Categories:       getAllCategoriesFromTransactions(transactions),
			SortCol:          sortColStr,
			SortDir:          sortDirStr,
		}
		err = templates.Home(w, homeParams)
		if err != nil {
			panic(err)
		}
	}
}

func getTransactions(db *database.DB, userId int64, limit int) ([]types.Transaction, error) {
	// items, err := db.GetAllItemsForUser(userId)
	// if err != nil {
	// 	return nil, nil, err
	// }
	// plaidClient.GetTransactions(items[0].AccessToken, "")
	transactions, err := db.GetTransactionsForUser(int(userId), limit, 0)
	if err != nil {
		panic(err)
	}

	return transactions, nil
}

func getAllCategoriesFromTransactions(transactions []types.Transaction) []string {
	catSet := make(map[string]bool)
	for _, trans := range transactions {
		catSet[trans.CategoryDetailed] = true
	}

	catSlice := make([]string, 0, len(catSet))
	for cat := range catSet {
		catSlice = append(catSlice, cat)
	}
	return catSlice
}
