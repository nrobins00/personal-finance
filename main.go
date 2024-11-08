package main

import (
	"context"
	"database/sql"
	b64 "encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
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
	"github.com/nrobins00/personal-finance/internal/database"
	"github.com/nrobins00/personal-finance/internal/plaidActions"
	"github.com/nrobins00/personal-finance/internal/types"
)

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "duration to wait for existing connections to close")

	var dir string

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	r := mux.NewRouter()

	// This will serve files under http://localhost:8080/static/<filename>
	r.PathPrefix("/static/").Handler(http.FileServer(http.Dir(dir)))

	//r.Handle("/", http.FileServer(http.Dir("./static/")))
	r.HandleFunc("/", signin).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/signin", signin).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/linktoken", createLinkToken).Methods(http.MethodPost, http.MethodOptions)
	//r.HandleFunc("/api/transactions", getTransactions).Methods(http.MethodGet, http.MethodOptions)
	//r.HandleFunc("/api/accounts", getAllAccounts).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/budget", getBudget).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/budget/set", setBudget).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/spendings", getSpendings).Methods(http.MethodGet, http.MethodOptions)

	r.HandleFunc("/home/{id:[0-9]+}", homePage)
	r.HandleFunc("/accounts/{id:[0-9]+}", accounts)
	r.HandleFunc("/link/{id:[0-9]+}", linkBank)
	r.HandleFunc("/new", newUser).Methods("GET")
	r.HandleFunc("/new", newUserPost).Methods("POST")
	r.HandleFunc("/api/publicToken/{id:[0-9]+}", exchangePublicToken).Methods(http.MethodPost, http.MethodOptions)

	//r.Use(mux.CORSMethodMiddleware(r))
	//r.Use(CorsMiddleware)

	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 0, //time.Second * 15,
		ReadTimeout:  0, //time.Second * 15,
		IdleTimeout:  0, //time.Second * 60,
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
	plaidClient plaidActions.PlaidClient
	db          *database.DB
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	db = database.CreateDatabase("test.db")

	clientId := os.Getenv("PLAID_CLIENT_ID")
	secret := os.Getenv("PLAID_SECRET")

	fmt.Printf("clientId: %s\n", clientId)
	fmt.Printf("secret: %s\n", secret)
	plaidClient = plaidActions.PlaidClient{ClientId: clientId, Secret: secret}
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

func newUser(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/signup.html")
}

func newUserPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	r.Form.Get("nonexistent")
	username := r.Form.Get("username")
	pass := r.Form.Get("pass")
	//TODO: validate username and pass
	//TODO: hash password?
	userId, err := db.CreateUser(username, pass)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	idCookie := fmt.Sprintf("userId=%d", userId)
	w.Header().Set("Set-Cookie", idCookie)
	w.WriteHeader(http.StatusOK)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		//TODO: redirect to new user or sign in page
		w.WriteHeader(404)
		return
	}
	query := r.URL.Query()
	allTransactionsSlice := query["allTransactions"]
	allTransactions := false
	if len(allTransactionsSlice) > 0 {
		if allTransactionsSlice[0] == "true" {
			allTransactions = true
		}
	}
	//fmt.Fprintf(w, "UserId: %v\n", vars["userId"])
	accounts, err := getAllAccounts(userId)
	fmt.Println("Accounts: ", accounts)

	transactionsLimit := 0
	if !allTransactions {
		transactionsLimit = 10
	}
	transactions, err, updateTokens := getTransactions(userId, transactionsLimit)
	if len(updateTokens) > 0 {
		// send to route that builds and sends linkUpdate HTML
		updateItems(w, r, updateTokens)
		return
	}
	if err != nil {
		// update link
		//panic(err)
		fmt.Println(err)
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
	templates := template.Must(template.ParseFiles("templates/navbar.tmpl", "templates/transactions.tmpl"))
	_, err = templates.ParseFiles("templates/home.tmpl")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)

	err = templates.ExecuteTemplate(w, "Home", struct {
		Spent            float32
		Budget           float32
		Transactions     []types.Transaction
		UserId           int64
		Page             string
		MoreTransactions bool
	}{
		Spent:            spendings,
		Budget:           budget,
		Transactions:     transactions,
		UserId:           userId,
		Page:             "home",
		MoreTransactions: !allTransactions,
	})
	if err != nil {
		panic(err)
	}
}

func updateItems(w http.ResponseWriter, r *http.Request, updateTokens []string) {
	template, err := template.ParseFiles("templates/updateLink.tmpl")
	if err != nil {
		panic(err)
	}
	err = template.ExecuteTemplate(w, "UpdateLink", struct {
		UpdateTokens []string
	}{
		UpdateTokens: updateTokens,
	})

	if err != nil {
		panic(err)
	}
}

func accounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	items, err := db.GetAllItemsForUser(userId)
	if err != nil {
		log.Fatal("error geting items: ", err)
	}
	allAccounts := make([]types.Account, 0)
	for _, key := range items {
		//try to get accounts from db
		//if they don't exist (i.e. this is the first time), grab them from Plaid
		accounts, err := db.GetAllAccounts(key.ItemKey)
		if err != nil {
			msg := fmt.Sprintf("error getting accounts for item: %v", key.ItemId)
			log.Fatal(msg, err)
		}
		if len(accounts) == 0 {
			//get accounts from Plaid
			accounts, err = plaidClient.GetAllAccounts(key.AccessToken)
			if err != nil {
				msg := fmt.Sprintf("error getting accounts for item: %v", key.ItemId)
				log.Fatal(msg, err)
			}
			db.InsertAccounts(userId, key.ItemKey, accounts)
		}
		allAccounts = append(allAccounts, accounts...)
	}

	fmt.Println(len(allAccounts))
	templates := template.Must(template.ParseFiles("templates/navbar.tmpl"))
	_, err = templates.ParseFiles("templates/accounts.tmpl")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)
	err = templates.ExecuteTemplate(w, "Accounts", struct {
		UserId   int
		Accounts []types.Account
		Page     string
	}{
		UserId:   int(userId),
		Accounts: allAccounts,
		Page:     "accounts",
	})

	if err != nil {
		panic(err)
	}
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
	idCookie := fmt.Sprintf("userId=%d", userId)
	w.Header().Set("Set-Cookie", idCookie)
	w.WriteHeader(http.StatusOK)
}

func linkBank(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("static/link.html")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func createLinkToken(w http.ResponseWriter, r *http.Request) {
	linkToken := plaidClient.GetLinkToken()
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
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() //why??
	t := struct {
		Public_token string
	}{}
	err = d.Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	//TODO: make sure token exists
	fmt.Println(t)
	publicToken := t.Public_token
	fmt.Printf("publicToken: %s", publicToken)
	accessToken, itemId := plaidClient.ExchangePublicToken(publicToken)
	fmt.Printf("userId: %d\nitemId: %s", userId, itemId)

	newId, err := db.CreateItem(userId, itemId, accessToken)
	if err != nil {
		panic(err)
	}

	fmt.Print(newId)

	resp := map[string]string{"access_token": accessToken}
	json, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(json)
}

func getAllAccounts(userId int64) ([]types.Account, error) {
	items, err := db.GetAllItemsForUser(userId)
	if err != nil {
		log.Fatal("error geting items: ", err)
	}

	allAccounts := make([]types.Account, 0)
	for _, key := range items {
		//try to get accounts from db
		//if they don't exist (i.e. this is the first time), grab them from Plaid
		//TODO: you can link accounts after the initial one, so we need to always check plaid?
		accounts, err := db.GetAllAccounts(key.ItemKey)
		if err != nil {
			msg := fmt.Sprintf("error getting accounts for item: %v", key.ItemId)
			log.Fatal(msg, err)
		}
		if len(accounts) == 0 {
			//get accounts from Plaid
			accounts, err = plaidClient.GetAllAccounts(key.AccessToken)
			if err != nil {
				msg := fmt.Sprintf("error getting accounts for item: %v", key.ItemId)
				log.Fatal(msg, err)
				return nil, err
			}
			db.InsertAccounts(userId, key.ItemKey, accounts)
		}
		allAccounts = append(allAccounts, accounts...)
	}
	return allAccounts, nil
}

func getTransactions(userId int64, limit int) ([]types.Transaction, error, []string) {
	fmt.Println(userId)
	items, err := db.GetAllItemsForUser(userId)
	if err != nil {
		log.Fatal(err)
	}
	added := make([]types.Transaction, 0)
	modified := make([]types.Transaction, 0)
	removed := make([]types.Transaction, 0)

	updateTokens := make([]string, 0)

	//TODO: replace this with webhooks
	for _, item := range items {
		add, mod, rem, newcursor, err, updateToken := plaidClient.GetTransactions(item.AccessToken, item.Cursor)
		if updateToken != "" {
			updateTokens = append(updateTokens, updateToken)
		}
		if err != nil {
			// do
			return nil, err, nil
		}
		if len(add) > 0 || len(mod) > 0 || len(rem) > 0 {
			err = db.UpdateTransactions(item.ItemId, add, mod, rem, newcursor)
		} else {
			fmt.Println("no transactions returned.")
		}
		if err != nil {
			return nil, err, nil
		}

		added = append(added, add...)
		modified = append(modified, mod...)
		removed = append(removed, rem...)
		fmt.Println("transactions: ", added, modified, removed)
		fmt.Println("newcursor: ", newcursor)
	}

	if len(updateTokens) > 0 {
		return nil, nil, updateTokens
	}

	transactions, err := db.GetTransactionsForUser(int(userId), limit, 0)
	fmt.Print("transaction line")
	if err != nil {
		panic(err)
	}

	return transactions, err, nil
}

func getBalance(w http.ResponseWriter, r *http.Request) {

}

func getBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	userId, err := getUserId(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	budget, err := db.GetBudget(userId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp := map[string]float32{"budget": budget}
	json, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

type BudgetRequest struct {
	Budget float32
}

func setBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	userId, err := getUserId(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var b BudgetRequest

	err = json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("Budget: %v userId: %v", b.Budget, userId)
	err = db.InsertBudget(userId, b.Budget)
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusOK)
}

func getUserId(r *http.Request) (int, error) {
	userIdCookie, err := r.Cookie("userId")
	if err != nil {
		return -1, err
	}
	userId, err := strconv.ParseInt(userIdCookie.Value, 10, 0)
	if err != nil {
		return -1, err
	}
	return int(userId), nil
}

func getSpendings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	userId, err := getUserId(r)
	if err != nil {
		log.Fatal(err)
		//TODO: just pass w into getUserId and write a bad status header
	}

	spendings, err := db.GetSpendingsForLastMonth(userId)
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusOK)
	resp := map[string]float32{"spendings": spendings}
	json, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(json)
}
