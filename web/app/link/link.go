package link

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/nrobins00/personal-finance/platform/database"
	"github.com/nrobins00/personal-finance/platform/plaidActions"
)

func ExchangePublicToken(db *database.DB, plaidClient plaidActions.PlaidClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			return
		}
		session := r.Context().Value("session").(*sessions.Session)
		if session.Values["userId"] == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userId := session.Values["userId"].(int64)

		d := json.NewDecoder(r.Body)
		d.DisallowUnknownFields() //why??
		t := struct {
			Public_token string
		}{}
		err := d.Decode(&t)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		//TODO: make sure token exists
		publicToken := t.Public_token
		fmt.Printf("publicToken: %s", publicToken)
		accessToken, itemId := plaidClient.ExchangePublicToken(publicToken)
		fmt.Printf("userId: %d\nitemId: %s", userId, itemId)

		newId, err := db.CreateItem(userId, itemId, accessToken)
		if err != nil {
			panic(err)
		}

		fmt.Print(newId)
		plaidClient.InitializeTransactions(accessToken)

		//TODO: definitely remove this
		resp := map[string]string{"access_token": accessToken}
		json, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write(json)
	}
}

func CreateLinkToken(plaidClient plaidActions.PlaidClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		linkToken := plaidClient.GetLinkToken()
		//w.Header().Set("Access-Control-Allow-Origin", "*")
		resp := map[string]string{"link_token": linkToken}
		json, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(json)
	}
}
