package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"strconv"

	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nrobins00/personal-finance/platform/authenticator"
	"github.com/nrobins00/personal-finance/platform/database"
	"github.com/nrobins00/personal-finance/platform/plaidActions"
	"github.com/nrobins00/personal-finance/platform/router"
	"github.com/nrobins00/personal-finance/types"
)

func main() {
	// var wait time.Duration
	// flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "duration to wait for existing connections to close")

	// var dir string

	// flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	// flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	rtr := router.New(auth, db, plaidClient)

	log.Print("Server listening on http://localhost:3000/")
	if err := http.ListenAndServe("0.0.0.0:3000", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}

	// r := mux.NewRouter()
	// // This will serve files under http://localhost:8080/static/<filename>
	// r.PathPrefix("/static/").Handler(http.FileServer(http.Dir(dir)))

	// r.HandleFunc("/", signin).Methods(http.MethodPost, http.MethodOptions, http.MethodGet)
	// r.HandleFunc("/signin", signin).Methods(http.MethodPost, http.MethodOptions)
	// r.HandleFunc("/api/linktoken", createLinkToken).Methods(http.MethodPost, http.MethodOptions)
	// r.HandleFunc("/api/budget", getBudget).Methods(http.MethodGet, http.MethodOptions)
	// r.HandleFunc("/api/budget/set", setBudget).Methods(http.MethodPost, http.MethodOptions)
	// r.HandleFunc("/api/spendings", getSpendings).Methods(http.MethodGet, http.MethodOptions)

	// //a := r.PathPrefix("/{id:[0-9]+}/").Subrouter()

	// // r.handleFunc("/api/webhook/
	// //a.Use(checkUserExists)

	// r.HandleFunc("/home/{id:[0-9]+}", homePage)
	// r.HandleFunc("/accounts/{id:[0-9]+}", accounts)
	// r.HandleFunc("/link/{id:[0-9]+}", linkBank)
	// r.HandleFunc("/new", newUser).Methods("GET")
	// r.HandleFunc("/new", newUserPost).Methods("POST")
	// r.HandleFunc("/api/publicToken/{id:[0-9]+}", exchangePublicToken).Methods(http.MethodPost, http.MethodOptions)
	// r.HandleFunc("/budget/{id:[0-9]+}", budget).Methods("GET")
	// r.HandleFunc("/budget/{id:[0-9]+}", updateBudget).Methods("POST")

	// srv := &http.Server{
	// 	Addr:         "0.0.0.0:8080",
	// 	WriteTimeout: 0, //time.Second * 15,
	// 	ReadTimeout:  0, //time.Second * 15,
	// 	IdleTimeout:  0, //time.Second * 60,
	// 	Handler:      r,
	// }
	// go func() {
	// 	if err := srv.ListenAndServe(); err != nil {
	// 		log.Println(err)
	// 	}
	// }()
	// log.Println("listening at ", srv.Addr)

	// c := make(chan os.Signal, 1)

	// signal.Notify(c, os.Interrupt)

	// <-c

	// ctx, cancel := context.WithTimeout(context.Background(), wait)
	// defer cancel()

	// srv.Shutdown(ctx)

	// log.Println("shutting down")
	// os.Exit(0)
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

func checkUserExists(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userIdStr := vars["id"]
		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			w.WriteHeader(404)
			return
		}
		if !db.CheckUserExists(userId) {
			fmt.Println("redirecting")
			http.Redirect(w, r, "/new", http.StatusTemporaryRedirect)
			//http.ServeFile(w, r, "static/signup.html")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	log.Println("hit this")
	http.ServeFile(w, r, "static/signup.html")
}

func newUserPost(w http.ResponseWriter, r *http.Request) {
	// r.ParseForm()
	// r.Form.Get("nonexistent")
	// username := r.Form.Get("username")
	// pass := r.Form.Get("pass")
	// //TODO: validate username and pass
	// //TODO: hash password?
	// userId, err := db.CreateUser(email)

	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// }

	// idCookie := fmt.Sprintf("userId=%d", userId)
	// w.Header().Set("Set-Cookie", idCookie)
	// w.WriteHeader(http.StatusOK)
	// newUrl := fmt.Sprintf("/%v/home", userId)
	// http.Redirect(w, r, newUrl, http.StatusTemporaryRedirect)
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

func budget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	budget, _ := db.GetBudget(int(userId))

	templates, err := template.ParseFiles("templates/budget.tmpl", "templates/navbar.tmpl")
	if err != nil {
		panic(err)
	}

	err = templates.ExecuteTemplate(w, "Budget", struct {
		Page          string
		UserId        int64
		CurrentBudget float32
	}{
		Page:          "budget",
		UserId:        userId,
		CurrentBudget: budget,
	})

	if err != nil {
		panic(err)
	}
}

func updateBudget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	r.ParseForm()
	amountStr := r.Form.Get("amount")
	amount, err := strconv.ParseFloat(amountStr, 32)
	if err != nil {
		panic(err)
	}

	db.InsertBudget(int(userId), float32(amount))
	//w.Write([]byte(fmt.Sprintf("budget updated to %v", amount)))
}

func signin(w http.ResponseWriter, r *http.Request) {
	//if r.Method == http.MethodOptions {
	//	return
	//}
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
