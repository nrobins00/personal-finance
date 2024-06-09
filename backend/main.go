package main

import (
	"context"
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

	r.HandleFunc("/signin", signin).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/", templTest).Methods(http.MethodGet)
	r.HandleFunc("/api/linktoken", createLinkToken).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/publicToken", exchangePublicToken).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/transactions", getTransactions).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/accounts", getAllAccounts).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/budget", getBudget).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/budget/set", setBudget).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/spendings", getSpendings).Methods(http.MethodGet, http.MethodOptions)

	r.HandleFunc("/home/{id:[0-9]+}", homePage)
	r.HandleFunc("/accounts/{id:[0-9]+}", accounts)

	//r.Use(mux.CORSMethodMiddleware(r))
	//r.Use(CorsMiddleware)

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

func homePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	//fmt.Fprintf(w, "UserId: %v\n", vars["userId"])
	spendings, err := db.GetSpendingsForLastMonth(int(userId))
	if err != nil {
		panic(err)
	}
	budget, err := db.GetBudget(int(userId))
	if err != nil {
		panic(err)
	}
	templates := template.Must(template.ParseFiles("templates/navbar.tmpl"))
	_, err = templates.ParseFiles("templates/home.tmpl")
	if err != nil {
		panic(err)
	}
	//w.WriteHeader(200)
	templates.ExecuteTemplate(w, "Home", struct {
		Spent  float32
		Budget float32
	}{
		Spent:  spendings,
		Budget: budget,
	})
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

	tmpl, err := template.ParseFiles("templates/accounts.tmpl")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)
	tmpl.Execute(w, struct {
		UserId   int
		Accounts []types.Account
	}{
		UserId:   int(userId),
		Accounts: allAccounts,
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
	idCookie := fmt.Sprintf("userId=%d", userId)
	w.Header().Set("Set-Cookie", idCookie)
	w.WriteHeader(http.StatusOK)
}

func templTest(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/link.html")
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
	userIdCookie, err := r.Cookie("userId")
	if err != nil || userIdCookie == nil {
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

func getAllAccounts(w http.ResponseWriter, r *http.Request) {
	userIdCookie, err := r.Cookie("userId")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	userId, err := strconv.ParseInt(userIdCookie.Value, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
	w.WriteHeader(http.StatusOK)
	resp := map[string][]types.Account{"accounts": allAccounts}
	body, err := json.Marshal(resp)
	w.Write(body)
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
	userIdCookie, err := r.Cookie("userId")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userId, err := strconv.ParseInt(userIdCookie.Value, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println(userId)
	items, err := db.GetAllItemsForUser(userId)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(items)
	added := make([]types.Transaction, 0)
	modified := make([]types.Transaction, 0)
	removed := make([]types.Transaction, 0)

	//TODO: replace this with webhooks
	for _, item := range items {
		add, mod, rem, newcursor, err := plaidClient.GetTransactions(item.AccessToken, item.Cursor)
		if err != nil {
			log.Fatal("error getting transactions: ", err)
		}
		if len(add) > 0 || len(mod) > 0 || len(rem) > 0 {
			err = db.UpdateTransactions(item.ItemId, add, mod, rem, newcursor)
		}
		if err != nil {
			log.Fatal(err)
		}

		added = append(added, add...)
		modified = append(modified, mod...)
		removed = append(removed, rem...)
		fmt.Println(newcursor)
	}

	transactions, err := db.GetTransactionsForUser(int(userId))
	if err != nil {
		log.Fatal(err)
	}

	resp := map[string][]types.Transaction{"transactions": transactions}
	json, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(json)
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
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
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
