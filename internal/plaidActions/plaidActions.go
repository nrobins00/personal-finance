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
	fmt.Println(accessKey)
	accountsGetRequest := plaid.NewAccountsGetRequest(accessKey)
	accountsGetResp, resp, err := c.client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*accountsGetRequest,
	).Execute()
	if err != nil {
		fmt.Println(*resp)
		fmt.Println(accountsGetResp)
		return []types.Account{}, err
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
			Type:             string(plaidAcc.GetType()),
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (c PlaidClient) GetTransactions(accessToken, cursor string) (
	add []types.Transaction,
	mod []types.Transaction,
	rem []types.Transaction,
	newCursor string,
	err error,
) {
	ctx := context.Background()
	oldCursor := cursor
	var added []plaid.Transaction
	var modified []plaid.Transaction
	var removed []plaid.RemovedTransaction
	hasMore := true

	for hasMore {
		request := plaid.NewTransactionsSyncRequest(accessToken)
		if cursor != "" {
			request.SetCursor(cursor)
		}
		resp, _, err := c.client.PlaidApi.TransactionsSync(
			ctx,
		).TransactionsSyncRequest(*request).Execute()
		if err != nil {
			log.Fatal(err)
			return add, mod, rem, oldCursor, err
		}

		added = append(added, resp.GetAdded()...)
		modified = append(modified, resp.GetModified()...)
		removed = append(removed, resp.GetRemoved()...)

		hasMore = resp.GetHasMore()

		cursor = resp.GetNextCursor()
	}

	add = make([]types.Transaction, 0)
	for _, tr := range added {
		add = append(add, types.Transaction{
			TransactionId:    tr.GetTransactionId(),
			Amount:           tr.GetAmount(),
			CategoryPrimary:  tr.GetPersonalFinanceCategory().Primary,
			CategoryDetailed: tr.GetPersonalFinanceCategory().Detailed,
			AccountId:        tr.GetAccountId(),
			AuthorizedDttm:   tr.GetAuthorizedDate(),
		})
	}

	mod = make([]types.Transaction, 0)
	for _, tr := range modified {
		mod = append(mod, types.Transaction{
			TransactionId:    tr.GetTransactionId(),
			Amount:           tr.GetAmount(),
			CategoryPrimary:  tr.GetPersonalFinanceCategory().Primary,
			CategoryDetailed: tr.GetPersonalFinanceCategory().Detailed,
			AccountId:        tr.GetAccountId(),
		})
	}

	rem = make([]types.Transaction, 0)
	for _, tr := range removed {
		rem = append(rem, types.Transaction{
			TransactionId: tr.GetTransactionId(),
		})
	}

	return add, mod, rem, cursor, nil
}
