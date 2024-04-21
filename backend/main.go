package main

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nrobins00/personal-finance/internal/database"
	"github.com/nrobins00/personal-finance/internal/plaidActions"
	"github.com/plaid/plaid-go/plaid"
)

func main() {

}

var (
	plaidClient plaidActions.PlaidClient
	db          *database.DB
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	db = database.CreateDatabase("test.db")

	clientId = os.Getenv("PLAID_CLIENT_ID")
	secret = os.Getenv("PLAID_SECRET")

	fmt.Printf("clientId: %s\n", clientId)
	fmt.Printf("secret: %s\n", secret)
	plaidClient := PlaidClient{ClientId: clientId, Secret: secret}
	plaidClient.InitClient()
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, authorization, Cookie")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}

func signin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	auth := r.Header.Get("Authorization")
	usernameAndPass, err := b64.StdEncoding.DecodeString(auth)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userAndPass := strings.Split(string(usernameAndPass), ":")
	if len(userAndPass) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	username := userAndPass[0]
	pass := userAndPass[1]
	fmt.Printf("username: %s, pass: %s", username, pass)
	userId, err := db.GetUserId(username, pass)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	idCookie := fmt.Sprintf("userId=%d; SameSite=None", userId)
	w.Header().Set("Set-Cookie", idCookie)
	w.WriteHeader(http.StatusOK)
}

func createLinkToken(w http.ResponseWriter, r *http.Request) {
	linkToken := plaidActions.GetLinkToken(client, clientId)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	resp := map[string]string{"link_token": linkToken}
	json, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

func exchangePublicToken(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	userIdCookie, err := r.Cookie("userId")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	userIdString := userIdCookie.Value
	userId, err := strconv.ParseInt(userIdString, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	//publicToken := c.Request.Body.Read()("public_token")
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() //why??
	t := struct {
		Public_token string
	}{}
	err = d.Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	fmt.Println(t)
	publicToken := t.Public_token
	fmt.Printf("publicToken: %s", publicToken)
	accessToken, itemId := plaidClient.ExchangePublicToken(publicToken)
	fmt.Printf("userId: %d\nitemId: %s", userId, itemId)

	db.CreateItem(userId, itemId, accessToken)

	resp := map[string]string{"access_token": accessToken}
	json, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(json)
}

type account struct {
	accountId                int
	name                     string
	availableBal, currentBal float32
}

type item struct {
	itemKey   int
	accessKey string
}

func getAllAccounts(w http.ResponseWriter, r *http.Request) {
	userIdCookie, err := r.Cookie("userId")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	userId, err := strconv.ParseInt(userIdCookie.Value, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	const itemQuery string = `
		SELECT itemKey, accessKey FROM item where userId = ?
	`

	const accQuery string = `
		SELECT accountId, name, availableBalance, currentBalance
		FROM account
		WHERE itemKey = ?
	`

	rows, err := db.Query(itemQuery, userId)
	if err != nil {
		log.Fatal(err)
	}
	items := make([]item, 0)
	defer rows.Close()
	for rows.Next() {
		var itemKey int
		var accessKey string
		if err := rows.Scan(&itemKey, &accessKey); err != nil {
			log.Fatal(err)
		}
		items = append(items, item{itemKey, accessKey})
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	accounts := make([]database.Account, 0)
	for _, key := range items {
		accountsGetRequest := plaid.NewAccountsGetRequest(key.accessKey)
		accountsGetResp, _, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
			*accountsGetRequest,
		).Execute()
		if err != nil {
			log.Fatal("accounts/get execute", err)
		}
		for _, acc := range accountsGetResp.GetAccounts() {
			account := database.Account{Base: acc, ItemKey: key.itemKey}
			accounts = append(accounts, account)
		}
	}

	db.InsertAccounts(userId, accounts)

	w.WriteHeader(http.StatusOK)
	resp := map[string][]database.Account{"accounts": accounts}
	body, err := json.Marshal(resp)
	w.Write(body)
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
	userIdCookie, err := r.Cookie("userId")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	userId, err := strconv.ParseInt(userIdCookie.Value, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	start := time.Now()
	const tokenQuery string = `
        SELECT itemId, accessKey FROM item where userId = ?
    `
	var accessToken string
	err = db.QueryRow(tokenQuery, userId).Scan(&accessToken)
	if err != nil {

		w.WriteHeader(http.StatusNotFound)
		//TODO: probably shouldn't crash here
		log.Fatal(err)
	}
	fmt.Println(accessToken)
	duration := time.Since(start)
	fmt.Println(duration)
	ctx := context.Background()
	var cursor string
	var added []plaid.Transaction
	var modified []plaid.Transaction
	var removed []plaid.RemovedTransaction
	//var accounts []plaid.AccountBase
	hasMore := true
	//options := plaid.TransactionsSyncRequestOptions{
	//    IncludePersonalFinanceCategory := true,
	//}

	for hasMore && len(added) < 10 {
		request := plaid.NewTransactionsSyncRequest(accessToken)
		//request.SetOptions(options)
		if cursor != "" {
			request.SetCursor(cursor)
		}
		resp, httpResp, err := client.PlaidApi.TransactionsSync(
			ctx,
		).TransactionsSyncRequest(*request).Execute()
		if err != nil {
			log.Fatal(httpResp.Body)
		}

		added = append(added, resp.GetAdded()...)
		modified = append(modified, resp.GetModified()...)
		removed = append(removed, resp.GetRemoved()...)

		hasMore = resp.GetHasMore()

		cursor = resp.GetNextCursor()
	}

	//enc := json.NewEncoder(c.Writer)
	//db.UpdateTransactions(item)
	fmt.Println(added, modified, removed)
	resp := map[string][]plaid.Transaction{"added": added}
	json, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(json)
}

func getBalance(w http.ResponseWriter, r *http.Request) {

}
