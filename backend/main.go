package main

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nrobins00/personal-finance/database"
	"github.com/nrobins00/personal-finance/plaidActions"
	"github.com/plaid/plaid-go/plaid"
)

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "duration to wait for existing connections to close")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/signin", signin).Methods(http.MethodPost, http.MethodOptions)
	//r.HandleFu("/signin", allowCors)
	r.HandleFunc("/api/linktoken", createLinkToken).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/publicToken", exchangePublicToken).Methods(http.MethodPost, http.MethodOptions)
	//r.OPTIONS("/api/publicToken", allowCors)
	r.HandleFunc("/api/transactions", getTransactions).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/accounts", getAllAccounts).Methods(http.MethodGet, http.MethodOptions)
	//r.OPTIONS("/api/transactions", allowCors)

	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(CorsMiddleware)

	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	srv.Shutdown(ctx)

	log.Println("shutting down")
	os.Exit(0)
}

var (
	client      *plaid.APIClient
	clientId    string
	secret      string
	accessToken string
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
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientId)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(plaid.Sandbox)
	client = plaid.NewAPIClient(configuration)
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
	ctx := context.Background()
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
	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	exchangePublicTokenResp, httpResp, err := client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	if err != nil {
		log.Fatal(httpResp.Body)
	}
	accessToken = exchangePublicTokenResp.GetAccessToken()
	itemId := exchangePublicTokenResp.GetItemId()
	fmt.Printf("userId: %d\nitemId: %s", userId, itemId)

	db.CreateItem(userId, itemId, accessToken)

	resp := map[string]string{"access_token": accessToken}
	json, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(json)
}

type item struct {
	itemId, accessToken string
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
		SELECT itemId, accessKey FROM item where userId = ?
	`
	rows, err := db.Query(itemQuery, userId)
	if err != nil {
		log.Fatal(err)
	}
	items := make([]item, 0)
	defer rows.Close()
	for rows.Next() {
		var itemId, accessToken string
		if err := rows.Scan(&itemId, &accessToken); err != nil {
			log.Fatal(err)
		}
		items = append(items, item{itemId, accessToken})
	}

	rerr := rows.Close()
	if rerr != nil {
		log.Fatal(err)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusOK)
	resp := map[string][]item{"items": items}
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
