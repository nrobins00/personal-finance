package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nrobins00/personal-finance/platform/authenticator"
	"github.com/nrobins00/personal-finance/platform/database"
	"github.com/nrobins00/personal-finance/platform/plaidActions"
	"github.com/nrobins00/personal-finance/platform/router"
)

func main() {
	var prod bool
	flag.BoolVar(&prod, "prod", false, "whether to use production settings")

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	clientId := os.Getenv("PLAID_CLIENT_ID")

	var secretEnvKey string
	if prod {
		secretEnvKey = "PLAID_SECRET_PROD"
	} else {
		secretEnvKey = "PLAID_SECRET_SANDBOX"
	}
	secret := os.Getenv(secretEnvKey)

	webhookUrl := os.Getenv("WEBHOOK_URL")

	fmt.Printf("clientId: %s\n", clientId)
	fmt.Printf("secret: %s\n", secret)
	fmt.Printf("webhookUrl: %s\n", webhookUrl)
	plaidClient := plaidActions.PlaidClient{ClientId: clientId, Secret: secret, WebhookUrl: webhookUrl}
	plaidClient.InitClient(prod)

	db := database.CreateDatabase("test.db")

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	rtr := router.New(auth, db, plaidClient)

	log.Print("Server listening on http://localhost:3000/")
	if err := http.ListenAndServe("0.0.0.0:3000", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
