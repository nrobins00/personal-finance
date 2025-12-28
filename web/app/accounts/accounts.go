package accounts

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/nrobins00/personal-finance/platform/database"
	"github.com/nrobins00/personal-finance/types"
	"github.com/nrobins00/personal-finance/web/templates"
)

func AccountsHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		context := r.Context()
		session := context.Value("session").(*sessions.Session)

		// if session.Values["userId"] == nil {
		// 	//TODO: redirect to new user or sign in page
		// 	w.WriteHeader(404)
		// 	return
		// }

		userId := session.Values["userId"].(int64)

		items, err := db.GetAllItemsForUser(userId)
		if err != nil {
			log.Fatal("error geting items: ", err)
		}
		allAccounts := make([]types.Account, 0)
		for _, key := range items {
			//try to get accounts from db
			accounts, err := db.GetAllAccounts(key.ItemKey)
			if err != nil {
				msg := fmt.Sprintf("error getting accounts for item: %v", key.ItemId)
				log.Fatal(msg, err)
			}
			allAccounts = append(allAccounts, accounts...)
		}

		params := templates.AccountsParams{
			Accounts: allAccounts,
		}
		err = templates.Accounts(w, params)

		if err != nil {
			panic(err)
		}
	}
}
