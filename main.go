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

	// r.HandleFunc("/api/budget", getBudget).Methods(http.MethodGet, http.MethodOptions)
	// r.HandleFunc("/api/budget/set", setBudget).Methods(http.MethodPost, http.MethodOptions)

	// //a := r.PathPrefix("/{id:[0-9]+}/").Subrouter()

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
