package plaidActions

import (
	"context"
	"fmt"
	"log"

	"github.com/plaid/plaid-go/plaid"
)

type PlaidClient struct {
	ClientId string
	Secret   string
	client   *plaid.APIClient
}

func (client *PlaidClient) InitClient() {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("plaid-client-id", client.ClientId)
	configuration.AddDefaultHeader("plaid-secret", client.Secret)
	configuration.UseEnvironment(plaid.Sandbox)
	client.client = plaid.NewAPIClient(configuration)
}

func (c PlaidClient) GetLinkToken(client *plaid.APIClient, clientId string) string {
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
	return resp.GetLinkToken()
}

func (c PlaidClient) ExchangePublicToken(publicToken string) (string, string) {
	ctx := context.Background()
	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	exchangePublicTokenResp, httpResp, err := c.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*exchangePublicTokenReq,
	).Execute()
	if err != nil {
		log.Fatal(httpResp.Body)
	}
	accessToken := exchangePublicTokenResp.GetAccessToken()
	itemId := exchangePublicTokenResp.GetItemId()
	return accessToken, itemId
}
