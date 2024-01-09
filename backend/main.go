package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/plaid/plaid-go/v17/plaid"
)

func main() {

	r := gin.Default()
	r.POST("/api/linktoken", createLinkToken)
	r.POST("/api/publicToken", exchangePublicToken)
	r.OPTIONS("/api/publicToken", allowCors)
	r.GET("/api/transactions", getTransactions)

	r.Run()
}

var (
	client      *plaid.APIClient
	clientId    string
	secret      string
	accessToken string
	db          *sql.DB
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	db = createDatabase()

	clientId = os.Getenv("PLAID_CLIENT_ID")
	secret = os.Getenv("PLAID_SECRET")

	fmt.Printf("clientId: %s", clientId)
	fmt.Printf("secret: %s", secret)
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", clientId)
	configuration.AddDefaultHeader("PLAID-SECRET", secret)
	configuration.UseEnvironment(plaid.Sandbox)
	client = plaid.NewAPIClient(configuration)
}

func allowCors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("Access-Control-Allow-Methods", "POST")
}

func createLinkToken(c *gin.Context) {
	ctx := context.Background()
	user := plaid.LinkTokenCreateRequestUser{ClientUserId: clientId}
	request := plaid.NewLinkTokenCreateRequest(
		"Plaid Test", "en", []plaid.CountryCode{plaid.COUNTRYCODE_US},
		user,
	)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH})
	request.SetLinkCustomizationName("default")
	resp, httpResp, err := client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		fmt.Println(httpResp.Body)
		log.Fatal(err)
	}
	linkToken := resp.GetLinkToken()
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{
		"link_token": linkToken})
}

func exchangePublicToken(c *gin.Context) {
	ctx := context.Background()
	//publicToken := c.Request.Body.Read()("public_token")
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	t := struct {
		Public_token string
	}{}
	err := d.Decode(&t)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "public_token was empty",
		})
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
    createItem( 
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func getTransactions(c *gin.Context) {
	start := time.Now()
	const tokenQuery string = `
        SELECT accessKey FROM item where userId = ?
    `
	var accessToken string
	err := db.QueryRow(tokenQuery, 1).Scan(&accessToken)
	if err != nil {
        //TODO: probably shouldn't crash here
		log.Fatal(err)
	}
	duration := time.Since(start)
	fmt.Println(duration)
	ctx := context.Background()
	var cursor string
	var added []plaid.Transaction
	var modified []plaid.Transaction
	var removed []plaid.RemovedTransaction
	hasMore := true
	//options := plaid.TransactionsSyncRequestOptions{
	//    IncludePersonalFinanceCategory := true,
	//}

	for hasMore {
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

	c.Header("Access-Control-Allow-Origin", "*")
	//enc := json.NewEncoder(c.Writer)
	fmt.Println(added, modified, removed)
	c.JSON(http.StatusOK, gin.H{
		"added": added,
	})
}

