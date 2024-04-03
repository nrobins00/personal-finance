package tokens

import (
	"context"
	"fmt"
	"log"

	"github.com/plaid/plaid-go/plaid"
)

func GetLinkToken(client *plaid.APIClient, clientId string) string {
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
