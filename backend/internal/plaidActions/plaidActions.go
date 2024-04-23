package plaidActions

import (
	"context"
	"fmt"
	"log"

	"github.com/nrobins00/personal-finance/internal/types"
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

func (c *PlaidClient) GetLinkToken() string {
	ctx := context.Background()
	user := plaid.LinkTokenCreateRequestUser{ClientUserId: c.ClientId}
	request := plaid.NewLinkTokenCreateRequest(
		"Plaid Test", "en", []plaid.CountryCode{plaid.COUNTRYCODE_US},
		user,
	)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH})
	request.SetLinkCustomizationName("default")
	fmt.Println(c.client)
	resp, httpResp, err := c.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		fmt.Println(httpResp.Body)
		log.Fatal(err)
	}
	return resp.GetLinkToken()
}

func (c *PlaidClient) ExchangePublicToken(publicToken string) (string, string) {
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

func (c *PlaidClient) GetAllAccounts(accessKey string) ([]types.Account, error) {
	ctx := context.Background()
	accountsGetRequest := plaid.NewAccountsGetRequest(accessKey)
	accountsGetResp, _, err := c.client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*accountsGetRequest,
	).Execute()
	if err != nil {
		return []types.Account{}, nil
	}
	plaidAccounts := accountsGetResp.GetAccounts()
	accounts := make([]types.Account, 0)
	for _, plaidAcc := range plaidAccounts {
		acc := types.Account{
			AccountId:        plaidAcc.GetAccountId(),
			AvailableBalance: plaidAcc.Balances.GetAvailable(),
			CurrentBalance:   plaidAcc.Balances.GetCurrent(),
			Mask:             plaidAcc.GetMask(),
			Name:             plaidAcc.GetName(),
			ItemKey:          0,
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}
