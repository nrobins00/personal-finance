package webhooks

import (
	"encoding/json"
	"net/http"

	"github.com/nrobins00/personal-finance/platform/database"
	"github.com/nrobins00/personal-finance/platform/plaidActions"
)

type Webhook struct {
	WebhookType              string `json:"webhook_type"`
	WebhookCode              string `json:"webhook_code"`
	ItemId                   string `json:"item_id"`
	InitialUpdateComplete    bool   `json:"initial_update_complete"`
	HistoricalUpdateComplete bool   `json:"historical_update_complete"`
}

func WebhookHandler(db *database.DB, plaidClient plaidActions.PlaidClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var webhook Webhook
		err := json.NewDecoder(r.Body).Decode(&webhook)
		if err != nil {
			panic(err)
		}
		switch webhook.WebhookType {
		case "TRANSACTIONS":
			updateTransactions(webhook.ItemId, db, plaidClient)
		case "ACCOUNTS":
			updateAccounts(webhook.ItemId, db, plaidClient)
		}
	}
}

func updateAccounts(itemId string, db *database.DB, plaidClient plaidActions.PlaidClient) {
	item, err := db.GetItem(itemId)
	if err != nil {
		panic(err)
	}
	accounts, err := plaidClient.GetAllAccounts(item.AccessToken)
	if err != nil {
		panic(err)
	}

	err = db.InsertAccounts(item.UserId, item.ItemKey, accounts)
	if err != nil {
		panic(err)
	}
}

func updateTransactions(itemId string, db *database.DB, plaidClient plaidActions.PlaidClient) {
	// Plaid suggests that this should just put work onto a queue
	// because Plaid expects a 200 resp within 10 seconds, and
	// multiple requests could come in very quickly

	// get all accounts (it's easier to store transactions if we have the account already)
	// step 2:

	updateAccounts(itemId, db, plaidClient)

	item, err := db.GetItem(itemId)
	if err != nil {
		panic(err)
	}

	add, mod, rem, newcursor, updateToken, err := plaidClient.GetTransactions(item.AccessToken, item.Cursor)
	if updateToken != "" {
		// TODO: probably should store these to the DB, check that when the user tries to load,
		//       and diesplay updateLink.tmpl if they have any to update???
		return

	}

	if len(add) > 0 || len(mod) > 0 || len(rem) > 0 {
		err = db.UpdateTransactions(item.ItemId, add, mod, rem, newcursor)
	}
	if err != nil {
		panic(err)
	}
}

// TODO: probably should store these to the DB, check that when the user tries to load,
//       and diesplay updateLink.tmpl if they have any to update???
// func updateItems(w http.ResponseWriter, updateTokens []string) {
// 	template, err := template.ParseFiles("templates/updateLink.tmpl")
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = template.ExecuteTemplate(w, "UpdateLink", struct {
// 		UpdateTokens []string
// 	}{
// 		UpdateTokens: updateTokens,
// 	})

// 	if err != nil {
// 		panic(err)
// 	}
// }
